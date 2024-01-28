package handler

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	baseMiddleware "github.com/thk-im/thk-im-base-server/middleware"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/rtc"
	userSdk "github.com/thk-im/thk-im-user-server/pkg/sdk"
)

func publishStream(appCtx *app.Context, rtcService rtc.Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.PublishReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("publish %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(userSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("publish %d %d", requestUid, req.UId)
			baseDto.ResponseForbidden(ctx)
			return
		}
		offer, errOffer := base64.StdEncoding.DecodeString(req.OfferSdp)
		if errOffer != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("publish %d %s", req.UId, req.OfferSdp)
			baseDto.ResponseForbidden(ctx)
			return
		}
		if conn, err := rtcService.RequestPublish(req.RoomId, string(offer), req.UId); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("publish %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			answer := base64.StdEncoding.EncodeToString([]byte(conn.ServerSdp().SDP))
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("publish %v %v", req, answer)
			baseDto.ResponseSuccess(ctx, &dto.PublishResp{AnswerSdp: answer, StreamKey: conn.Key()})
		}
	}
}

func playStream(appCtx *app.Context, rtcService rtc.Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims := ctx.MustGet(baseMiddleware.ClaimsKey).(baseDto.ThkClaims)
		req := &dto.PlayReq{}
		if err := ctx.BindJSON(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("playStream %s", err.Error())
			baseDto.ResponseBadRequest(ctx)
			return
		}
		requestUid := ctx.GetInt64(userSdk.UidKey)
		if requestUid > 0 && requestUid != req.UId {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("playStream %d %d", requestUid, req.UId)
			baseDto.ResponseForbidden(ctx)
			return
		}
		offer, errOffer := base64.StdEncoding.DecodeString(req.OfferSdp)
		if errOffer != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("playStream %d %s", req.UId, req.OfferSdp)
			baseDto.ResponseForbidden(ctx)
			return
		}
		req.OfferSdp = string(offer)
		if answer, streamKey, err := rtcService.RequestPlay(req); err != nil {
			appCtx.Logger().WithFields(logrus.Fields(claims)).Errorf("playStream %v %s", req, err.Error())
			baseDto.ResponseInternalServerError(ctx, err)
		} else {
			playResp := &dto.PlayResp{AnswerSdp: answer, StreamKey: streamKey}
			appCtx.Logger().WithFields(logrus.Fields(claims)).Infof("playStream %v %v", req, playResp)
			baseDto.ResponseSuccess(ctx, playResp)
		}
	}
}
