package model

import (
	"encoding/json"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
)

const (
	ModeChat      = 1
	ModeAudio     = 2
	ModeVideo     = 3
	ModeVoiceRoom = 4
)

// Room 房间
type Room struct {
	Id           string           `json:"id"`                     // 房间id
	Mode         int              `json:"mode"`                   // 模式， 1普通聊天 2语音电话 3视频电话 4语音房 5视频房
	OwnerId      int64            `json:"owner_id"`               // 房间创建者id
	CreateTime   int64            `json:"create_time"`            // 房间创建时间
	MediaParams  *dto.MediaParams `json:"media_params"`           // 媒体参数
	Participants []*Participant   `json:"participants,omitempty"` // 房间实际参与人
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
