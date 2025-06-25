package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseMiddleware "github.com/thk-im/thk-im-base-server/middleware"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/logic"
	msgSdk "github.com/thk-im/thk-im-msgapi-server/pkg/sdk"
)

func createRoom(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
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

		if resp, err := l.CreateRoom(req, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("createRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("createRoom %v %v", req, resp)
			baseDto.ResponseSuccess(ctx, resp)
		}
	}
}

func callRoomMembers(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
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

		if err := l.CallRoomMembers(req, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("callRoomMembers %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("callRoomMembers %v", req)
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}

func cancelCallRoomMembers(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
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

		if err := l.CancelCallRoomMembers(req, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("cancelCallRoomMembers %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("cancelCallRoomMembers %v", req)
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}

func deleteRoom(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
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

		if err := l.DeleteRoom(req, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("deleteRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("deleteRoom %v", req)
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}

func joinRoom(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
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

		if resp, err := l.JoinRoom(req, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("joinRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("joinRoom %v %v", req, resp)
			baseDto.ResponseSuccess(ctx, resp)
		}
	}
}

func refuseJoinRoom(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
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

		if err := l.RefuseJoinRoom(req, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("refuseJoinRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("refuseJoinRoom %v", req)
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}

func inviteJoinRoom(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
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

		if err := l.InviteJoinRoom(req, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("inviteJoinRoom %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("inviteJoinRoom %v", req)
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}

func leaveRoomMember(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
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

		if err := l.RoomMemberLeave(req, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("leaveRoomMember %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("leaveRoomMember %v", req)
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}

func KickoffRoomMember(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.KickoffMemberReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("KickoffRoomMember %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(msgSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("KickoffRoomMember %v", req)
			baseDto.ResponseForbidden(ctx)
			return
		}

		if err := l.KickoffRoomMember(req, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("KickoffRoomMember %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("KickoffRoomMember %v", req)
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}

func findRoomById(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		roomId := ctx.Param("id")
		if len(roomId) == 0 {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("findRoomById %s", roomId)
			baseDto.ResponseBadRequest(ctx)
		}
		if rsp, err := l.QueryRoom(roomId, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("findRoomById %v %s", roomId, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("findRoomById %s %v", roomId, rsp)
			baseDto.ResponseSuccess(ctx, rsp)
		}
	}
}
