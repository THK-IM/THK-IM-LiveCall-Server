package dto

type (
	RoomCreateReq struct {
		UId  int64 `json:"u_id"`
		Mode int   `json:"mode"`
	}

	RoomJoinReq struct {
		UId    int64  `json:"u_id"`
		RoomId string `json:"room_id"`
		Role   int8   `json:"role"`
	}
)
