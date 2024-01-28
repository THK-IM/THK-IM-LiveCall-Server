package room

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/cache"
	"strconv"
	"time"
)

const (
	CacheKey                   = "live_server:room:%s"
	ParticipantsKey            = "live_server:room:%s:participants"
	ParticipantRequestRoomTime = "live_server:room:%s:uid:%d:request_time"
)

type Service interface {
	CreateRoom(*dto.RoomCreateReq) (*Room, error)
	FindRoomById(id string) (*Room, error)
	DestroyRoom(id string) error
	// NodePublicIp 所在节点公网ip地址
	NodePublicIp() string
	JoinRoom(req *dto.RoomJoinReq) (*Room, error)
	GetRequestJoinRoomTime(roomId string, uId int64) (int64, error)
	OnParticipantJoin(roomId, streamKey string, joinTime int64, role int, uId int64) error
	OnParticipantLeave(roomId, streamKey string, uId int64) error
}

func NewService(
	cache cache.Service,
	logger *logrus.Entry,
) Service {
	return &ServiceImpl{
		logger: logger,
		cache:  cache,
	}
}

type ServiceImpl struct {
	logger   *logrus.Entry
	cache    cache.Service
	publicIp string
}

func (r *ServiceImpl) NodePublicIp() string {
	return r.publicIp
}

func (r *ServiceImpl) CreateRoom(req *dto.RoomCreateReq) (*Room, error) {
	id := fmt.Sprintf("%d_%d", (time.Now().UnixNano())/int64(time.Second), req.UId)
	room := &Room{
		Id:         id,
		Mode:       req.Mode,
		OwnerId:    req.UId,
		CreateTime: time.Now().UnixMilli(),
	}
	roomCacheKey := r.getRoomCacheKey(room.Id)
	if value, err := r.cache.Get(roomCacheKey); err != nil {
		if !errors.Is(err, redis.Nil) {
			return nil, err
		}
	} else {
		strValue, ok := value.(string)
		if ok && len(strValue) > 0 {
			return nil, errors.New("room existed")
		}
	}
	if jsonStr, err := room.Json(); err != nil {
		return nil, err
	} else {
		err = r.cache.SetEx(roomCacheKey, jsonStr, time.Hour*24)
		if err != nil {
			return nil, err
		}
		requestTimeKey := r.getParticipantRequestRoomTimeKey(room.Id, req.UId)
		err = r.cache.SetEx(requestTimeKey, time.Now().UnixMilli(), time.Minute*5)
		return room, err
	}
}

func (r *ServiceImpl) JoinRoom(req *dto.RoomJoinReq) (*Room, error) {
	room, err := r.FindRoomById(req.RoomId)
	if err != nil {
		return nil, err
	}
	requestTimeKey := r.getParticipantRequestRoomTimeKey(room.Id, req.UId)
	err = r.cache.SetEx(requestTimeKey, time.Now().UnixMilli(), time.Minute*5)
	return room, err
}

func (r *ServiceImpl) FindRoomById(id string) (*Room, error) {
	value, err := r.cache.Get(r.getRoomCacheKey(id))
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	roomJson, ok := value.(string)
	if !ok {
		return nil, errors.New("InternalServerError")
	}
	room, e := NewRoomByJson([]byte(roomJson))
	if e != nil {
		return nil, e
	}
	if members, errMembers := r.cache.HValues(r.getParticipantsCacheKey(id)); errMembers == nil {
		participants := make([]*Participant, 0)
		for _, m := range members {
			if participant, errJson := NewParticipantByJson([]byte(m)); errJson == nil {
				participants = append(participants, participant)
			}
		}
		room.Participants = participants
	}
	return room, nil
}

func (r *ServiceImpl) DestroyRoom(id string) error {
	if err := r.cache.Del(r.getRoomCacheKey(id)); err != nil {
		return err
	}
	if err := r.cache.Del(r.getParticipantsCacheKey(id)); err != nil {
		return err
	}

	// 发送停止房间Stream的消息
	event := &DestroyRoomEvent{
		RoomId: id,
	}
	if js, e := json.Marshal(event); e == nil {
		if e = r.cache.Pub(DestroyRoomEventKey, string(js)); e != nil {
			r.logger.Error("DataStreamStop: ", id, e)
		}
	} else {
		r.logger.Error("DataStreamStop: ", id, e)
	}
	return nil
}

func (r *ServiceImpl) getParticipantRequestRoomTimeKey(roomId string, userId int64) string {
	return fmt.Sprintf(ParticipantRequestRoomTime, roomId, userId)
}

func (r *ServiceImpl) getRoomCacheKey(roomId string) string {
	return fmt.Sprintf(CacheKey, roomId)
}

func (r *ServiceImpl) getParticipantsCacheKey(roomId string) string {
	return fmt.Sprintf(ParticipantsKey, roomId)
}

func (r *ServiceImpl) OnParticipantJoin(roomId, streamKey string, joinTime int64, role int, uId int64) error {
	participant := &Participant{
		UId:       uId,
		Role:      role,
		JoinTime:  joinTime,
		StreamKey: &streamKey,
	}
	pJson, err := participant.Json()
	if err == nil {
		roomCacheKey := r.getRoomCacheKey(roomId)
		room, errRoom := r.FindRoomById(roomId)
		if errRoom != nil {
			return errRoom
		}
		if room.Mode == ModeVoiceRoom {
			err = r.cache.Expire(roomCacheKey, time.Hour*24*365)
		} else {
			err = r.cache.Expire(roomCacheKey, time.Hour*24)
		}
		if err != nil {
			return err
		}
		cacheKey := r.getParticipantsCacheKey(roomId)
		err = r.cache.HSet(cacheKey, streamKey, pJson, time.Hour*24*30)
	}
	return err
}

func (r *ServiceImpl) GetRequestJoinRoomTime(roomId string, uId int64) (int64, error) {
	requestTimeKey := r.getParticipantRequestRoomTimeKey(roomId, uId)
	v, err := r.cache.Get(requestTimeKey)
	if err != nil {
		return 0, err
	}
	r.logger.Infof("GetRequestJoinRoomTime %s, %d, %v", roomId, uId, v)
	t, ok := v.(string)
	if ok {
		it, _ := strconv.Atoi(t)
		return int64(it), nil
	} else {
		return 0, nil
	}
}

func (r *ServiceImpl) OnParticipantLeave(roomId, streamKey string, uId int64) error {
	r.logger.Info("OnParticipantLeave", roomId, uId, streamKey)
	cacheKey := r.getParticipantsCacheKey(roomId)
	if err := r.cache.HDel(cacheKey, streamKey); err != nil {
		return err
	}
	if ps, err := r.cache.HValues(cacheKey); err != nil {
		r.logger.Error("OnParticipantLeave", roomId, uId, streamKey, err)
		return err
	} else {
		r.logger.Info("OnParticipantLeave", roomId, uId, ps)
		if ps == nil || len(ps) == 0 {
			return r.DestroyRoom(roomId)
		} else {
			return nil
		}
	}
}
