package dto

import "encoding/json"

type LiveCallNotify struct {
	RoomId     string  `json:"room_id"`
	Mode       int     `json:"mode"`
	OwnerId    int64   `json:"owner_id"`
	CreateTime int64   `json:"create_time"`
	Members    []int64 `json:"members"`
}

func (l *LiveCallNotify) JsonString() string {
	b, err := json.Marshal(l)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}
