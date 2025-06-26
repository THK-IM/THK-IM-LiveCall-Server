package rtc

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pion/ice/v4"
	"github.com/pion/webrtc/v4"
	"github.com/sirupsen/logrus"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/common"
	"github.com/thk-im/thk-im-livecall-server/pkg/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/errorx"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/room"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/stat"
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
	RequestPublish(req *dto.PublishReq, claims baseDto.ThkClaims) (*Pusher, error)
	// RequestPlay 请求拉流
	RequestPlay(req *dto.PlayReq, claims baseDto.ThkClaims) (string, string, error)
	// WriteKeyFrame 写关键帧
	WriteKeyFrame(roomId string, streamKey string)
	// RTCEngine 引擎
	RTCEngine() *webrtc.SettingEngine
	// Logger 日志
	Logger() *logrus.Entry
	// Callback 回调
	Callback() Callback
	// StatService 质量分析服务
	StatService() stat.Service
}

func (r serviceImpl) InitServer() {
	r.appCtx.RoomCache().Sub(RequestSubscribeEventKey, func(msg string) {
		event := &RequestSubscribeEvent{}
		if err := json.Unmarshal([]byte(msg), event); err != nil {
			r.logger.Error("Sub: ", ResponseSubscribeEventKey, " err: ", err)
			return
		}
		puller, err := r.play(event.Req, event.Claims)
		if err != nil {
			r.logger.Error("Sub: ", ResponseSubscribeEventKey, " err: ", err)
			return
		}
		if puller == nil {
			r.logger.Tracef("Sub: %s, stream: %s is not existed", ResponseSubscribeEventKey, event.Req.StreamKey)
			return
		}
		answer := base64.StdEncoding.EncodeToString([]byte(puller.ServerSdp().SDP))
		pullRespEvent := &ResponseSubscribeEvent{
			Answer:    answer,
			StreamKey: event.Req.StreamKey,
			Uid:       event.Req.UId,
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
		r.logger.Trace("CacheService Sub: ", room.DestroyRoomEventKey, " msg: ", msg)
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

func (r serviceImpl) RequestPlay(req *dto.PlayReq, claims baseDto.ThkClaims) (string, string, error) {
	rm, _ := r.appCtx.RoomService().FindRoomById(req.RoomId)
	if rm == nil {
		return "", "", errorx.ErrRoomNotExisted
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
		Req:    req,
		Claims: claims,
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

func (r serviceImpl) RequestPublish(req *dto.PublishReq, claims baseDto.ThkClaims) (*Pusher, error) {
	_room, _ := r.appCtx.RoomService().FindRoomById(req.RoomId)
	if _room == nil {
		return nil, errors.New("room is not existed")
	}
	key := fmt.Sprintf(PublishStreamKey, req.RoomId, req.UId, common.GenUUid())
	pusher, err := MakePusher(r, req, claims, _room.Mode, key)
	if err != nil {
		r.logger.Errorf("createLiveConnection error: %s", err.Error())
		return nil, err
	}
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	if r.roomPusherMap[req.RoomId] == nil {
		r.roomPusherMap[req.RoomId] = make(map[string]*Pusher)
	}
	r.roomPusherMap[req.RoomId][key] = pusher
	pusher.Serve()
	return pusher, nil
}

func (r serviceImpl) WriteKeyFrame(roomId string, streamKey string) {
	r.logger.Tracef("puller write key frame, roomId: %s, streamKey: %s", roomId, streamKey)
	pusher, err := r.getPusher(roomId, streamKey)
	if err != nil {
		return
	}
	pusher.WriteKeyFrame()
}

func (r serviceImpl) RTCEngine() *webrtc.SettingEngine {
	return r.settingEngine
}

func (r serviceImpl) Logger() *logrus.Entry {
	return r.logger
}

func (r serviceImpl) Callback() Callback {
	return r
}

func (r serviceImpl) StatService() stat.Service {
	return r.appCtx.StatService()
}

func (r serviceImpl) OnPusherConnected(roomId, key, subKey string, uId int64, claims baseDto.ThkClaims) {
	if pusher, err := r.getPusher(roomId, key); err == nil {
		role := dto.Audience
		if len(pusher.TrackMap()) > 0 {
			role = dto.Broadcast
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
		participants, errParticipants := r.appCtx.RoomService().OnParticipantJoin(roomId, key, pusherJoinTime, role, uId)
		if errParticipants != nil {
			r.logger.Error("OnParticipantJoin err:", err)
		}
		// 通过默认DC通道发给自己
		if dc := pusher.dcMap[""]; dc != nil {
			for _, p := range participants {
				event = &PublishEvent{
					RoomId:    roomId,
					UId:       p.UId,
					StreamKey: *p.StreamKey,
					Role:      p.Role,
				}
				if msg, errJson := json.Marshal(event); errJson == nil {
					notifyMsg := NewStreamNotify(string(msg))
					if notifyMsg != nil {
						if e := dc.SendText(*notifyMsg); e != nil {
							r.logger.Errorf("notifyClientNewStream, %s , uid: %d, event uid:%d", e.Error(), pusher.UId(), p.UId)
						}
					}
				}
			}
		}
	}

}

func (r serviceImpl) OnPusherSteaming(roomId, key, subKey string, uid int64, time int64, claims baseDto.ThkClaims) {
	// 30s通知一次
	if time%(30*1000) != 0 {
		return
	}
	go func() {
		err := r.appCtx.RoomService().OnParticipantPushStreamEvent(roomId, key, uid, claims)
		if err != nil {
			r.logger.Error("OnPusherSteaming: ", roomId, uid, key, subKey, err)
		}
	}()
}

func (r serviceImpl) OnPusherClosed(roomId, key, subKey string, uId int64, claims baseDto.ThkClaims) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()
	r.logger.Trace("OnPusherConnected: ", roomId, uId, key, subKey)
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

func (r serviceImpl) OnPullerConnected(roomId, _, subKey string, _ int64, claims baseDto.ThkClaims) {
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

func (r serviceImpl) OnPullStreaming(roomId, key, subKey string, uid int64, time int64, claims baseDto.ThkClaims) {
}

func (r serviceImpl) OnPullerClosed(roomId, key, subKey string, _ int64, claims baseDto.ThkClaims) {
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

func (r serviceImpl) OnDataChannelEvent(event *DataChannelEvent, claims baseDto.ThkClaims) {
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

func (r serviceImpl) getPusher(roomId, publishKey string) (*Pusher, error) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()
	if r.roomPusherMap[roomId] == nil {
		return nil, nil
	} else {
		return r.roomPusherMap[roomId][publishKey], nil
	}
}

func (r serviceImpl) userSubscribeStreamKey(streamKey string, uId int64) string {
	return fmt.Sprintf(SubscribeStreamKey, streamKey, uId)
}

func (r serviceImpl) play(req *dto.PlayReq, claims baseDto.ThkClaims) (*Puller, error) {
	pusher, _ := r.getPusher(req.RoomId, req.StreamKey)
	if pusher == nil {
		return nil, errorx.ErrPusherNotExisted
	}
	trackMap := pusher.TrackMap()
	if trackMap == nil {
		return nil, errorx.ErrPusherNotExisted
	}
	puller, err := MakePuller(r, req, claims, trackMap)
	if err != nil {
		return nil, err
	}
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	pusher.AddPuller(puller)
	puller.Serve()
	return puller, nil
}

func (r serviceImpl) notifyClientNewStream(msg string, publishEvent *PublishEvent) {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()
	// 通知当前节点下房间内所有用户 有新的流接入
	pusherMap := r.roomPusherMap[publishEvent.RoomId]
	if pusherMap != nil {
		for _, v := range pusherMap {
			r.logger.Tracef("notifyClientNewStream: %s, uid: %d, %d", msg, v.UId(), publishEvent.UId)
			if v.UId() != publishEvent.UId {
				if dc := v.dcMap[""]; dc != nil {
					r.logger.Tracef("notifyClientNewStream, uid: %d : dc is not nil,  %d", v.UId(), publishEvent.UId)
					notifyMsg := NewStreamNotify(msg)
					if notifyMsg != nil {
						if e := dc.SendText(*notifyMsg); e != nil {
							r.logger.Errorf("notifyClientNewStream, err: %s, uid: %d, eventUid: %d", e.Error(), v.UId(), publishEvent.UId)
						}
					}
				} else {
					r.logger.Tracef("notifyClientNewStream, uid: %d : dc is nil, event uid: %d ", v.UId(), publishEvent.UId)
				}
			}
		}
	}
}

func (r serviceImpl) notifyClientRemoveStream(msg string, publishEvent *PublishEvent) {
	r.logger.Trace("notifyClientRemoveStream", msg)
	// 通知当前节点下房间内所有用户 有新的流移除
	pusherMap := r.roomPusherMap[publishEvent.RoomId]
	if pusherMap != nil {
		for _, v := range pusherMap {
			if v.UId() != publishEvent.UId {
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
	settingEngine.LoggerFactory = &LoggerFactory{LogEntry: appCtx.Logger()}
	if source.TcpPort > 0 {
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
		settingEngine.SetNetworkTypes([]webrtc.NetworkType{
			webrtc.NetworkTypeUDP4,
			webrtc.NetworkTypeUDP6,
		})
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
