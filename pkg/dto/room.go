package dto

type (
	RoomCreateReq struct {
		Uid  string `json:"uid"`
		Mode int    `json:"mode"`
	}

	RoomJoinReq struct {
		Uid    string `json:"uid"`
		RoomId string `json:"room_id"`
		Role   int8   `json:"role"`
	}
)
