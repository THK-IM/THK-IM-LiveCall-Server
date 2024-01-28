package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseMiddleware "github.com/thk-im/thk-im-base-server/middleware"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	msgDto "github.com/thk-im/thk-im-msgapi-server/pkg/dto"
	userSdk "github.com/thk-im/thk-im-user-server/pkg/sdk"
)

const (
	InviteLiveCall = 400
	CancelLiveCall = 401
)

func createRoom(appCtx *app.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.RoomCreateReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("createRoom %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(userSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("createRoom %d %d", requestUid, req.UId)
			baseDto.ResponseForbidden(ctx)
			return
		}
		if rsp, err := appCtx.RoomService().CreateRoom(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("createRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("createRoom %d %d", req.UId, req.Mode)
			if len(req.Members) > 0 {
				notify := &dto.LiveCallNotify{
					RoomId:     rsp.Id,
					Mode:       rsp.Mode,
					OwnerId:    rsp.OwnerId,
					CreateTime: rsp.CreateTime,
					Members:    req.Members,
				}
				pushMessage := &msgDto.PushMessageReq{
					UIds:        req.Members,
					Type:        InviteLiveCall,
					Body:        notify.JsonString(),
					OfflinePush: true,
				}
				if resp, errPush := appCtx.MsgApi().PushMessage(pushMessage, claims); err != nil {
					appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("createRoom %v %s", pushMessage, errPush.Error())
					baseDto.ResponseInternalServerError(ctx, errPush)
					return
				} else {
					appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("createRoom %v %s", pushMessage, resp)
				}
			}
			baseDto.ResponseSuccess(ctx, rsp)
		}
	}
}

func deleteRoom(appCtx *app.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.RoomDelReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("deleteRoom %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(userSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("deleteRoom %d %d", requestUid, req.UId)
			baseDto.ResponseForbidden(ctx)
			return
		}
		room, errRoom := appCtx.RoomService().FindRoomById(req.RoomId)
		if errRoom != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("deleteRoom %v %s", req.RoomId, errRoom.Error())
			baseDto.ResponseInternalServerError(ctx, errRoom)
			return
		}
		if room == nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("deleteRoom %v %s", req.RoomId, "room not existed")
			baseDto.ResponseSuccess(ctx, nil)
			return
		}

		if len(room.Participants) > 0 {
			uIds := make([]int64, 0)
			for _, p := range room.Participants {
				uIds = append(uIds, p.UId)
			}

			notify := &dto.LiveCallNotify{
				RoomId:     room.Id,
				OwnerId:    room.OwnerId,
				Mode:       room.Mode,
				CreateTime: room.CreateTime,
				Members:    uIds,
			}
			pushMessage := &msgDto.PushMessageReq{
				UIds:        uIds,
				Type:        CancelLiveCall,
				Body:        notify.JsonString(),
				OfflinePush: true,
			}
			if _, err := appCtx.MsgApi().PushMessage(pushMessage, claims); err != nil {
				appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("deleteRoom %v %s", req.RoomId, err.Error())
				baseDto.ResponseInternalServerError(ctx, err)
				return
			}
		}

		if err := appCtx.RoomService().DestroyRoom(req.RoomId); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("deleteRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("deleteRoom %d", req.UId)
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}

func joinRoom(appCtx *app.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.RoomJoinReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("joinRoom %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(userSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("joinRoom %d %d", requestUid, req.UId)
			baseDto.ResponseForbidden(ctx)
			return
		}
		if rsp, err := appCtx.RoomService().JoinRoom(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("joinRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("joinRoom %d %s %d", req.UId, req.RoomId, req.Role)
			baseDto.ResponseSuccess(ctx, rsp)
		}
	}
}

func findRoomById(appCtx *app.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		roomId := ctx.Param("id")
		if len(roomId) == 0 {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("findRoomById %s", roomId)
			baseDto.ResponseBadRequest(ctx)
		}
		if rsp, err := appCtx.RoomService().FindRoomById(roomId); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("findRoomById %v %s", roomId, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("findRoomById %s %v", roomId, rsp)
			baseDto.ResponseSuccess(ctx, rsp)
		}
	}
}
