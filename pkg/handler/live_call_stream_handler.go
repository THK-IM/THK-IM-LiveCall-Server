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

func publishStream(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewWebRTCStreamLogic(appCtx)
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.PublishStreamReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("publishStream %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(msgSdk.UidKey)
		if requestUid <= 0 {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("publishStream %d", requestUid)
			baseDto.ResponseForbidden(ctx)
			return
		}
		req.Uid = requestUid

		if resp, err := l.PublishStream(req, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("publishStream %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("publishStream %v %v", req, resp)
			baseDto.ResponseSuccess(ctx, resp)
		}
	}
}

func subscribeStream(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewWebRTCStreamLogic(appCtx)
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.SubscribeStreamReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("subscribeStream %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(msgSdk.UidKey)
		if requestUid <= 0 {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("subscribeStream %d", requestUid)
			baseDto.ResponseForbidden(ctx)
			return
		}
		req.Uid = requestUid

		if resp, err := l.SubscribeStream(req, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("subscribeStream %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("subscribeStream %v %v", req, resp)
			baseDto.ResponseSuccess(ctx, resp)
		}
	}
}

func updateStreamStatus(appCtx *app.Context) gin.HandlerFunc {
	l := logic.NewWebRTCStreamLogic(appCtx)
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.StreamStatusUpdateReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("updateStreamStatus %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(msgSdk.UidKey)
		if requestUid <= 0 {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("subscribeStream %d", requestUid)
			baseDto.ResponseForbidden(ctx)
			return
		}
		req.Uid = requestUid

		if err := l.UpdateStreamStatus(req, claims); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("updateStreamStatus %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Tracef("updateStreamStatus %v %v", req, "success")
			baseDto.ResponseSuccess(ctx, nil)
		}
	}
}
