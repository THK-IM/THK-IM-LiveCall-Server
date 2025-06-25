package dto

type (
	CheckLiveCallCreateReq struct {
		UId      int64 `json:"uid" form:"uid"` // 1普通聊天 2语音电话 3视频电话 4语音房 5视频房
		RoomType int   `json:"room_type" form:"room_type"`
	}

	CheckLiveJoinReq struct {
		Room *Room `json:"room" form:"room"`
		UId  int64 `json:"u_id" form:"u_id"`
	}

	CheckLiveInviteReq struct {
		Room       *Room   `json:"room" form:"room"`
		InviteUIds []int64 `json:"invite_u_ids" form:"invite_u_ids"`
		RequestUId int64   `json:"request_u_id" form:"request_u_id"`
	}

	CheckLiveCallStatusReq struct {
		Room *Room `json:"room" form:"room"`
	}
)
