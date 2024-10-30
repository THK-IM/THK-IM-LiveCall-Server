package dto

type (
	RoomCreateReq struct {
		UId  int64 `json:"u_id"`
		Mode int   `json:"mode"`
	}

	RoomCallReq struct {
		UId      int64   `json:"u_id"`
		RoomId   string  `json:"room_id"`
		Msg      string  `json:"msg"`
		Members  []int64 `json:"members"`
		Duration int64   `json:"duration"` // 单位s
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
		Msg    string `json:"msg"`
	}

	InviteJoinRoomReq struct {
		UId        int64   `json:"u_id"`
		InviteUIds []int64 `json:"invite_u_ids"`
		RoomId     string  `json:"room_id"`
		Msg        string  `json:"msg"`
		Duration   int64   `json:"duration"` // 单位s
	}

	KickoffMemberReq struct {
		UId        int64  `json:"u_id"`
		RoomId     string `json:"room_id"`
		Msg        string `json:"msg"`
		KickoffUId int64  `json:"kickoff_u_id"`
	}
)
