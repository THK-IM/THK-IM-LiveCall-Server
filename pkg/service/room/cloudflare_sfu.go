package room

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseErrorx "github.com/thk-im/thk-im-base-server/errorx"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/errorx"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/signal"
)

const RoomsKey = "live_server:rooms:"

func NewCloudflareSFURoomService(appCtx *app.Context) Service {
	return &CloudflareSFURoomService{
		baseRoomService: baseRoomService{
			appCtx:        appCtx,
			signalService: signal.NewSignalService(appCtx),
		},
	}
}

type CloudflareSFURoomService struct {
	baseRoomService
}

//func (w CloudflareSFURoomService) api() sdk.SfuApi {
//	if w.appCtx.Context.SdkMap["cloudflare_connect_api"] == nil {
//		return nil
//	}
//	return w.appCtx.Context.SdkMap["cloudflare_connect_api"].(sdk.SfuApi)
//}

func (w CloudflareSFURoomService) CreateRoom(req *dto.RoomCreateReq, claims baseDto.ThkClaims) (*dto.RoomJoinResp, error) {
	lockerKey := fmt.Sprintf(SessionLockerKey, req.SessionId)
	locker := w.appCtx.NewLocker(lockerKey, 3000, 3000)
	success, errLock := locker.Lock()
	if errLock != nil {
		return nil, errLock
	}
	if !success {
		return nil, baseErrorx.ErrInternalServerError
	}
	defer func() {
		_, _ = locker.Release()
	}()

	sessionCacheKey := w.getSessionCacheKey(req.SessionId)
	roomId, errExist := w.appCtx.RedisCache().Get(context.Background(), sessionCacheKey).Result()
	if errExist != nil && !errors.Is(errExist, redis.Nil) {
		return nil, errExist
	}
	resp := &dto.RoomJoinResp{}
	if roomId != "" {
		room, errRoom := w.FindRoomById(roomId, claims)
		if errRoom != nil && !errors.Is(errRoom, redis.Nil) {
			return nil, errRoom
		}
		if room != nil {
			for _, p := range room.Participants {
				if p.UId == req.UId && p.StreamKey != "" {
					err := w.DestroyRoom(room.Id, claims)
					if err != nil {
						return nil, err
					}
				}
			}
			resp.Room = room
		}
	}

	if resp.Room == nil {
		id := w.appCtx.SnowflakeNode().Generate().Base36()
		room, errCreateRoom := w.baseRoomService.CreateRoom(id, "WebRTC", req, claims)
		if errCreateRoom != nil {
			return nil, errCreateRoom
		}
		if room == nil {
			return nil, baseErrorx.ErrInternalServerError
		}
		resp.Room = room
	}
	_, err := w.appCtx.RedisCache().SAdd(context.Background(), RoomsKey, []string{resp.Room.Id}).Result()
	if err != nil {
		return nil, err
	}
	err = w.AddRoomMember(resp.Room.Id, req.UId, claims)
	return resp, err
}

func (w CloudflareSFURoomService) RequestJoinRoom(req *dto.RoomJoinReq, claims baseDto.ThkClaims) (*dto.RoomJoinResp, error) {
	room, errRoom := w.baseRoomService.FindRoomById(req.RoomId, claims)
	if errRoom != nil {
		return nil, errRoom
	}
	if room == nil {
		return nil, errorx.ErrRoomNotExisted
	}
	return &dto.RoomJoinResp{
		Room:  room,
		Token: "",
	}, nil
}

func (w CloudflareSFURoomService) CheckRooms() error {
	claims := baseDto.ThkClaims{}
	claims.PutValue(baseDto.TraceID, fmt.Sprintf("CheckRooms-%d", time.Now().UnixNano()))
	claims.PutValue(baseDto.SpanID, "1")
	for {
		id, err := w.appCtx.RedisCache().SPop(context.Background(), RoomsKey).Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil
			}
		}
		err = w.checkRoom(id, claims)
		if err != nil {
			return err
		}
		time.Sleep(5 * time.Second)
	}
}

func (w CloudflareSFURoomService) checkRoom(id string, claims baseDto.ThkClaims) error {
	//room, err := w.FindRoomById(id, claims)
	//if err != nil {
	//	return err
	//}
	//if room == nil {
	//	return nil
	//}
	//
	//needDestroy := true
	//for _, p := range room.Participants {
	//	if p.StreamKey != "" {
	//		statusResp, errStatus := w.api().GetSessionState(p.StreamKey)
	//		if errStatus != nil {
	//			w.appCtx.Logger().Errorf("checkRoom %s %v", id, errStatus)
	//			needDestroy = false
	//			break
	//		}
	//		if len(statusResp.Tracks) > 0 || len(statusResp.DataChannels) > 0 {
	//			// 还在通话，放回房间
	//			needDestroy = false
	//			break
	//		}
	//	}
	//}
	//timeout := time.Now().UnixMilli() - room.CreateTime
	//
	//if needDestroy && (timeout > 60*1000) {
	//	err = w.DestroyRoom(id, claims)
	//	if err != nil {
	//		w.appCtx.Logger().Errorf("checkRoom %s %v", id, err)
	//		return err
	//	}
	//} else {
	//	_, err = w.appCtx.RedisCache().SAdd(context.Background(), RoomsKey, []string{id}).Result()
	//	if err != nil {
	//		w.appCtx.Logger().Errorf("checkRoom %s %v", id, err)
	//		return err
	//	}
	//}
	return nil
}
