package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseMiddleware "github.com/thk-im/thk-im-base-server/middleware"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	msgDto "github.com/thk-im/thk-im-msgapi-server/pkg/dto"
	msgSdk "github.com/thk-im/thk-im-msgapi-server/pkg/sdk"
	"time"
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
		requestUid := ctx.GetInt64(msgSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("createRoom %d %d", requestUid, req.UId)
			baseDto.ResponseForbidden(ctx)
			return
		}
		if rsp, err := appCtx.RoomService().CreateRoom(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("createRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("createRoom %v %v", "success", rsp)
			baseDto.ResponseSuccess(ctx, rsp)
		}
	}
}

func callRoomMembers(appCtx *app.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.RoomCallReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("callRoomMembers %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(msgSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("callRoomMembers %d %d", requestUid, req.UId)
			baseDto.ResponseForbidden(ctx)
			return
		}
		room, errRoom := appCtx.RoomService().FindRoomById(req.RoomId)
		if errRoom != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("callRoomMembers %v %s", req.RoomId, errRoom.Error())
			baseDto.ResponseInternalServerError(ctx, errRoom)
			return
		}
		if room == nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("callRoomMembers %v %s", req.RoomId, "room not existed")
			baseDto.ResponseSuccess(ctx, nil)
			return
		}
		if room.OwnerId != requestUid {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("callRoomMembers %d %d", room.OwnerId, requestUid)
			baseDto.ResponseForbidden(ctx)
		}

		if len(req.Members) == 0 {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("callRoomMembers room.Members 0")
			baseDto.ResponseBadRequest(ctx)
		}

		signal := dto.MakeBeingRequestedSignal(
			room.Id, req.Members, room.Mode, req.Msg, req.UId, room.CreateTime,
			time.Now().UnixMilli()+req.Duration*1000,
		)
		pushMessage := &msgDto.PushMessageReq{
			UIds:        req.Members,
			Type:        PushMessageTypeLiveCall,
			Body:        signal.JsonString(),
			OfflinePush: true,
		}
		if _, err := appCtx.MsgApi().PushMessage(pushMessage, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("callRoomMembers %v %s", req.RoomId, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
			return
		}
	}
}

func cancelCallRoomMembers(appCtx *app.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.CancelCallingReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("cancelCallRoomMembers %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(msgSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("cancelCallRoomMembers %d %d", requestUid, req.UId)
			baseDto.ResponseForbidden(ctx)
			return
		}
		room, errRoom := appCtx.RoomService().FindRoomById(req.RoomId)
		if errRoom != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("cancelCallRoomMembers %v %s", req.RoomId, errRoom.Error())
			baseDto.ResponseInternalServerError(ctx, errRoom)
			return
		}
		if room == nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("cancelCallRoomMembers %v %s", req.RoomId, "room not existed")
			baseDto.ResponseSuccess(ctx, nil)
			return
		}
		if room.OwnerId != requestUid {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("cancelCallRoomMembers %d %d", room.OwnerId, requestUid)
			baseDto.ResponseForbidden(ctx)
		}

		if len(req.Members) == 0 {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("cancelCallRoomMembers room.Members 0")
			baseDto.ResponseBadRequest(ctx)
		}

		signal := dto.MakeCancelRequestingSignal(
			room.Id, req.Msg, room.CreateTime, time.Now().UnixMilli(),
		)
		pushMessage := &msgDto.PushMessageReq{
			UIds:        req.Members,
			Type:        PushMessageTypeLiveCall,
			Body:        signal.JsonString(),
			OfflinePush: true,
		}
		if _, err := appCtx.MsgApi().PushMessage(pushMessage, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("cancelCallRoomMembers %v %s", req.RoomId, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
			return
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
		requestUid := ctx.GetInt64(msgSdk.UidKey)
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
			members := make([]int64, 0)
			for _, p := range room.Participants {
				members = append(members, p.UId)
			}
			signal := dto.MakeEndCallSignal(
				room.Id, "", req.UId, time.Now().UnixMilli(),
			)
			pushMessage := &msgDto.PushMessageReq{
				UIds:        members,
				Type:        PushMessageTypeLiveCall,
				Body:        signal.JsonString(),
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
		requestUid := ctx.GetInt64(msgSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("joinRoom %d %d", requestUid, req.UId)
			baseDto.ResponseForbidden(ctx)
			return
		}
		if room, err := appCtx.RoomService().RequestJoinRoom(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("joinRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			if room == nil {
				appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("joinRoom %d %s not found", req.UId, req.RoomId)
				baseDto.ResponseInternalServerError(ctx, errors.New("room not found"))
			} else {
				signal := dto.MakeAcceptRequestSignal(
					room.Id, "", req.UId, time.Now().UnixMilli(),
				)
				pushMessage := &msgDto.PushMessageReq{
					UIds:        []int64{room.OwnerId},
					Type:        PushMessageTypeLiveCall,
					Body:        signal.JsonString(),
					OfflinePush: true,
				}
				_, _ = appCtx.MsgApi().PushMessage(pushMessage, claims)
				appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("joinRoom %d %s %d", req.UId, req.RoomId, req.Role)
				baseDto.ResponseSuccess(ctx, room)
			}
		}
	}
}

func refuseJoinRoom(appCtx *app.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.RefuseJoinRoomReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("refuseJoinRoom %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(msgSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("refuseJoinRoom %d %v", requestUid, req)
			baseDto.ResponseForbidden(ctx)
			return
		}

		room, err := appCtx.RoomService().FindRoomById(req.RoomId)
		if err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("refuseJoinRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
			return
		}
		if room == nil {
			baseDto.ResponseSuccess(ctx, nil)
			return
		}
		members := make([]int64, 0)
		for _, p := range room.Participants {
			if p.UId != req.UId {
				members = append(members, p.UId)
			}
		}
		signal := dto.MakeRejectRequestSignal(
			room.Id, req.Msg, req.UId, time.Now().UnixMilli(),
		)
		pushMessage := &msgDto.PushMessageReq{
			UIds:        members,
			Type:        PushMessageTypeLiveCall,
			Body:        signal.JsonString(),
			OfflinePush: true,
		}
		if resp, errPush := appCtx.MsgApi().PushMessage(pushMessage, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("refuseJoinRoom %v %s", pushMessage, errPush.Error())
			baseDto.ResponseInternalServerError(ctx, errPush)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("refuseJoinRoom %v %v", pushMessage, resp)
			baseDto.ResponseSuccess(ctx, nil)
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
		requestUid := ctx.GetInt64(msgSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("inviteJoinRoom %v", req)
			baseDto.ResponseForbidden(ctx)
			return
		}

		room, err := appCtx.RoomService().FindRoomById(req.RoomId)
		if err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("inviteJoinRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
			return
		}
		if room == nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("joinRoom %d %s not found", req.UId, req.RoomId)
			baseDto.ResponseInternalServerError(ctx, errors.New("room not found"))
			return
		}
		signal := dto.MakeBeingRequestedSignal(
			room.Id, req.InviteUIds, room.Mode, req.Msg, req.UId, room.CreateTime,
			time.Now().UnixMilli()+req.Duration*1000,
		)
		pushMessage := &msgDto.PushMessageReq{
			UIds:        req.InviteUIds,
			Type:        PushMessageTypeLiveCall,
			Body:        signal.JsonString(),
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

func leaveRoomMember(appCtx *app.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.RoomMemberLeaveReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("leaveRoomMember %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(msgSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("leaveRoomMember %v", req)
			baseDto.ResponseForbidden(ctx)
			return
		}

		room, err := appCtx.RoomService().FindRoomById(req.RoomId)
		if err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("leaveRoomMember %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
			return
		}
		if room != nil {
			members := make([]int64, 0)
			for _, p := range room.Participants {
				members = append(members, p.UId)
			}
			if len(members) > 0 {
				signal := dto.MakeHangupSignal(
					room.Id, req.Msg, req.UId, time.Now().UnixMilli(),
				)
				pushMessage := &msgDto.PushMessageReq{
					UIds:        members,
					Type:        PushMessageTypeLiveCall,
					Body:        signal.JsonString(),
					OfflinePush: true,
				}
				if resp, errPush := appCtx.MsgApi().PushMessage(pushMessage, claims); err != nil {
					appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("leaveRoomMember %v %s", pushMessage, errPush.Error())
					baseDto.ResponseInternalServerError(ctx, errPush)
				} else {
					appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("leaveRoomMember %v %v", pushMessage, resp)
					baseDto.ResponseSuccess(ctx, nil)
				}
			}
		} else {
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}

func kickRoomMember(appCtx *app.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.KickoffMemberReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("kickRoomMember %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(msgSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("kickRoomMember %v", req)
			baseDto.ResponseForbidden(ctx)
			return
		}

		room, err := appCtx.RoomService().FindRoomById(req.RoomId)
		if err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("kickRoomMember %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
			return
		}
		if room == nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("kickRoomMember %d %s not found", req.UId, req.RoomId)
			baseDto.ResponseInternalServerError(ctx, errors.New("room not found"))
			return
		}
		hasPermission := room.OwnerId == req.UId
		if !hasPermission {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("kickRoomMember %v", req)
			baseDto.ResponseForbidden(ctx)
			return
		}
		members := make([]int64, 0)
		for _, p := range room.Participants {
			members = append(members, p.UId)
		}
		signal := dto.MakeKickMemberSignal(
			room.Id, req.Msg, req.UId, time.Now().UnixMilli(), req.KickoffUIds,
		)
		pushMessage := &msgDto.PushMessageReq{
			UIds:        members,
			Type:        PushMessageTypeLiveCall,
			Body:        signal.JsonString(),
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
