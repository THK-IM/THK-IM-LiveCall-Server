package rtc

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pion/ice/v2"
	"github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/common"
	"github.com/thk-im/thk-im-livecall-server/pkg/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/room"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/room/model"
	"net"
	"sync"
	"time"
)

const (
	PublishStreamKey   = "/stream/%s/%d/%s"
	SubscribeStreamKey = "%s/sub/%d"
)

type Service interface {
	// InitServer 初始化服务
	InitServer()
	// RequestPublish 请求推流
	RequestPublish(roomId, offerSdp string, uid int64) (*Pusher, error)
	// RequestPlay 请求拉流
	RequestPlay(req *dto.PlayReq) (string, string, error)
	// OnPusherConnected 开始推流
	OnPusherConnected(roomId, key, subKey string, uid int64)
	// OnPusherClosed 关闭
	OnPusherClosed(roomId, key, subKey string, uid int64)
	OnPullerConnected(roomId, key, subKey string, uid int64)
	OnPullerClosed(roomId, key, subKey string, uid int64)
}

func (r serviceImpl) InitServer() {
	r.appCtx.RoomCache().Sub(RequestSubscribeEventKey, func(msg string) {
		req := &RequestSubscribeEvent{}
		if err := json.Unmarshal([]byte(msg), req); err != nil {
			r.logger.Error("Sub: ", ResponseSubscribeEventKey, " err: ", err)
			return
		}
		puller, err := r.play(req.RoomId, req.StreamKey, req.OfferSdp, req.Uid)
		if err != nil {
			r.logger.Error("Sub: ", ResponseSubscribeEventKey, " err: ", err)
			return
		}
		if puller == nil {
			r.logger.Infof("Sub: %s, stream: %s is not existed", ResponseSubscribeEventKey, req.StreamKey)
			return
		}
		answer := base64.StdEncoding.EncodeToString([]byte(puller.ServerSdp().SDP))
		pullRespEvent := &ResponseSubscribeEvent{
			Answer:    answer,
			StreamKey: req.StreamKey,
			Uid:       req.Uid,
		}
		eventJson, errJson := json.Marshal(pullRespEvent)
		if errJson != nil {
			r.logger.Error("Sub: ", ResponseSubscribeEventKey, " err: ", errJson)
		}
		err = r.appCtx.RoomCache().Pub(ResponseSubscribeEventKey, string(eventJson))
		if err != nil {
			r.logger.Error("Pub: ", ResponseSubscribeEventKey, " err: ", err)
		}
	})
	r.appCtx.RoomCache().Sub(ResponseSubscribeEventKey, func(msg string) {
		pullRespEvent := &ResponseSubscribeEvent{}
		err := json.Unmarshal([]byte(msg), pullRespEvent)
		if err != nil {
			r.logger.Error(err)
			return
		}
		key := r.userSubscribeStreamKey(pullRespEvent.StreamKey, pullRespEvent.Uid)
		if f := r.onPullRequestEvent[key]; f != nil {
			f(pullRespEvent)
		}
	})

	r.appCtx.RoomCache().Sub(NotifyClientNewStreamEventKey, func(msg string) {
		publishEvent := &PublishEvent{}
		err := json.Unmarshal([]byte(msg), publishEvent)
		if err != nil {
			r.logger.Error("Sub: ", NotifyClientNewStreamEventKey, " err: ", err)
			return
		}
		r.notifyClientNewStream(msg, publishEvent)
	})

	r.appCtx.RoomCache().Sub(NotifyClientRemoveStreamEventKey, func(msg string) {
		publishEvent := &PublishEvent{}
		err := json.Unmarshal([]byte(msg), publishEvent)
		if err != nil {
			r.logger.Error("Sub: ", NotifyClientRemoveStreamEventKey, " err: ", err)
			return
		}
		r.notifyClientRemoveStream(msg, publishEvent)
	})

	r.appCtx.RoomCache().Sub(DataChannelEventKey, func(msg string) {
		event := &DataChannelEvent{}
		if err := json.Unmarshal([]byte(msg), event); err != nil {
			r.logger.Error("Sub: ", DataChannelEventKey, " err: ", err)
			return
		}
		r.rwMutex.RLock()
		defer r.rwMutex.RUnlock()
		pusherMap := r.roomPusherMap[event.RoomId]
		for _, v := range pusherMap {
			if v.key != event.StreamKey {
				v.ReceiveDataChannelEvent(event)
			}
		}
	})
	// 停止房间内stream的消息监听
	r.appCtx.RoomCache().Sub(room.DestroyRoomEventKey, func(msg string) {
		r.logger.Info("CacheService Sub: ", room.DestroyRoomEventKey, " msg: ", msg)
		event := &room.DestroyRoomEvent{}
		if err := json.Unmarshal([]byte(msg), event); err != nil {
			r.logger.Error("Sub: ", room.DestroyRoomEventKey, " err: ", err)
			return
		}
		r.rwMutex.RLock()
		defer r.rwMutex.RUnlock()
		pusherMap := r.roomPusherMap[event.RoomId]
		for _, v := range pusherMap {
			v.Stop()
		}
	})
}

func (r serviceImpl) notifyClientNewStream(msg string, publishEvent *PublishEvent) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()
	// 通知当前节点下房间内所有用户 有新的流接入
	pusherMap := r.roomPusherMap[publishEvent.RoomId]
	if pusherMap != nil {
		for _, v := range pusherMap {
			r.logger.Infof("notifyClientNewStream: %s, uid: %d, %d", msg, v.uid, publishEvent.UId)
			if v.uid != publishEvent.UId {
				if dc := v.dcMap[""]; dc != nil {
					r.logger.Infof("notifyClientNewStream, uid: %d : dc is not nil,  %d", v.uid, publishEvent.UId)
					notifyMsg := NewStreamNotify(msg)
					if notifyMsg != nil {
						if e := dc.SendText(*notifyMsg); e != nil {
							r.logger.Errorf("notifyClientNewStream, err: %s, uid: %d, eventUid: %d", e.Error(), v.uid, publishEvent.UId)
						}
					}
				} else {
					r.logger.Infof("notifyClientNewStream, uid: %d : dc is nil, event uid: %d ", v.uid, publishEvent.UId)
				}
			}
		}
	}
}

func (r serviceImpl) notifyClientRemoveStream(msg string, publishEvent *PublishEvent) {
	r.logger.Info("notifyClientRemoveStream", msg)
	// 通知当前节点下房间内所有用户 有新的流移除
	pusherMap := r.roomPusherMap[publishEvent.RoomId]
	if pusherMap != nil {
		for _, v := range pusherMap {
			if v.uid != publishEvent.UId {
				if dc := v.dcMap[""]; dc != nil {
					notifyMsg := RemoveStreamNotify(msg)
					if notifyMsg != nil {
						if e := dc.SendText(*notifyMsg); e != nil {
							r.logger.Error("notifyClientRemoveStream, ", e)
						}
					}
				}
			}
		}
	}
}

func (r serviceImpl) OnPusherConnected(roomId, key, subKey string, uId int64) {
	if pusher, err := r.getPusher(roomId, key); err == nil {
		role := model.Audience
		if len(pusher.TrackMap()) > 0 {
			role = model.Broadcast
		}
		event := &PublishEvent{
			RoomId:    roomId,
			UId:       uId,
			StreamKey: key,
			Role:      role,
		}
		pusherJoinTime := time.Now().UnixMilli()

		// 需要通知所有节点的Pushers, 通过Pusher的DataChannel将流新增事件通知到客户端
		if js, e := json.Marshal(event); e == nil {
			if e = r.appCtx.RoomCache().Pub(NotifyClientNewStreamEventKey, string(js)); e != nil {
				r.logger.Error("OnPusherConnected: ", roomId, uId, key, subKey, e)
			}
		} else {
			r.logger.Error("OnPusherConnected: ", roomId, uId, key, subKey, e)
		}

		// 房间服务更新参与人
		if err = r.appCtx.RoomService().OnParticipantJoin(roomId, key, pusherJoinTime, role, uId); err != nil {
			r.logger.Error("OnParticipantJoin err:", err)
		}

		// 当前用户请求加入房间的时间
		requestJoinTime, errJ := r.appCtx.RoomService().GetRequestJoinRoomTime(roomId, uId)
		if errJ != nil {
			r.logger.Error("GetRequestJoinRoomTime err:", err)
		}

		// 遍历房间参与人，看谁是在自己请求加入到正式加入之间进入的房间的用户，通过channel下发给当前参与人
		if rm, errRoom := r.appCtx.RoomService().FindRoomById(roomId); errRoom == nil {
			if rm != nil {
				for _, p := range rm.Participants {
					r.logger.Infof("notifyClientNewStream uId: %d p: %d, JoinTime: %d, requestJoinTime: %d, pusherJoinTime: %d", uId, p.UId, p.JoinTime, requestJoinTime, pusherJoinTime)
					if p.JoinTime < pusherJoinTime && p.JoinTime > requestJoinTime {
						event = &PublishEvent{
							RoomId:    roomId,
							UId:       p.UId,
							StreamKey: *p.StreamKey,
							Role:      p.Role,
						}
						msg, errJson := json.Marshal(event)
						if errJson != nil {
							r.logger.Errorf("notifyClientNewStream json err, %v", event)
						}
						if dc := pusher.dcMap[""]; dc != nil {
							notifyMsg := NewStreamNotify(string(msg))
							if notifyMsg != nil {
								if e := dc.SendText(*notifyMsg); e != nil {
									r.logger.Errorf("notifyClientNewStream, %s , uid: %d, event uid:%d", e.Error(), pusher.uid, p.UId)
								}
							}
						} else {
							r.logger.Infof("notifyClientNewStream, uid: %d , dc is not nil  event uid: %d ", pusher.uid, p.UId)
						}
					}
				}
			}
		} else {
			r.logger.Errorf("errRoom, %s ", errRoom.Error())
		}

		pusher.WriteKeyFrame()
	}

}

func (r serviceImpl) OnPusherClosed(roomId, key, subKey string, uId int64) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()
	r.logger.Info("OnPusherConnected: ", roomId, uId, key, subKey)
	pusherMap := r.roomPusherMap[roomId]
	if pusherMap != nil {
		if pusherMap[key] != nil {
			delete(pusherMap, key)
		}
	}
	// 需要通知所有节点的Pushers, 通过Pusher的DataChannel将流移除事件通知到客户端
	event := &PublishEvent{
		RoomId:    roomId,
		UId:       uId,
		StreamKey: key,
	}
	if js, e := json.Marshal(event); e == nil {
		if e = r.appCtx.RoomCache().Pub(NotifyClientRemoveStreamEventKey, string(js)); e != nil {
			r.logger.Error("OnPusherClosed: ", roomId, uId, key, subKey, e)
		}
	} else {
		r.logger.Error("OnPusherClosed: ", roomId, uId, key, subKey, e)
	}

	if err := r.appCtx.RoomService().OnParticipantLeave(roomId, key, uId); err != nil {
		r.logger.Error("OnPusherClosed err:", err)
	}

}

func (r serviceImpl) OnPullerConnected(roomId, _, subKey string, _ int64) {
	// r.rwMutex.RLock()
	// defer r.rwMutex.RUnlock()
	// pusherMap := r.roomPusherMap[roomId]
	// if pusherMap != nil {
	// 	pusher := pusherMap[subKey]
	// 	if pusher != nil {
	// 		go pusher.WriteKeyFrame()
	// 	}
	// }
}

func (r serviceImpl) OnPullerClosed(roomId, key, subKey string, _ int64) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()
	if r.roomPusherMap[roomId] == nil {
		return
	} else {
		if r.roomPusherMap[roomId][subKey] != nil {
			r.roomPusherMap[roomId][subKey].RemovePuller(key)
		}
	}
}

func (r serviceImpl) getPusher(roomId, publishKey string) (*Pusher, error) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()
	if r.roomPusherMap[roomId] == nil {
		return nil, nil
	} else {
		return r.roomPusherMap[roomId][publishKey], nil
	}
}

func (r serviceImpl) RequestPlay(req *dto.PlayReq) (string, string, error) {
	rm, _ := r.appCtx.RoomService().FindRoomById(req.RoomId)
	if rm == nil {
		return "", "", errors.New("room is not existed")
	}
	if rm.Mode == model.ModeChat {
		return "", "", errors.New("no stream pull")
	}
	if rm.Mode != model.ModeAudio && rm.Mode != model.ModeVideo && rm.Mode != model.ModeVoiceRoom {
		return "", "", errors.New("mode error")
	}
	answerSdp := ""
	quit := make(chan bool)
	eventFunc := func(resp *ResponseSubscribeEvent) {
		if resp.Uid == req.UId && resp.StreamKey == req.StreamKey {
			answerSdp = resp.Answer
			common.SafeOnceWrite(quit, true)
		}
	}

	key := r.userSubscribeStreamKey(req.StreamKey, req.UId)
	r.onPullRequestEvent[key] = eventFunc
	defer delete(r.onPullRequestEvent, key)

	pullReqEvent := &RequestSubscribeEvent{
		RoomId:    req.RoomId,
		Uid:       req.UId,
		OfferSdp:  req.OfferSdp,
		StreamKey: req.StreamKey,
	}
	msg, err := json.Marshal(pullReqEvent)
	if err != nil {
		return "", "", err
	}
	err = r.appCtx.RoomCache().Pub(RequestSubscribeEventKey, string(msg))
	if err != nil {
		return "", "", err
	}
	go func(q chan bool) {
		time.Sleep(2 * time.Second)
		common.SafeOnceWrite(q, true)
	}(quit)

	<-quit
	if answerSdp == "" {
		return answerSdp, "", errors.New("timeout")
	} else {
		return answerSdp, key, nil
	}
}

func (r serviceImpl) play(roomId, subStreamKey, offerSdp string, uid int64) (*Puller, error) {
	pusher, _ := r.getPusher(roomId, subStreamKey)
	if pusher == nil {
		return nil, errors.New("pusher not exist")
	}
	trackMap := pusher.TrackMap()
	if trackMap == nil {
		return nil, errors.New("track not exist")
	}
	puller, err := MakePuller(
		r.settingEngine, r.logger, uid, roomId, subStreamKey, offerSdp,
		trackMap, r.appCtx.StatService(), r.OnPullerConnected, r.OnPullerClosed,
	)
	if err != nil {
		return nil, err
	}
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	pusher.AddPuller(puller)
	puller.Serve()
	return puller, nil
}

func (r serviceImpl) onDataChannelEvent(event *DataChannelEvent) {
	if d, err := json.Marshal(event); err != nil {
		r.logger.Error("onDataChannelEvent, ", err)
	} else {
		content := string(d)
		// DataChannel事件发送到所有的节点
		if err = r.appCtx.RoomCache().Pub(DataChannelEventKey, content); err != nil {
			r.logger.Error("onDataChannelEvent, ", err)
		}
	}
}

func (r serviceImpl) RequestPublish(roomId, offerSdp string, uid int64) (*Pusher, error) {
	_room, _ := r.appCtx.RoomService().FindRoomById(roomId)
	if _room == nil {
		return nil, errors.New("room is not existed")
	}
	key := fmt.Sprintf(PublishStreamKey, roomId, uid, common.GenUUid())
	pusher, err := MakePusher(
		r.settingEngine, r.logger, _room.Mode, uid, roomId, key, offerSdp, r.appCtx.StatService(),
		r.OnPusherConnected, r.OnPusherClosed, r.onDataChannelEvent)
	if err != nil {
		r.logger.Errorf("createLiveConnection error: %s", err.Error())
		return nil, err
	}
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	if r.roomPusherMap[roomId] == nil {
		r.roomPusherMap[roomId] = make(map[string]*Pusher)
	}
	r.roomPusherMap[roomId][key] = pusher
	pusher.Serve()
	return pusher, nil
}

func (r serviceImpl) WriteKeyFrame(roomId string, subKey string) {
	r.logger.Infof("puller write key frame, roomId: %s, subKey: %s", roomId, subKey)
	pusher, err := r.getPusher(roomId, subKey)
	if err != nil {
		return
	}
	pusher.WriteKeyFrame()
}

func (r serviceImpl) userSubscribeStreamKey(streamKey string, uId int64) string {
	return fmt.Sprintf(SubscribeStreamKey, streamKey, uId)
}

type serviceImpl struct {
	settingEngine      *webrtc.SettingEngine
	appCtx             *app.Context
	logger             *logrus.Entry
	rwMutex            *sync.RWMutex
	roomPusherMap      map[string]map[string]*Pusher
	onPullRequestEvent map[string]onPullRequestEvent
}

func NewRtcService(source *conf.Rtc, appCtx *app.Context) Service {
	settingEngine := webrtc.SettingEngine{}
	if source.TcpPort > 0 {
		// Enable support only for TCP ICE candidates.
		settingEngine.SetNetworkTypes([]webrtc.NetworkType{
			webrtc.NetworkTypeTCP4,
			webrtc.NetworkTypeTCP6,
		})
		tcpListener, err := net.ListenTCP("tcp", &net.TCPAddr{
			IP:   net.IP{0, 0, 0, 0},
			Port: source.TcpPort,
		})
		if err != nil {
			panic(err)
		}
		tcpMux := webrtc.NewICETCPMux(nil, tcpListener, 8*1024)
		settingEngine.SetICETCPMux(tcpMux)
		settingEngine.SetICETimeouts(
			time.Duration(source.Timeout)*time.Millisecond,
			time.Duration(source.Timeout)*time.Millisecond,
			time.Duration(source.Timeout)*time.Millisecond,
		)
	} else if source.UdpPort > 0 {
		mux, err := ice.NewMultiUDPMuxFromPort(source.UdpPort)
		if err != nil {
			panic(err)
		}
		settingEngine.SetICETimeouts(
			time.Duration(source.Timeout)*time.Millisecond,
			time.Duration(source.Timeout)*time.Millisecond,
			time.Duration(source.Timeout)*time.Millisecond,
		)
		settingEngine.SetICEUDPMux(mux)
		settingEngine.SetNAT1To1IPs([]string{source.NodeIp}, webrtc.ICECandidateTypeHost)
	} else {
		panic(errors.New("error rtc"))
	}

	logEntry := appCtx.Logger().WithField("search_index", "RtcService")
	return &serviceImpl{
		settingEngine:      &settingEngine,
		logger:             logEntry,
		appCtx:             appCtx,
		rwMutex:            &sync.RWMutex{},
		roomPusherMap:      make(map[string]map[string]*Pusher),
		onPullRequestEvent: make(map[string]onPullRequestEvent),
	}
}
