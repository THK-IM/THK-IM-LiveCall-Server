package room

import "encoding/json"

const (
	ModeChat      = 1
	ModeAudio     = 2
	ModeVideo     = 3
	ModeVoiceRoom = 4
)

// Room 房间
type Room struct {
	Id           string         `json:"id"`
	Mode         int            `json:"mode"`
	OwnerId      string         `json:"owner_id"`
	CreateTime   int64          `json:"create_time"`
	Participants []*Participant `json:"participants,omitempty"`
}

func (r *Room) Json() (string, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(b), err
}

func NewRoomByJson(b []byte) (*Room, error) {
	room := &Room{}
	err := json.Unmarshal(b, room)
	return room, err
}
