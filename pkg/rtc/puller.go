package rtc

import (
	"fmt"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v4"
	"github.com/sirupsen/logrus"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/stat"
	"sync"
	"time"
)

type Puller struct {
	rtcService  Service
	req         *dto.PlayReq
	claims      baseDto.ThkClaims
	key         string
	subKey      string
	mutex       *sync.RWMutex
	clientSdp   *webrtc.SessionDescription
	serverSdp   *webrtc.SessionDescription
	peerConn    *webrtc.PeerConnection
	interceptor *stats.InterceptorFactory
	statsGetter *stats.Interceptor
	logger      *logrus.Entry
	statService stat.Service
	createTime  int64
}

func MakePuller(rtcService Service, req *dto.PlayReq, claims baseDto.ThkClaims, trackMap map[string]*webrtc.TrackLocalStaticRTP) (*Puller, error) {
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}

	i := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		return nil, err
	}

	if intervalPliFactory, e := intervalpli.NewReceiverInterceptor(); e != nil {
		return nil, e
	} else {
		i.Add(intervalPliFactory)
	}
	statsInterceptorFactory, e := stats.NewInterceptor()
	if e != nil {
		return nil, e
	}
	var statsGetter *stats.Interceptor
	statsInterceptorFactory.OnNewPeerConnection(func(s string, getter stats.Getter) {
		rtcService.Logger().Tracef("OnNewPeerConnection, interceptor, %s", s)
		statsGetter, _ = getter.(*stats.Interceptor)
	})
	i.Add(statsInterceptorFactory)

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i), webrtc.WithSettingEngine(*rtcService.RTCEngine()))
	pc, err := api.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return nil, err
	}

	for _, v := range trackMap {
		if _, err = pc.AddTransceiverFromTrack(v, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendonly}); err != nil {
			_ = pc.Close()
			return nil, err
		}
	}

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  req.OfferSdp,
	}
	err = pc.SetRemoteDescription(offer)
	if err != nil {
		_ = pc.Close()
		return nil, err
	}
	var answer webrtc.SessionDescription
	answer, err = pc.CreateAnswer(nil)
	if err != nil {
		_ = pc.Close()
		return nil, err
	}
	gatherComplete := webrtc.GatheringCompletePromise(pc)
	err = pc.SetLocalDescription(answer)
	if err != nil {
		_ = pc.Close()
		return nil, err
	}
	<-gatherComplete
	answer = *(pc.CurrentLocalDescription())
	key := fmt.Sprintf(SubscribeStreamKey, req.StreamKey, req.UId)
	return &Puller{
		rtcService:  rtcService,
		req:         req,
		claims:      claims,
		key:         key,
		mutex:       &sync.RWMutex{},
		clientSdp:   &offer,
		serverSdp:   &answer,
		peerConn:    pc,
		interceptor: statsInterceptorFactory,
		statsGetter: statsGetter,
		logger:      rtcService.Logger().WithField("room", req.RoomId).WithField("uid", req.UId).WithField("subKey", req.StreamKey),
		statService: rtcService.StatService(),
		createTime:  time.Now().UnixMilli(),
	}, nil
}

func (c *Puller) Key() string {
	return c.key
}

func (c *Puller) SubKey() string {
	return c.subKey
}

func (c *Puller) UId() int64 {
	return c.req.UId
}

func (c *Puller) RoomId() string {
	return c.req.RoomId
}

func (c *Puller) ServerSdp() *webrtc.SessionDescription {
	return c.serverSdp
}

func (c *Puller) Serve() {
	c.peerConn.OnSignalingStateChange(func(state webrtc.SignalingState) {
		c.logger.Tracef("Puller State changed: Singal %v", state)
	})
	c.peerConn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		c.logger.Tracef("Puller State changed: ice %v", connectionState)
	})
	c.peerConn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		c.logger.Tracef("Puller State changed: Connect %v", state)
		if state >= webrtc.PeerConnectionStateDisconnected {
			c.Stop()
		}
		if state == webrtc.PeerConnectionStateConnected {
			c.onConnected()
		}
	})
}

func (c *Puller) onConnected() {
	c.rtcService.Callback().OnPullerConnected(c.RoomId(), c.Key(), c.SubKey(), c.UId(), c.claims)
	senders := c.peerConn.GetSenders()
	for _, sender := range senders {
		encodings := sender.GetParameters().Encodings
		rtcpBuf := make([]byte, 1500)
		for _, en := range encodings {
			initSt := stat.NewStat(sender.Track().ID(), c.key, int64(sender.Track().Kind()), c.SubKey(), c.RoomId(), c.UId())
			go func(_en webrtc.RTPEncodingParameters) {
				currentLost := int64(0)
				for {
					time.Sleep(time.Second) // 1s收集一次
					c.mutex.RLock()
					peerConnection := c.peerConn
					statsGetter := c.statsGetter
					c.mutex.RUnlock()
					if peerConnection != nil {
						if statsGetter != nil {
							if c.peerConn.ConnectionState() == webrtc.PeerConnectionStateConnected {
								if _, _, rtcpErr := sender.Read(rtcpBuf); rtcpErr != nil {
									return
								}
								ssrc := _en.RTPCodingParameters.SSRC
								sts := c.statsGetter.Get(uint32(ssrc))
								if sts != nil {
									st := stat.ReNewStat(initSt, int64(sts.OutboundRTPStreamStats.BytesSent),
										int64(sts.OutboundRTPStreamStats.HeaderBytesSent), int64(sts.OutboundRTPStreamStats.PacketsSent),
										sts.RemoteInboundRTPStreamStats.PacketsLost, sts.RemoteInboundRTPStreamStats.Jitter)
									if c.statService != nil {
										c.statService.CollectStat(st)
									}
									// 产生丢包，发送关键帧
									if sts.InboundRTPStreamStats.PacketsLost != currentLost {
										// c.writeKeyFrame(c.roomId, c.subKey)
										currentLost = sts.InboundRTPStreamStats.PacketsLost
									}
								}
							}
						} else {
							break
						}
					} else {
						break
					}
				}
			}(en)
		}
	}
}

func (c *Puller) Stop() {
	c.mutex.Lock()
	if c.statsGetter != nil {
		if err := c.statsGetter.Close(); err != nil {
			c.logger.Error("statsGetter Close err:", err)
		}
		c.statsGetter = nil
	}
	if c.peerConn != nil {
		if err := c.peerConn.Close(); err != nil {
			c.logger.Errorf("Stop, current connect status: %d, err: %v", c.peerConn.ConnectionState(), err)
		} else {
			c.peerConn = nil
		}
	}
	c.mutex.Unlock()
	c.rtcService.Callback().OnPullerConnected(c.RoomId(), c.Key(), c.SubKey(), c.UId(), c.claims)
}
