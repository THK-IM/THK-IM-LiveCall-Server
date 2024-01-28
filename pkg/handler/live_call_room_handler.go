package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseMiddleware "github.com/thk-im/thk-im-base-server/middleware"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	userSdk "github.com/thk-im/thk-im-user-server/pkg/sdk"
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
			baseDto.ResponseSuccess(ctx, rsp)
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
