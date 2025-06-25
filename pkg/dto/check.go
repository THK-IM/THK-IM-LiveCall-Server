package dto

type (
	CheckCreateRoomReq struct {
		UId      int64 `json:"uid" form:"uid"`
		RoomType int   `json:"room_type" form:"room_type"`
	}

	CheckJoinRoomReq struct {
		UId      int64 `json:"uid" form:"uid"`
		RoomId   int64 `json:"room_id" form:"room_id"`
		RoomType int   `json:"room_type" form:"room_type"`
	}
)
