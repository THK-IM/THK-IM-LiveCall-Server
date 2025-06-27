package handler

import (
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/rtc"
	msgsdk "github.com/thk-im/thk-im-msgapi-server/pkg/sdk"
)

func RegisterRtcHandler(appCtx *app.Context, rtcService rtc.Service) {
	httpEngine := appCtx.HttpEngine()
	loginApi := appCtx.LoginApi()
	userTokenAuth := msgsdk.UserTokenAuth(loginApi, appCtx.Logger())
	httpEngine.Use(userTokenAuth)
	liveCallRoute := httpEngine.Group("/live_call")

	room := liveCallRoute.Group("/room")
	room.POST("", createRoom(appCtx))
	room.POST("/call", callRoomMembers(appCtx))
	room.POST("/cancel_call", cancelCallRoomMembers(appCtx))
	room.GET("/:id", findRoomById(appCtx))
	room.POST("/member/join", joinRoom(appCtx))
	room.POST("/member/invite", inviteJoinRoom(appCtx))
	room.POST("/member/refuse_join", refuseJoinRoom(appCtx))
	room.POST("/member/kick", KickoffRoomMember(appCtx))
	room.POST("/member/leave", leaveRoomMember(appCtx))
	room.DELETE("", deleteRoom(appCtx))

	stream := liveCallRoute.Group("/stream")
	stream.POST("/publish", publishStream(appCtx, rtcService))
	stream.POST("/play", playStream(appCtx, rtcService))
}
