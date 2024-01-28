package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/rtc"
)

func RegisterRtcHandler(engine *gin.Engine, app *app.Context, rtcService rtc.Service) {

	room := engine.Group("/room")
	// curl -i -X POST "http://127.0.0.1:18100/room" -d '{"uid": "1"}'
	room.POST("", createRoom(app))
	room.GET("", findRooms(app))
	room.GET("/:id", findRoomById(app))
	room.GET("/:id/stream", findStreamByRoomId(app))
	// curl -i -X POST "http://127.0.0.1:18100/room/join" -d '{"u_id": 2, "room_id": "1", "role": 1, "token": "xxxxxx"}'
	room.POST("/join", joinRoom(app))

	stream := engine.Group("/stream")
	stream.POST("/publish", publish(rtcService))
	stream.POST("/play", play(rtcService))
}
