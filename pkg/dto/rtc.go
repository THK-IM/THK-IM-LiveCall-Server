package dto

type PublishReq struct {
	RoomId   string `json:"room_id"`
	UId      int64  `json:"u_id"`
	OfferSdp string `json:"offer_sdp"`
}

type PublishResp struct {
	AnswerSdp string `json:"answer_sdp"`
	StreamKey string `json:"stream_key"`
}

type PlayReq struct {
	RoomId    string `json:"room_id"`
	UId       int64  `json:"u_id"`
	OfferSdp  string `json:"offer_sdp"`
	StreamKey string `json:"stream_key"`
}

type PlayResp struct {
	AnswerSdp string `json:"answer_sdp"`
	StreamKey string `json:"stream_key"`
}
