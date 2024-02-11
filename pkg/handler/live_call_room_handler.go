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
	PushMessageTypeLiveCall = 400
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
		if len(req.Members) < 0 {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("createRoom %v", req)
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
			notify := &dto.LiveCallSignal{
				RoomId:     rsp.Id,
				Mode:       rsp.Mode,
				OwnerId:    rsp.OwnerId,
				CreateTime: rsp.CreateTime,
				Members:    req.Members,
				MsgType:    dto.InviteLiveCall,
				OperatorId: req.UId,
			}
			pushMessage := &msgDto.PushMessageReq{
				UIds:        req.Members,
				Type:        PushMessageTypeLiveCall,
				Body:        notify.JsonString(),
				OfflinePush: true,
			}
			if resp, errPush := appCtx.MsgApi().PushMessage(pushMessage, claims); err != nil {
				appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("createRoom %v %s", pushMessage, errPush.Error())
				baseDto.ResponseInternalServerError(ctx, errPush)
			} else {
				appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("createRoom %v %v", pushMessage, resp)
				baseDto.ResponseSuccess(ctx, rsp)
			}
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

		if len(room.Members) > 0 {
			notify := &dto.LiveCallSignal{
				RoomId:     room.Id,
				OwnerId:    room.OwnerId,
				Mode:       room.Mode,
				CreateTime: room.CreateTime,
				Members:    room.Members,
				MsgType:    dto.EndLiveCall,
				OperatorId: req.UId,
			}
			pushMessage := &msgDto.PushMessageReq{
				UIds:        room.Members,
				Type:        PushMessageTypeLiveCall,
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

func hangup(appCtx *app.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.RefuseJoinRoomReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("hangup %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(userSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("hangup %d %v", requestUid, req)
			baseDto.ResponseForbidden(ctx)
			return
		}

		room, err := appCtx.RoomService().FindRoomById(req.RoomId)
		if err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("hangup %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			members := room.Members
			newMembers := make([]int64, 0)
			for _, m := range members {
				if m != req.UId {
					newMembers = append(newMembers, m)
				}
			}
			notify := &dto.LiveCallSignal{
				RoomId:     room.Id,
				Mode:       room.Mode,
				OwnerId:    room.OwnerId,
				CreateTime: room.CreateTime,
				Members:    newMembers,
				MsgType:    dto.HangupLiveCall,
				OperatorId: req.UId,
			}
			pushMessage := &msgDto.PushMessageReq{
				UIds:        newMembers,
				Type:        PushMessageTypeLiveCall,
				Body:        notify.JsonString(),
				OfflinePush: true,
			}
			if resp, errPush := appCtx.MsgApi().PushMessage(pushMessage, claims); err != nil {
				appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("hangup %v %s", pushMessage, errPush.Error())
				baseDto.ResponseInternalServerError(ctx, errPush)
			} else {
				appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("hangup %v %v", pushMessage, resp)
				baseDto.ResponseSuccess(ctx, nil)
			}
		}
	}
}

func inviteJoinRoom(appCtx *app.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.InviteJoinRoomReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("inviteJoinRoom %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(userSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("inviteJoinRoom %v", req)
			baseDto.ResponseForbidden(ctx)
			return
		}

		room, err := appCtx.RoomService().FindRoomById(req.RoomId)
		if err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("inviteJoinRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			hasPermission := false
			for _, p := range room.Members {
				if p == req.UId {
					hasPermission = true
					break
				}
			}
			if !hasPermission {
				appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("inviteJoinRoom %v", req)
				baseDto.ResponseForbidden(ctx)
			}
			err = appCtx.RoomService().AddRoomMember(room, req.InviteUIds)
			if err != nil {
				baseDto.ResponseInternalServerError(ctx, err)
			}
			notify := &dto.LiveCallSignal{
				RoomId:     room.Id,
				Mode:       room.Mode,
				OwnerId:    room.OwnerId,
				CreateTime: room.CreateTime,
				Members:    room.Members,
				MsgType:    dto.InviteLiveCall,
				OperatorId: req.UId,
			}
			pushMessage := &msgDto.PushMessageReq{
				UIds:        room.Members,
				Type:        PushMessageTypeLiveCall,
				Body:        notify.JsonString(),
				OfflinePush: true,
			}
			if resp, errPush := appCtx.MsgApi().PushMessage(pushMessage, claims); err != nil {
				appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("inviteJoinRoom %v %s", pushMessage, errPush.Error())
				baseDto.ResponseInternalServerError(ctx, errPush)
			} else {
				appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("inviteJoinRoom %v %v", pushMessage, resp)
				baseDto.ResponseSuccess(ctx, nil)
			}
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
