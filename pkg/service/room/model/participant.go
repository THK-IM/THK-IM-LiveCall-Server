package model

import "encoding/json"

const (
	Audience  = 1
	Broadcast = 2
)

type Participant struct {
	UId       int64   `json:"u_id"`                 // 用户id
	Role      int     `json:"role"`                 // 1主播 2观众
	JoinTime  int64   `json:"join_time"`            // 加入时间
	StreamKey *string `json:"stream_key,omitempty"` // 推流key
}

func (r *Participant) Json() (string, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(b), err
}

func NewParticipantByJson(b []byte) (*Participant, error) {
	p := &Participant{}
	err := json.Unmarshal(b, p)
	return p, err
}
