package dto

type PublishStreamReq struct {
	RoomId string `json:"room_id"`
	Type   string `json:"type"`
	Sdp    string `json:"sdp"`
	Uid    int64  `json:"uid"`
}

type PublishStreamResp struct {
	SessionId string `json:"session_id"`
	Sdp       string `json:"sdp"`
	Type      string `json:"type"`
}

type SubscribeStreamReq struct {
	RoomId    string `json:"room_id"`
	SessionId string `json:"session_id"`
	Sdp       string `json:"sdp"`
	Uid       int64  `json:"uid"`
}

type SubscribeStreamResp struct {
	Renegotiation bool   `json:"renegotiation"`
	Sdp           string `json:"sdp"`
	Type          string `json:"type"`
}

type StreamStatusUpdateReq struct {
	RoomId    string `json:"room_id"`
	SessionId string `json:"session_id"`
	Status    string `json:"status"` // begin/ing/end
	Uid       int64  `json:"uid"`
}
