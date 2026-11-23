package room

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseErr "github.com/thk-im/thk-im-base-server/errorx"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/signal"
)

const (
	SessionLockerKey           = "live_server:session:lk:%d"
	SessionKey                 = "live_server:session:%d"
	RLockerKey                 = "live_server:room:lk:%s"
	RCacheKey                  = "live_server:room:%s"
	ParticipantsKey            = "live_server:room:%s:ps"
	ParticipantRequestRoomTime = "live_server:room:%s:uid:%d:r_t"
)

var deleteSessionMapping = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then return redis.call('DEL', KEYS[1]) end
return 0`)

type Service interface {
	// CreateRoom 创建房间
	CreateRoom(req *dto.RoomCreateReq, claims baseDto.ThkClaims) (*dto.RoomJoinResp, error)
	// FindRoomById 通过id查询房间信息
	FindRoomById(id string, claims baseDto.ThkClaims) (*dto.Room, error)
	// DestroyRoom  通过id销毁房间
	DestroyRoom(id string, claims baseDto.ThkClaims) error
	// AddRoomMember 添加房间成员
	AddRoomMember(id string, uId int64, claims baseDto.ThkClaims) error
	// RefuseJoinRoom 拒绝加入
	RefuseJoinRoom(id string, uId int64, isBusy bool, claims baseDto.ThkClaims) error
	// RequestJoinRoom 请求加入房间
	RequestJoinRoom(req *dto.RoomJoinReq, claims baseDto.ThkClaims) (*dto.RoomJoinResp, error)
	// OnUserJoinEvent 房间参与人加入房间回调
	OnUserJoinEvent(event *dto.RoomUserJoinEvent, claims baseDto.ThkClaims) error
	// OnUserLeaveEvent 房间参与人离开房间回调
	OnUserLeaveEvent(event *dto.RoomUserLevelEvent, claims baseDto.ThkClaims) error
	// OnUserPushEvent 房间参与人推流事件
	OnUserPushEvent(event *dto.RoomUserPushStreamEvent, claims baseDto.ThkClaims) error
	OnUserStopEvent(event *dto.RoomUserPushStreamEvent, claims baseDto.ThkClaims) error
	KeepVoiceRoom(room *dto.Room) error
	PreparePublish(roomID string, uid int64, key string) error
	// CheckRooms 检查房间是否关闭
	CheckRooms() error
}

type baseRoomService struct {
	appCtx        *app.Context
	signalService signal.Service
}

func (r baseRoomService) CreateRoom(id, engine string, req *dto.RoomCreateReq, claims baseDto.ThkClaims) (*dto.Room, error) {
	room := &dto.Room{
		Id:           id,
		Engine:       engine,
		Mode:         req.Mode,
		OwnerId:      req.UId,
		MediaParams:  req.MediaParams,
		SessionId:    &req.SessionId,
		CreateTime:   time.Now().UnixMilli(),
		Participants: make([]*dto.Participant, 0),
	}
	room.MediaParams.VideoWidth = 1280
	room.MediaParams.VideoHeight = 720
	room.MediaParams.VideoFps = 30
	room.MediaParams.VideoMaxBitrate = 512 * 8 * 1024
	room.MediaParams.AudioMaxBitrate = 24 * 8 * 1024 // 24KB

	jsonStr, err := room.Json()
	if err != nil {
		return nil, err
	}

	roomCacheKey := r.getRoomCacheKey(id)
	err = r.appCtx.RedisCache().Set(context.Background(), roomCacheKey, jsonStr, time.Hour).Err()
	if err != nil {
		return nil, err
	}

	sessionCacheKey := r.getSessionCacheKey(req.SessionId)
	ttl := time.Minute
	if room.Mode == 4 {
		ttl = time.Hour
	}
	errSession := r.appCtx.RedisCache().Set(context.Background(), sessionCacheKey, room.Id, ttl).Err()
	if errSession != nil {
		return nil, errSession
	}
	return room, nil
}

func (r baseRoomService) FindRoomById(id string, claims baseDto.ThkClaims) (*dto.Room, error) {
	roomJson, err := r.appCtx.RedisCache().Get(context.Background(), r.getRoomCacheKey(id)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	room, e := dto.NewRoomByJson([]byte(roomJson))
	if e != nil {
		return nil, e
	}
	members, errMembers := r.appCtx.RedisCache().HVals(context.Background(), r.getParticipantsCacheKey(id)).Result()
	if errMembers != nil {
		return nil, errMembers
	}
	participants := make([]*dto.Participant, 0)
	for _, m := range members {
		if participant, errJson := dto.NewParticipantByJson([]byte(m)); errJson == nil {
			participants = append(participants, participant)
		}
	}
	room.Participants = participants
	return room, nil
}

func (r baseRoomService) DestroyRoom(id string, claims baseDto.ThkClaims) error {
	lockerKey := fmt.Sprintf(RLockerKey, id)
	locker := r.appCtx.NewLocker(lockerKey, 3000, 3000)
	success, errLock := locker.Lock()
	if errLock != nil {
		return errLock
	}
	if !success {
		return baseErr.ErrInternalServerError
	}
	defer func() {
		_, _ = locker.Release()
	}()

	roomVo, errRoom := r.FindRoomById(id, claims)
	if errRoom != nil {
		return errRoom
	}
	if roomVo == nil {
		r.appCtx.Logger().Trace("DestroyRoom is nil", id)
		return nil
	}

	r.appCtx.Logger().Trace("DestroyRoom sendLiveCallMsg", id)
	errSend := r.sendLiveCallEndMsg(roomVo, claims)
	if errSend != nil {
		r.appCtx.Logger().Error("DestroyRoom sendLiveCallMsg", roomVo, errSend)
	}
	if err := r.appCtx.RedisCache().Del(context.Background(), r.getRoomCacheKey(roomVo.Id)).Err(); err != nil {
		return err
	}
	if roomVo.SessionId != nil {
		if err := deleteSessionMapping.Run(context.Background(), r.appCtx.RedisCache(), []string{r.getSessionCacheKey(*roomVo.SessionId)}, roomVo.Id).Err(); err != nil {
			return err
		}
	}
	if err := r.appCtx.RedisCache().Del(context.Background(), r.getParticipantsCacheKey(roomVo.Id)).Err(); err != nil {
		return err
	}

	return nil
}

func (r baseRoomService) AddRoomMember(id string, uId int64, claims baseDto.ThkClaims) error {
	participant := &dto.Participant{
		UId:    uId,
		Role:   dto.Broadcast,
		Refuse: 0,
	}
	pJson, err := participant.Json()
	if err != nil {
		return err
	}

	cacheKey := r.getParticipantsCacheKey(id)
	err = r.appCtx.RedisCache().HSetNX(context.Background(), cacheKey, fmt.Sprintf("%d", uId), pJson).Err()
	return err
}

func (r baseRoomService) RefuseJoinRoom(id string, uId int64, isBusy bool, claims baseDto.ThkClaims) error {
	refuse := 1
	if isBusy {
		refuse = 2
	}
	participant := &dto.Participant{
		UId:    uId,
		Role:   dto.Broadcast,
		Refuse: refuse,
	}
	pJson, err := participant.Json()
	if err != nil {
		return nil
	}

	cacheKey := r.getParticipantsCacheKey(id)
	err = r.appCtx.RedisCache().HSet(context.Background(), cacheKey, fmt.Sprintf("%d", uId), pJson).Err()
	return err
}

func (r baseRoomService) OnUserJoinEvent(event *dto.RoomUserJoinEvent, claims baseDto.ThkClaims) error {
	r.appCtx.Logger().Tracef("OnUserJoinEvent %v", event)
	room, errRoom := r.FindRoomById(event.RoomId, claims)
	if errRoom != nil {
		return errRoom
	}
	if room == nil {
		return nil
	}
	_, err := r.updateMember(room.Id, event.UserId, func(m *memberState) bool {
		if m.JoinTime > 0 && m.LeaveTime == 0 {
			return false
		}
		m.JoinTime, m.LeaveTime, m.Refuse = event.Timestamp, 0, 0
		return true
	})
	if err != nil {
		return err
	}
	return r.appCtx.RedisCache().Expire(context.Background(), r.getRoomCacheKey(room.Id), time.Hour).Err()
}

func (r baseRoomService) OnUserLeaveEvent(event *dto.RoomUserLevelEvent, claims baseDto.ThkClaims) error {
	room, errRoom := r.FindRoomById(event.RoomId, claims)
	if errRoom != nil {
		return errRoom
	}
	if room == nil {
		return nil
	}

	changed, err := r.updateMember(room.Id, event.UserId, func(m *memberState) bool {
		if m.LeaveTime != 0 {
			return false
		}
		m.LeaveTime, m.StreamKey, m.IssuedKey = event.Timestamp, "", ""
		return true
	})
	if err != nil || !changed {
		return err
	}

	count := 0
	for _, p := range room.Participants {
		if p.UId != event.UserId && p.LeaveTime == 0 && p.JoinTime > 0 {
			count++
		}
	}
	r.appCtx.Logger().Tracef("OnParticipantLeave %v, user count %d", event, count)

	if count == 0 {
		errDestroy := r.DestroyRoom(room.Id, claims)
		if errDestroy != nil {
			r.appCtx.Logger().Error("OnParticipantLeave DestroyRoom", event, errDestroy)
		}
	}
	return nil
}

func (r baseRoomService) OnUserPushEvent(event *dto.RoomUserPushStreamEvent, claims baseDto.ThkClaims) error {
	return r.streamEvent(event, true, claims)
}

func (r baseRoomService) CheckRooms() error {
	return nil
}

func (r baseRoomService) sendLiveCallEndMsg(room *dto.Room, claims baseDto.ThkClaims) error {
	return r.signalService.SendLiveCallMsgByEnded(room, claims)
}

func (r baseRoomService) getRoomCacheKey(roomId string) string {
	return fmt.Sprintf(RCacheKey, roomId)
}

func (r baseRoomService) getSessionCacheKey(sessionId int64) string {
	return fmt.Sprintf(SessionKey, sessionId)
}

func (r baseRoomService) getParticipantsCacheKey(roomId string) string {
	return fmt.Sprintf(ParticipantsKey, roomId)
}

func (r baseRoomService) getParticipantRequestRoomTimeKey(roomId string, userId int64) string {
	return fmt.Sprintf(ParticipantRequestRoomTime, roomId, userId)
}
