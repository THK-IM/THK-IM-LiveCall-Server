package rtc

import (
	"fmt"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/intervalpli"
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/webrtc/v4"
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/stat"
	"sync"
	"time"
)

type Puller struct {
	roomId          string
	uid             int64
	key             string
	subKey          string
	createTime      int64
	mutex           *sync.RWMutex
	clientSdp       *webrtc.SessionDescription
	serverSdp       *webrtc.SessionDescription
	peerConn        *webrtc.PeerConnection
	interceptor     *stats.InterceptorFactory
	statsGetter     *stats.Interceptor
	logger          *logrus.Entry
	statService     stat.Service
	onConnClosed    onConnClosed
	onConnConnected onConnConnected
}

func MakePuller(settingEngine *webrtc.SettingEngine, logger *logrus.Entry, uid int64, roomId, subKey, offerSdp string,
	trackMap map[string]*webrtc.TrackLocalStaticRTP, statService stat.Service,
	onConnConnected onConnConnected, onConnClosed onConnClosed) (*Puller, error) {
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
		logger.Tracef("OnNewPeerConnection, interceptor, %s", s)
		statsGetter, _ = getter.(*stats.Interceptor)
	})
	i.Add(statsInterceptorFactory)

	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i), webrtc.WithSettingEngine(*settingEngine))
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
		SDP:  offerSdp,
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
	key := fmt.Sprintf(SubscribeStreamKey, subKey, uid)
	return &Puller{
		roomId:          roomId,
		uid:             uid,
		key:             key,
		subKey:          subKey,
		createTime:      time.Now().UnixMilli(),
		mutex:           &sync.RWMutex{},
		clientSdp:       &offer,
		serverSdp:       &answer,
		peerConn:        pc,
		interceptor:     statsInterceptorFactory,
		statsGetter:     statsGetter,
		logger:          logger.WithField("room", roomId).WithField("uid", uid).WithField("subKey", subKey),
		onConnClosed:    onConnClosed,
		onConnConnected: onConnConnected,
		statService:     statService,
	}, nil
}

func (c *Puller) Key() string {
	return c.key
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
	c.onConnConnected(c.roomId, c.key, c.subKey, c.uid)
	senders := c.peerConn.GetSenders()
	for _, sender := range senders {
		encodings := sender.GetParameters().Encodings
		rtcpBuf := make([]byte, 1500)
		for _, en := range encodings {
			initSt := stat.NewStat(sender.Track().ID(), c.key, int64(sender.Track().Kind()), c.subKey, c.roomId, c.uid)
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
	if c.onConnClosed != nil {
		c.onConnClosed(c.roomId, c.key, c.subKey, c.uid)
		c.onConnClosed = nil
	}
}
