package dto

import "encoding/json"

const (
	InviteLiveCall = 1
	HangupLiveCall = 2
	EndLiveCall    = 3
)

type LiveCallSignal struct {
	RoomId     string  `json:"room_id"`
	Mode       int     `json:"mode"`
	OwnerId    int64   `json:"owner_id"`
	CreateTime int64   `json:"create_time"`
	Members    []int64 `json:"members"`
	MsgType    int     `json:"msg_type"`
	OperatorId int64   `json:"operator_id"`
}

func (l *LiveCallSignal) JsonString() string {
	b, err := json.Marshal(l)
	if err != nil {
		return ""
	} else {
		return string(b)
	}
}
