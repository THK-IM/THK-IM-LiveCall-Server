package handler

import (
	"crypto/subtle"
	"os"

	"github.com/gin-gonic/gin"
	baseMiddleware "github.com/thk-im/thk-im-base-server/middleware"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	msgsdk "github.com/thk-im/thk-im-msgapi-server/pkg/sdk"
)

func RegisterRtcHandler(appCtx *app.Context) {
	httpEngine := appCtx.HttpEngine()
	loginApi := appCtx.LoginApi()
	userTokenAuth := msgsdk.UserTokenAuth(loginApi, appCtx.Logger())
	ipAuth := baseMiddleware.WhiteIpAuth(appCtx.Config().IpWhiteList, appCtx.Logger())
	liveCallRoute := httpEngine.Group("/live_call")

	room := liveCallRoute.Group("/room", userTokenAuth)
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

	rtcEvent := liveCallRoute.Group("/rtc_event")
	rtcEvent.Use(ipAuth)
	if token := os.Getenv("THK_RTC_EVENT_TOKEN"); token != "" {
		rtcEvent.Use(func(ctx *gin.Context) {
			if subtle.ConstantTimeCompare([]byte(ctx.GetHeader("X-RTC-Event-Token")), []byte(token)) != 1 {
				ctx.AbortWithStatus(403)
			}
		})
	} else {
		rtcEvent.Use(userTokenAuth)
	}
	rtcEvent.POST("/user_join", rtcUserJoinEvent(appCtx))
	rtcEvent.POST("/user_leave", rtcUserLeaveEvent(appCtx))
	rtcEvent.POST("/user_push", rtcUserPushEvent(appCtx))

	streamRoute := liveCallRoute.Group("/stream")
	streamRoute.Use(userTokenAuth)
	streamRoute.POST("/publish", publishStream(appCtx))
	streamRoute.POST("/subscribe", subscribeStream(appCtx))
	streamRoute.PUT("/status", updateStreamStatus(appCtx))
}
