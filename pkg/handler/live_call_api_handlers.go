package handler

import (
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/rtc"
	userSdk "github.com/thk-im/thk-im-user-server/pkg/sdk"
)

func RegisterRtcHandler(appCtx *app.Context, rtcService rtc.Service) {

	httpEngine := appCtx.HttpEngine()
	userApi := appCtx.UserApi()
	userTokenAuth := userSdk.UserTokenAuth(userApi, appCtx.Logger())
	httpEngine.Use(userTokenAuth)

	room := httpEngine.Group("/room")
	room.POST("", createRoom(appCtx))
	room.GET("/:id", findRoomById(appCtx))
	room.POST("/member/join", joinRoom(appCtx))
	room.POST("/member/invite", inviteJoinRoom(appCtx))
	room.POST("/member/hangup", hangup(appCtx))
	room.DELETE("", deleteRoom(appCtx))

	stream := httpEngine.Group("/stream")
	stream.POST("/publish", publishStream(appCtx, rtcService))
	stream.POST("/play", playStream(appCtx, rtcService))
}
