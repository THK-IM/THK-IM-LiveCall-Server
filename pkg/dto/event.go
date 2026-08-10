package dto

type (
	RoomUserJoinEvent struct {
		RoomId    string `json:"room_id"`
		UserId    int64  `json:"user_id"`
		Timestamp int64  `json:"timestamp"`
	}

	RoomUserPushStreamEvent struct {
		RoomId    string `json:"room_id"`
		UserId    int64  `json:"user_id"`
		StreamKey string `json:"stream_key"`
		Timestamp int64  `json:"timestamp"`
	}

	RoomUserLevelEvent struct {
		RoomId    string `json:"room_id"`
		UserId    int64  `json:"user_id"`
		Timestamp int64  `json:"timestamp"`
	}
)
