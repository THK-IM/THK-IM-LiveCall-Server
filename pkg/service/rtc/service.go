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
	"sync"
	"time"
)

const (
	PublishStreamKey    = "/stream/%s/%s/%s"
	SubscribeStreamKey  = "%s/sub/%s"
	DataChannelCacheKey = "/datachannel/%s"
)

type Service interface {
	// InitServer 初始化服务
	InitServer()
	// RequestPublish 请求推流
	RequestPublish(roomId, uid, offerSdp string) (*Pusher, error)
	// RequestPlay 请求拉流
	RequestPlay(req *dto.PlayReq) (string, string, error)
	// OnPusherConnected 开始推流
	OnPusherConnected(roomId, uId, key, subKey string)
	// OnPusherClosed 关闭
	OnPusherClosed(roomId, uId, key, subKey string)
	OnPullerConnected(roomId, uId, key, subKey string)
	OnPullerClosed(roomId, uId, key, subKey string)
}

func (r serviceImpl) InitServer() {
	r.appCtx.CacheService().Sub(RequestSubscribeEventKey, func(msg string) {
		req := &RequestSubscribeEvent{}
		if err := json.Unmarshal([]byte(msg), req); err != nil {
			r.logger.Error("Sub: ", ResponseSubscribeEventKey, " err: ", err)
			return
		}
		puller, err := r.play(req.RoomId, req.Uid, req.StreamKey, req.OfferSdp)
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
		s, ef := json.Marshal(pullRespEvent)
		if ef != nil {
			r.logger.Error("Sub: ", ResponseSubscribeEventKey, " err: ", err)
		}
		err = r.appCtx.CacheService().Pub(ResponseSubscribeEventKey, string(s))
		if err != nil {
			r.logger.Error("Pub: ", ResponseSubscribeEventKey, " err: ", err)
		}
	})
	r.appCtx.CacheService().Sub(ResponseSubscribeEventKey, func(msg string) {
		pullRespEvent := &ResponseSubscribeEvent{}
		err := json.Unmarshal([]byte(msg), pullRespEvent)
		if err != nil {
			r.logger.Error(err)
			return
		}
		key := fmt.Sprintf(SubscribeStreamKey, pullRespEvent.StreamKey, pullRespEvent.Uid)
		if f := r.onPullRequestEvent[key]; f != nil {
			f(pullRespEvent)
		}
	})

	r.appCtx.CacheService().Sub(NotifyClientNewStreamEventKey, func(msg string) {
		r.logger.Infof("NewStream:%s", msg)
		publishEvent := &PublishEvent{}
		err := json.Unmarshal([]byte(msg), publishEvent)
		if err != nil {
			r.logger.Error("Sub: ", NotifyClientNewStreamEventKey, " err: ", err)
			return
		}
		r.notifyClientNewStream(msg, publishEvent)
	})

	r.appCtx.CacheService().Sub(NotifyClientRemoveStreamEventKey, func(msg string) {
		r.logger.Infof("NotifyClientRemoveStreamEventKey")
		publishEvent := &PublishEvent{}
		err := json.Unmarshal([]byte(msg), publishEvent)
		if err != nil {
			r.logger.Error("Sub: ", NotifyClientRemoveStreamEventKey, " err: ", err)
			return
		}
		r.notifyClientRemoveStream(msg, publishEvent)
	})

	r.appCtx.CacheService().Sub(DataChannelEventKey, func(msg string) {
		r.logger.Error("CacheService Sub: ", DataChannelEventKey, " msg: ", msg)
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
	r.appCtx.CacheService().Sub(room.DestroyRoomEventKey, func(msg string) {
		r.logger.Error("CacheService Sub: ", room.DestroyRoomEventKey, " msg: ", msg)
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
			r.logger.Infof("notifyClientNewStream: %s, uid: %s, %s", msg, v.uid, publishEvent.Uid)
			if v.uid != publishEvent.Uid {
				if dc := v.dcMap[""]; dc != nil {
					r.logger.Infof("notifyClientNewStream, uid: %s : dc is not nil,  %s", v.uid, publishEvent.Uid)
					notifyMsg := NewStreamNotify(msg)
					if notifyMsg != nil {
						if e := dc.SendText(*notifyMsg); e != nil {
							r.logger.Errorf("notifyClientNewStream, err: %s, uid: %s, eventUid: %s", e.Error(), v.uid, publishEvent.Uid)
						}
					}
				} else {
					r.logger.Infof("notifyClientNewStream, uid: %s : dc is nil, event uid: %s ", v.uid, publishEvent.Uid)
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
			if v.uid != publishEvent.Uid {
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

func (r serviceImpl) OnPusherConnected(roomId, uId, key, subKey string) {
	if pusher, err := r.getPusher(roomId, key); err == nil {
		role := room.Audience
		if len(pusher.TrackMap()) > 0 {
			role = room.Broadcast
		}
		event := &PublishEvent{
			RoomId:    roomId,
			Uid:       uId,
			StreamKey: key,
			Role:      role,
		}
		joinTime := time.Now().UnixMilli()
		// 需要通知所有节点的Pushers, 通过Pusher的DataChannel将流新增事件通知到客户端
		if js, e := json.Marshal(event); e == nil {
			if e = r.appCtx.CacheService().Pub(NotifyClientNewStreamEventKey, string(js)); e != nil {
				r.logger.Error("OnPusherConnected: ", roomId, uId, key, subKey, e)
			}
		} else {
			r.logger.Error("OnPusherConnected: ", roomId, uId, key, subKey, e)
		}
		// 房间服务更新参与人
		if err = r.appCtx.RoomService().OnParticipantJoin(roomId, uId, key, joinTime, role); err != nil {
			r.logger.Error("OnParticipantJoin err:", err)
		}
		requestJoinTime, errJ := r.appCtx.RoomService().GetJoinRoomTime(roomId, uId)
		if errJ != nil {
			r.logger.Error("GetJoinRoomTime err:", err)
		}

		// 找出自己请求房间信息时到自己推流成功这段时间内加入的人，推送给当前客户端
		if rm, errRoom := r.appCtx.RoomService().FindRoomById(roomId); errRoom == nil {
			if rm != nil {
				r.logger.Infof("joinTime, uid: %s, requestJoinTime: %d, joinTime:%d, %v", uId, requestJoinTime, joinTime, rm)
				for _, p := range rm.Participants {
					r.logger.Infof("joinTime, uid: %s, requestJoinTime: %d, joinTime:%d, pJoint: %d", uId, requestJoinTime, joinTime, p.JoinTime)
					if p.JoinTime < joinTime && p.JoinTime > requestJoinTime {
						event = &PublishEvent{
							RoomId:    roomId,
							Uid:       p.Uid,
							StreamKey: *p.StreamKey,
							Role:      p.Role,
						}
						r.logger.Infof("ADDNotify, uid: %s : event uid: %s ", pusher.uid, p.Uid)
						msg, errJson := json.Marshal(event)
						if errJson != nil {
							r.logger.Errorf("notifyClientNewStream json err, %v", event)
						}
						if dc := pusher.dcMap[""]; dc != nil {
							r.logger.Infof("notifyClientNewStream, uid: %s : dc is not nil ", pusher.uid)
							notifyMsg := NewStreamNotify(string(msg))
							if notifyMsg != nil {
								if e := dc.SendText(*notifyMsg); e != nil {
									r.logger.Errorf("notifyClientNewStream, %s , uid: %s, event uid:%s", e.Error(), pusher.uid, p.Uid)
								}
							}
						} else {
							r.logger.Infof("notifyClientNewStream, uid: %s , dc is not nil  event uid: %s ", pusher.uid, p.Uid)
						}
					}
				}
			}
		} else {
			r.logger.Errorf("errRoom, %s ", errRoom.Error())
		}
	}

}

func (r serviceImpl) OnPusherClosed(roomId, uId, key, subKey string) {
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
		Uid:       uId,
		StreamKey: key,
	}
	if js, e := json.Marshal(event); e == nil {
		if e = r.appCtx.CacheService().Pub(NotifyClientRemoveStreamEventKey, string(js)); e != nil {
			r.logger.Error("OnPusherClosed: ", roomId, uId, key, subKey, e)
		}
	} else {
		r.logger.Error("OnPusherClosed: ", roomId, uId, key, subKey, e)
	}

	if err := r.appCtx.RoomService().OnParticipantLeave(roomId, uId, key); err != nil {
		r.logger.Error("OnPusherClosed err:", err)
	}

}

func (r serviceImpl) OnPullerConnected(roomId, _, _, subKey string) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()
	pusherMap := r.roomPusherMap[roomId]
	if pusherMap != nil {
		pusher := pusherMap[subKey]
		if pusher != nil {
			go pusher.WriteKeyFrame()
		}
	}
}

func (r serviceImpl) OnPullerClosed(roomId, _, key, subKey string) {
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
	_room, _ := r.appCtx.RoomService().FindRoomById(req.RoomId)
	if _room == nil {
		return "", "", errors.New("room is not existed")
	}
	if _room.Mode == room.ModeChat {
		return "", "", errors.New("no stream pull")
	}
	if _room.Mode != room.ModeAudio && _room.Mode != room.ModeVideo && _room.Mode != room.ModeVoiceRoom {
		return "", "", errors.New("mode error")
	}
	answerSdp := ""
	quit := make(chan bool)
	eventFunc := func(resp *ResponseSubscribeEvent) {
		if resp.Uid == req.Uid &&
			resp.StreamKey == req.StreamKey {
			answerSdp = resp.Answer
			common.SafeOnceWrite(quit, true)
		}
	}

	key := fmt.Sprintf(SubscribeStreamKey, req.StreamKey, req.Uid)
	r.onPullRequestEvent[key] = eventFunc
	defer delete(r.onPullRequestEvent, key)

	pullReqEvent := &RequestSubscribeEvent{
		RoomId:    req.RoomId,
		Uid:       req.Uid,
		OfferSdp:  req.OfferSdp,
		StreamKey: req.StreamKey,
	}
	msg, err := json.Marshal(pullReqEvent)
	if err != nil {
		return "", "", err
	}
	err = r.appCtx.CacheService().Pub(RequestSubscribeEventKey, string(msg))
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

func (r serviceImpl) play(roomId, uid, subStreamKey, offerSdp string) (*Puller, error) {
	pusher, _ := r.getPusher(roomId, subStreamKey)
	if pusher == nil {
		return nil, errors.New("pusher not exist")
	}
	trackMap := pusher.TrackMap()
	if trackMap == nil {
		return nil, errors.New("track not exist")
	}
	puller, err := MakePuller(
		r.settingEngine, r.logger, roomId, uid, subStreamKey, offerSdp,
		trackMap, r.appCtx.StatService(), r.OnPullerConnected, r.OnPullerClosed, r.WriteKeyFrame,
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
		if err = r.appCtx.CacheService().Pub(DataChannelEventKey, content); err != nil {
			r.logger.Error("onDataChannelEvent, ", err)
		}
	}
}

func (r serviceImpl) RequestPublish(roomId, uid, offerSdp string) (*Pusher, error) {
	_room, _ := r.appCtx.RoomService().FindRoomById(roomId)
	if _room == nil {
		return nil, errors.New("room is not existed")
	}
	key := fmt.Sprintf(PublishStreamKey, roomId, uid, common.GenUUid())
	pusher, err := MakePusher(
		r.settingEngine, r.logger, _room.Mode, roomId, uid, key, offerSdp, r.appCtx.StatService(),
		nil, r.OnPusherConnected, r.OnPusherClosed, r.onDataChannelEvent)
	if err != nil {
		r.logger.Errorf("createLiveConnection error: %s", err.Error())
		return nil, err
	}
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	if r.roomPusherMap[roomId] == nil {
		r.roomPusherMap[roomId] = make(map[string]*Pusher, 0)
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

	logEntry := appCtx.Logger().WithField("search_index", "RtcService")
	return &serviceImpl{
		settingEngine:      &settingEngine,
		logger:             logEntry,
		appCtx:             appCtx,
		rwMutex:            &sync.RWMutex{},
		roomPusherMap:      make(map[string]map[string]*Pusher, 0),
		onPullRequestEvent: make(map[string]onPullRequestEvent, 0),
	}
}
