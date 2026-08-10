package dto

import "encoding/json"

const (
	Audience  = 1
	Broadcast = 2
)

type Participant struct {
	UId       int64  `json:"u_id"`       // 用户id
	Role      int    `json:"role"`       // 1推流 2观众
	Refuse    int    `json:"refuse"`     // 是否拒绝 0未拒绝 1 拒绝 2 通话中拒绝
	JoinTime  int64  `json:"join_time"`  // 加入时间
	LeaveTime int64  `json:"leave_time"` // 离开时间
	StreamKey string `json:"stream_key"` // 订阅流的key
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
