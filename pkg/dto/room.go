package dto

type (
	RoomCreateReq struct {
		UId     int64   `json:"u_id"`
		Mode    int     `json:"mode"`
		Members []int64 `json:"members"`
	}

	RoomCallReq struct {
		UId      int64  `json:"u_id"`
		RoomId   string `json:"room_id"`
		Msg      string `json:"msg"`
		Duration int64  `json:"duration"` // 单位s
	}

	RoomDelReq struct {
		UId    int64  `json:"u_id"`
		RoomId string `json:"room_id"`
	}

	RoomJoinReq struct {
		UId    int64  `json:"u_id"`
		RoomId string `json:"room_id"`
		Role   int8   `json:"role"`
	}

	RefuseJoinRoomReq struct {
		UId    int64  `json:"u_id"`
		RoomId string `json:"room_id"`
	}

	InviteJoinRoomReq struct {
		UId        int64   `json:"u_id"`
		InviteUIds []int64 `json:"invite_u_ids"`
		RoomId     string  `json:"room_id"`
		Msg        string  `json:"msg"`
	}
)
