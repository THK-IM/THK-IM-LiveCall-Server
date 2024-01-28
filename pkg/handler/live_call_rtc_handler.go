package handler

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/rtc"
)

func publish(rtcService rtc.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		publishReq := &dto.PublishReq{}
		if err := context.BindJSON(publishReq); err != nil {
			dto.ResponseBadRequest(context, err)
			return
		}
		offer, eb := base64.StdEncoding.DecodeString(publishReq.OfferSdp)
		if eb != nil {
			dto.ResponseBadRequest(context, eb)
			return
		}
		if conn, err := rtcService.RequestPublish(publishReq.RoomId, publishReq.Uid, string(offer)); err != nil {
			dto.ResponseInternalServerError(context, err)
		} else {
			answer := base64.StdEncoding.EncodeToString([]byte(conn.ServerSdp().SDP))
			dto.ResponseSuccess(context, &dto.PublishResp{AnswerSdp: answer, StreamKey: conn.Key()})
		}
	}
}

func play(rtcService rtc.Service) gin.HandlerFunc {
	return func(context *gin.Context) {
		playReq := &dto.PlayReq{}
		if err := context.BindJSON(playReq); err != nil {
			dto.ResponseBadRequest(context, err)
			return
		}
		offer, eb := base64.StdEncoding.DecodeString(playReq.OfferSdp)
		if eb != nil {
			dto.ResponseBadRequest(context, eb)
			return
		}
		playReq.OfferSdp = string(offer)
		if answer, streamKey, err := rtcService.RequestPlay(playReq); err != nil {
			dto.ResponseInternalServerError(context, err)
		} else {
			dto.ResponseSuccess(context, &dto.PlayResp{AnswerSdp: answer, StreamKey: streamKey})
		}
	}
}
