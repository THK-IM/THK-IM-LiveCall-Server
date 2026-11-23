package room

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/errorx"
)

// IssuedKey is private to Redis. Only a ready stream is exposed in the room snapshot.
type memberState struct {
	dto.Participant
	IssuedKey string `json:"issued_key,omitempty"`
}

func (m *memberState) streamChanged(key string, start bool) bool {
	if key == "" || m.LeaveTime != 0 || m.Refuse != 0 {
		return false
	}
	if start {
		if m.IssuedKey != key || m.StreamKey == key {
			return false
		}
		m.StreamKey = key
		m.Role = dto.Broadcast
		return true
	}
	if m.IssuedKey != key && m.StreamKey != key {
		return false
	}
	if m.IssuedKey == key {
		m.IssuedKey = ""
	}
	if m.StreamKey == key {
		m.StreamKey = ""
		m.Role = dto.Audience
	}
	return true
}

var compareMember = redis.NewScript(`
if redis.call('HGET', KEYS[1], ARGV[1]) ~= ARGV[2] then return 0 end
redis.call('HSET', KEYS[1], ARGV[1], ARGV[3])
return 1`)

func (r baseRoomService) updateMember(roomID string, uid int64, update func(*memberState) bool) (bool, error) {
	ctx := context.Background()
	cache := r.appCtx.RedisCache()
	key, field := r.getParticipantsCacheKey(roomID), fmt.Sprint(uid)
	for {
		before, err := cache.HGet(ctx, key, field).Result()
		if errors.Is(err, redis.Nil) {
			return false, errorx.ErrNoPermission
		}
		if err != nil {
			return false, err
		}
		var member memberState
		if err := json.Unmarshal([]byte(before), &member); err != nil {
			return false, err
		}
		if !update(&member) {
			return false, nil
		}
		after, err := json.Marshal(member)
		if err != nil {
			return false, err
		}
		ok, err := compareMember.Run(ctx, cache, []string{key}, field, before, after).Int()
		if err != nil || ok == 1 {
			return ok == 1, err
		}
	}
}

func (r baseRoomService) PreparePublish(roomID string, uid int64, key string) error {
	ok, err := r.updateMember(roomID, uid, func(m *memberState) bool {
		if key == "" || m.LeaveTime != 0 || m.Refuse != 0 {
			return false
		}
		m.IssuedKey = key
		if m.JoinTime == 0 {
			m.JoinTime = time.Now().UnixMilli()
		}
		return true
	})
	if err == nil && !ok {
		return errorx.ErrNoPermission
	}
	return err
}

func (r baseRoomService) streamEvent(event *dto.RoomUserPushStreamEvent, start bool, claims baseDto.ThkClaims) error {
	room, err := r.FindRoomById(event.RoomId, claims)
	if err != nil || room == nil {
		return err
	}
	changed, err := r.updateMember(room.Id, event.UserId, func(m *memberState) bool {
		return m.streamChanged(event.StreamKey, start)
	})
	if err != nil || !changed {
		return err
	}
	var ids []int64
	for _, p := range room.Participants {
		if p.UId != event.UserId && p.LeaveTime == 0 && p.Refuse == 0 {
			ids = append(ids, p.UId)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	makeSignal := dto.MakeParticipantPushStreamSignal
	if !start {
		makeSignal = dto.MakeParticipantLeaveSignal
	}
	return r.signalService.PushSignal(makeSignal(room.Id, event.StreamKey, event.UserId, event.Timestamp), ids, claims)
}

func (r baseRoomService) OnUserStopEvent(event *dto.RoomUserPushStreamEvent, claims baseDto.ThkClaims) error {
	return r.streamEvent(event, false, claims)
}

var refreshVoiceRoom = redis.NewScript(`
if redis.call('EXISTS', KEYS[1]) == 0 then return 0 end
redis.call('EXPIRE', KEYS[1], ARGV[2])
redis.call('EXPIRE', KEYS[2], ARGV[2])
if redis.call('GET', KEYS[3]) == ARGV[1] then redis.call('EXPIRE', KEYS[3], ARGV[2]) end
return 1`)

func (r baseRoomService) KeepVoiceRoom(room *dto.Room) error {
	if room.Mode != 4 || room.SessionId == nil {
		return nil
	}
	keys := []string{r.getRoomCacheKey(room.Id), r.getParticipantsCacheKey(room.Id), r.getSessionCacheKey(*room.SessionId)}
	return refreshVoiceRoom.Run(context.Background(), r.appCtx.RedisCache(), keys, room.Id, int(time.Hour.Seconds())).Err()
}
