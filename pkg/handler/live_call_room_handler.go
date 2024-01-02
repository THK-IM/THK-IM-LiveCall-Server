package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
)

func createRoom(appCtx *app.Context) gin.HandlerFunc {
	return func(context *gin.Context) {
		createRoomReq := &dto.RoomCreateReq{}
		if e := context.BindJSON(createRoomReq); e != nil {
			dto.ResponseBadRequest(context, e)
			return
		}
		if room, err := appCtx.RoomService().CreateRoom(createRoomReq); err != nil {
			dto.ResponseInternalServerError(context, err)
		} else {
			dto.ResponseSuccess(context, room)
		}
	}
}

func joinRoom(appCtx *app.Context) gin.HandlerFunc {
	return func(context *gin.Context) {
		joinRoomReq := &dto.RoomJoinReq{}
		if e := context.BindJSON(joinRoomReq); e != nil {
			dto.ResponseBadRequest(context, e)
			return
		}
		if room, err := appCtx.RoomService().JoinRoom(joinRoomReq); err != nil {
			dto.ResponseInternalServerError(context, err)
		} else {
			appCtx.Logger().Infof("joinRoom, uid: %s, room: %v", joinRoomReq.Uid, room)
			dto.ResponseSuccess(context, room)
		}
	}
}

func findRooms(appCtx *app.Context) gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}

func findRoomById(appCtx *app.Context) gin.HandlerFunc {
	return func(context *gin.Context) {
		roomId := context.Param("id")
		if room, err := appCtx.RoomService().FindRoomById(roomId); err != nil {
			dto.ResponseInternalServerError(context, err)
		} else {
			dto.ResponseSuccess(context, room)
		}
	}
}

func findStreamByRoomId(app *app.Context) gin.HandlerFunc {
	return func(context *gin.Context) {

	}
}
