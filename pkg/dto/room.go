package dto

type (
	RoomCreateReq struct {
		UId             int64 `json:"u_id"`
		Mode            int   `json:"mode"`
		VideoMaxBitrate int   `json:"video_max_bitrate"` // 视频最大码率
		VideoWidth      int   `json:"video_width"`       // 视频分辨率宽
		VideoHeight     int   `json:"video_height"`      // 视频分辨率高
		VideoFps        int   `json:"video_fps"`         // 视频每秒帧
		AudioMaxBitrate int   `json:"audio_max_bitrate"` // 音频最大码率
	}

	RoomCallReq struct {
		UId      int64   `json:"u_id"`
		RoomId   string  `json:"room_id"`
		Msg      string  `json:"msg"`
		Members  []int64 `json:"members"`
		Duration int64   `json:"duration"` // 单位s
	}

	CancelCallingReq struct {
		UId     int64   `json:"u_id"`
		RoomId  string  `json:"room_id"`
		Msg     string  `json:"msg"`
		Members []int64 `json:"members"`
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

	RoomMemberLeaveReq struct {
		UId    int64  `json:"u_id"`
		RoomId string `json:"room_id"`
		Msg    string `json:"msg"`
	}

	KickoffMemberReq struct {
		UId         int64   `json:"u_id"`
		RoomId      string  `json:"room_id"`
		Msg         string  `json:"msg"`
		KickoffUIds []int64 `json:"kickoff_u_ids"`
	}
)
