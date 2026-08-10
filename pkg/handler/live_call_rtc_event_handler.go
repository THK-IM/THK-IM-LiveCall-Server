package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseMiddleware "github.com/thk-im/thk-im-base-server/middleware"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/logic"
	rtcDto "github.com/thk-im/thk-im-rtc-server/pkg/dto"
)

func rtcUserJoinEvent(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &rtcDto.UserJoinEvent{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("rtcUserJoinEvent %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		userId, errId := strconv.ParseInt(req.UserId, 10, 64)
		if errId != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("rtcUserJoinEvent %s %v", req.UserId, errId.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("rtcUserJoinEvent req: %v", req)
		event := &dto.RoomUserJoinEvent{
			UserId:    userId,
			RoomId:    req.ChannelId,
			Timestamp: req.Timestamp,
		}
		if err := l.OnUserJoinEvent(event, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("rtcUserJoinEvent %s", err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("rtcUserJoinEvent success %v", req)
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}

func rtcUserLeaveEvent(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &rtcDto.UserLeaveEvent{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("rtcUserLeaveEvent %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		userId, errId := strconv.ParseInt(req.UserId, 10, 64)
		if errId != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("rtcUserLeaveEvent %s %v", req.UserId, errId.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}

		appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("rtcUserLeaveEvent req: %v", req)

		event := &dto.RoomUserLevelEvent{
			UserId:    userId,
			RoomId:    req.ChannelId,
			Timestamp: req.Timestamp,
		}
		if err := l.OnUserLeaveEvent(event, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("rtcUserLeaveEvent %s", err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("rtcUserLeaveEvent success %v", req)
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}

func rtcUserPushEvent(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewRoomLogic(appCtx)
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &rtcDto.UserPushEvent{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("rtcUserPushEvent %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		userId, errId := strconv.ParseInt(req.UserId, 10, 64)
		if errId != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("rtcUserPushEvent %s %v", req.UserId, errId.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("rtcUserPushEvent req: %v", req)
		event := &dto.RoomUserPushStreamEvent{
			UserId:    userId,
			RoomId:    req.ChannelId,
			Timestamp: req.Timestamp,
		}
		if err := l.OnUserPushEvent(event, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("rtcUserPushEvent %s", err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("rtcUserPushEvent success %v", req)
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}
