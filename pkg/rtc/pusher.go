package rtc

import (
	"errors"
	"fmt"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/stats"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/stat"
	"io"
	"strconv"
	"sync"
	"time"
)

type Pusher struct {
	roomId             string
	roomMode           int
	key                string
	uid                int64
	createTime         int64
	mutex              *sync.RWMutex
	pullerMap          map[string]*Puller
	onConnClosed       onConnClosed
	onConnConnected    onConnConnected
	onDataChannelEvent onDataChannelEvent
	logger             *logrus.Entry
	statService        stat.Service
	statsGetter        *stats.Interceptor
	interceptor        *stats.InterceptorFactory
	clientSdp          *webrtc.SessionDescription
	serverSdp          *webrtc.SessionDescription
	peerConn           *webrtc.PeerConnection
	trackMap           map[string]*webrtc.TrackLocalStaticRTP
	dcMap              map[string]*webrtc.DataChannel
}

func (c *Pusher) DcMap() map[string]*webrtc.DataChannel {
	return c.dcMap
}

func MakePusher(settingEngine *webrtc.SettingEngine, logger *logrus.Entry, roomMode int, uid int64, roomId, key, offerSdp string,
	statService stat.Service, onConnConnected onConnConnected, onConnClosed onConnClosed,
	onDataChannelEvent onDataChannelEvent) (*Pusher, error) {
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}
	i := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		return nil, err
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
	if roomMode == dto.ModeVideo {
		if _, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
			_ = pc.Close()
			return nil, err
		}
		if _, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
			_ = pc.Close()
			return nil, err
		}
	} else if roomMode == dto.ModeAudio || roomMode == dto.ModeVoiceRoom {
		if _, err = pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
			_ = pc.Close()
			return nil, err
		}
	} else if roomMode != dto.ModeChat {
		_ = pc.Close()
		return nil, errors.New("mode not support")
	}
	dcMap := make(map[string]*webrtc.DataChannel)
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
	answer = *(pc.LocalDescription())
	return &Pusher{
		roomId:             roomId,
		roomMode:           roomMode,
		uid:                uid,
		key:                key,
		dcMap:              dcMap,
		clientSdp:          &offer,
		serverSdp:          &answer,
		mutex:              &sync.RWMutex{},
		peerConn:           pc,
		statsGetter:        statsGetter,
		statService:        statService,
		onConnConnected:    onConnConnected,
		onDataChannelEvent: onDataChannelEvent,
		onConnClosed:       onConnClosed,
		interceptor:        statsInterceptorFactory,
		logger:             logger.WithField("Pusher", key),
		trackMap:           make(map[string]*webrtc.TrackLocalStaticRTP),
		pullerMap:          make(map[string]*Puller),
		createTime:         time.Now().UnixMilli(),
	}, nil
}

func (c *Pusher) TrackMap() map[string]*webrtc.TrackLocalStaticRTP {
	return c.trackMap
}

func (c *Pusher) Key() string {
	return c.key
}

func (c *Pusher) ServerSdp() *webrtc.SessionDescription {
	return c.serverSdp
}

func (c *Pusher) AddPuller(puller *Puller) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.pullerMap[puller.key] = puller
}

func (c *Pusher) RemovePuller(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.pullerMap[key] != nil {
		delete(c.pullerMap, key)
	}
}

func (c *Pusher) Serve() {
	c.peerConn.OnDataChannel(c.OnDataChannel)
	c.peerConn.OnTrack(c.OnTrack)
	c.peerConn.OnSignalingStateChange(func(state webrtc.SignalingState) {
		c.logger.Tracef("State Changed: Signal %s", state.String())
	})

	c.peerConn.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		c.logger.Tracef("State Changed: ICEState %s", connectionState.String())
	})
	c.peerConn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		c.logger.Tracef("State Changed: PeerConn %s", state.String())
		if state >= webrtc.PeerConnectionStateDisconnected {
			c.Stop()
		}
		if state == webrtc.PeerConnectionStateConnected {

		}
	})
}

func (c *Pusher) OnDataChannel(d *webrtc.DataChannel) {
	if c.roomMode == dto.ModeChat {
		c.onConnConnected(c.roomId, c.key, "", c.uid)
	}
	d.OnOpen(func() {
		c.dcMap[d.Label()] = d
		c.logger.Trace("Datachannel OnOpen: ", d.Label(), c.uid)
		if c.onDataChannelEvent != nil {
			ordered := d.Ordered()
			protocol := d.Protocol()
			negotiated := d.Negotiated()
			event := &DataChannelEvent{
				StreamKey: c.key,
				Label:     d.Label(),
				RoomId:    c.roomId,
				Uid:       c.uid,
				EventType: DataChannelNewEvent,
				Channel: &webrtc.DataChannelInit{
					Ordered:           &ordered,
					MaxPacketLifeTime: d.MaxPacketLifeTime(),
					MaxRetransmits:    d.MaxRetransmits(),
					Protocol:          &protocol,
					Negotiated:        &negotiated,
					ID:                d.ID(),
				},
			}
			c.onDataChannelEvent(event)
		}
	})
	d.OnMessage(func(msg webrtc.DataChannelMessage) {
		c.logger.Trace("Datachannel OnMessage: ", d.Label())
		if c.onDataChannelEvent != nil {
			event := &DataChannelEvent{
				StreamKey: c.key,
				Label:     d.Label(),
				RoomId:    c.roomId,
				Uid:       c.uid,
				EventType: DataChannelMsgEvent,
				Message:   &msg,
			}
			c.onDataChannelEvent(event)
		}
	})
	d.OnClose(func() {
		c.logger.Trace("Datachannel OnClose: ", d.Label())
		if c.onDataChannelEvent != nil {
			event := &DataChannelEvent{
				StreamKey: c.key,
				Label:     d.Label(),
				RoomId:    c.roomId,
				Uid:       c.uid,
				EventType: DataChannelCloseEvent,
			}
			c.dcMap[event.Label] = d
			c.onDataChannelEvent(event)
		}
	})
	d.OnError(func(err error) {
		c.logger.Error("data channel err: ", err)
	})
}

func (c *Pusher) ReceiveDataChannelEvent(event *DataChannelEvent) {
	if c.dcMap == nil {
		c.logger.Error("ReceiveDataChannelEvent, err: ", "dc map is nil")
		return
	}
	if c.peerConn == nil {
		c.logger.Error("ReceiveDataChannelEvent, err: ", "peerConn is nil")
		return
	}
	if event.EventType == DataChannelNewEvent {
		if c.dcMap[event.Label] == nil {
			if _, err := c.peerConn.CreateDataChannel(event.Label, event.Channel); err != nil {
				c.logger.Error("DataChannelNewEvent, err:", err)
			}
		}
	} else if event.EventType == DataChannelMsgEvent {
		d := c.dcMap[event.Label]
		c.logger.Error("ReceiveDataChannelEvent DataChannelMsgEvent event: ", string(event.Message.Data), ", msg: ", c.uid)
		if d != nil {
			if event.Message.IsString {
				dataChannelMsg := DataChanelMsg(string(event.Message.Data))
				if dataChannelMsg != nil {
					if e := d.SendText(*dataChannelMsg); e != nil {
						c.logger.Error("DataChannelMsgEvent, send message err: ", e)
					}
				}
			} else {
				if e := d.Send(event.Message.Data); e != nil {
					c.logger.Error("DataChannelMsgEvent, send message err: ", e)
				}
			}
		} else {
			c.logger.Error("DataChannelMsgEvent, err: ", "channel not existed")
		}
	} else if event.EventType == DataChannelCloseEvent {
		if event.Label != "" {
			d := c.dcMap[event.Label]
			if d != nil {
				if err := d.Close(); err != nil {
					c.logger.Error("DataChannelCloseEvent, err:", err)
				}
				c.dcMap[event.Label] = nil
			}
		}
	}
}

func (c *Pusher) OnTrack(remote *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
	track, err := webrtc.NewTrackLocalStaticRTP(remote.Codec().RTPCodecCapability,
		fmt.Sprintf("%s/id/%d", remote.Kind().String(), c.uid), fmt.Sprintf("%s/stream/%d", remote.Kind().String(), c.uid),
	)
	if err != nil {
		return
	} else {
		c.trackMap[strconv.Itoa(int(remote.SSRC()))] = track
	}
	if (c.roomMode == dto.ModeAudio || c.roomMode == dto.ModeVoiceRoom) && remote.Kind() == webrtc.RTPCodecTypeAudio {
		c.onConnConnected(c.roomId, c.key, "", c.uid)
	} else if c.roomMode == dto.ModeVideo && remote.Kind() == webrtc.RTPCodecTypeVideo {
		c.onConnConnected(c.roomId, c.key, "", c.uid)
	}
	go func(ssrc uint32) {
		initSt := stat.NewStat(remote.ID(), c.key, int64(remote.Kind()), "", c.roomId, c.uid)
		currentLost := int64(0)
		for {
			time.Sleep(time.Second) // 1s收集一次
			c.mutex.RLock()
			peerConnection := c.peerConn
			statsGetter := c.statsGetter
			c.mutex.RUnlock()
			if peerConnection != nil {
				if statsGetter != nil {
					sts := statsGetter.Get(ssrc)
					if sts != nil {
						st := stat.ReNewStat(initSt, int64(sts.InboundRTPStreamStats.PacketsReceived),
							int64(sts.InboundRTPStreamStats.HeaderBytesReceived), int64(sts.InboundRTPStreamStats.PacketsReceived),
							sts.InboundRTPStreamStats.PacketsLost, (sts.InboundRTPStreamStats.Jitter)/float64(remote.Codec().ClockRate))
						if c.statService != nil {
							c.statService.CollectStat(st)
						}
						// 产生丢包，发送关键帧
						if sts.InboundRTPStreamStats.PacketsLost != currentLost {
							currentLost = sts.InboundRTPStreamStats.PacketsLost
						}
					}
				} else {
					break
				}
			} else {
				break
			}
		}
	}(uint32(remote.SSRC()))

	go func(_remote *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP) {
		for {
			c.mutex.RLock()
			peerConnection := c.peerConn
			c.mutex.RUnlock()
			if peerConnection != nil {
				e := _remote.SetReadDeadline(time.Now().Add(2 * time.Second))
				if e != nil {
					c.logger.Error(e)
					continue
				}
				packet, _, readErr := _remote.ReadRTP()
				if readErr != nil {
					c.logger.Error(readErr)
				} else {
					if writeErr := localTrack.WriteRTP(packet); writeErr != nil && !errors.Is(writeErr, io.ErrClosedPipe) {
						c.logger.Error(writeErr)
					}
				}
			} else {
				break
			}
		}
	}(remote, track)
}

func (c *Pusher) WriteKeyFrame() {
	go func() {
		for {
			c.mutex.RLock()
			conn := c.peerConn
			if conn == nil {
				c.mutex.RUnlock()
				break
			}
			if conn.ConnectionState() > webrtc.PeerConnectionStateConnected {
				c.mutex.RUnlock()
				break
			}
			for k := range c.trackMap {
				ssrc, err := strconv.Atoi(k)
				if err == nil {
					if conn != nil {
						if writeErr := conn.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: uint32(ssrc)}}); writeErr != nil {
							c.logger.Errorf("WriteKeyFrame, err: %s", writeErr.Error())
						}
					}
				} else {
					c.logger.Errorf("WriteKeyFrame, err: %s", err.Error())
				}
			}
			c.mutex.RUnlock()
			time.Sleep(2 * time.Second)
		}
	}()
}

func (c *Pusher) Stop() {
	c.logger.Trace("Stop:", c.uid, c.key)
	c.mutex.Lock()
	if c.dcMap != nil {
		if len(c.dcMap) > 0 {
			for _, v := range c.dcMap {
				if err := v.Close(); err != nil {
					c.logger.Error("data channel Close err:", err)
				}
			}
		}
		c.dcMap = make(map[string]*webrtc.DataChannel)
	}
	if c.statsGetter != nil {
		if err := c.statsGetter.Close(); err != nil {
			c.logger.Error("statsGetter Close err:", err)
		}
		c.statsGetter = nil
	}
	if c.peerConn != nil {
		if err := c.peerConn.Close(); err != nil {
			c.logger.Errorf("Stop, current connect err: %v", err)
		} else {
			c.peerConn = nil
		}
	}
	c.mutex.Unlock()
	c.logger.Trace("Stop: call onConnClosed ", c.uid, c.key)
	if c.onConnClosed != nil {
		c.onConnClosed(c.roomId, c.key, "", c.uid)
		c.onConnClosed = nil
	}
	if len(c.pullerMap) > 0 {
		for _, v := range c.pullerMap {
			v.Stop()
		}
		c.pullerMap = make(map[string]*Puller)
	}

}
