package model

import (
	"encoding/json"
)

const (
	ModeChat      = 1
	ModeAudio     = 2
	ModeVideo     = 3
	ModeVoiceRoom = 4
)

type MediaPrams struct {
	VideoMaxBitrate int `json:"video_max_bitrate"` // 视频最大码率
	VideoWidth      int `json:"video_width"`       // 视频分辨率宽
	VideoHeight     int `json:"video_height"`      // 视频分辨率高
	VideoFps        int `json:"video_fps"`         // 视频每秒帧
	AudioMaxBitrate int `json:"audio_max_bitrate"` // 音频最大码率
}

// Room 房间
type Room struct {
	Id           string         `json:"id"`                     // 房间id
	Mode         int            `json:"mode"`                   // 模式， 1普通聊天 2语音电话 3视频电话 4语音房 5视频房
	OwnerId      int64          `json:"owner_id"`               // 房间创建者id
	CreateTime   int64          `json:"create_time"`            // 房间创建时间
	MediaPrams   MediaPrams     `json:"media_prams"`            // 媒体参数
	Participants []*Participant `json:"participants,omitempty"` // 房间实际参与人
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
