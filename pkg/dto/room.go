package dto

import "encoding/json"

type (
	MediaParams struct {
		VideoMaxBitrate int `json:"video_max_bitrate"` // 视频最大码率
		VideoWidth      int `json:"video_width"`       // 视频分辨率宽
		VideoHeight     int `json:"video_height"`      // 视频分辨率高
		VideoFps        int `json:"video_fps"`         // 视频每秒帧
		AudioMaxBitrate int `json:"audio_max_bitrate"` // 音频最大码率
	}

	RoomCreateReq struct {
		UId         int64        `json:"u_id"`
		Mode        int          `json:"mode"` // 1普通聊天 2语音电话 3视频电话 4语音房 5视频房
		MediaParams *MediaParams `json:"media_params"`
	}

	RoomCallReq struct {
		UId      int64   `json:"u_id"`
		RoomId   string  `json:"room_id"`
		Msg      string  `json:"msg"`
		Members  []int64 `json:"members"`
		Duration int64   `json:"duration"` // 单位s
	}

	CancelCallingReq struct {
		UId     int64   `json:"u_id"`
		RoomId  string  `json:"room_id"`
		Msg     string  `json:"msg"`
		Members []int64 `json:"members"`
	}

	RoomDelReq struct {
		UId    int64  `json:"u_id"`
		RoomId string `json:"room_id"`
	}

	RoomJoinReq struct {
		UId    int64  `json:"u_id"`
		RoomId string `json:"room_id"`
		Role   int8   `json:"role"`
	}

	RefuseJoinRoomReq struct {
		UId    int64  `json:"u_id"`
		RoomId string `json:"room_id"`
		Msg    string `json:"msg"`
	}

	InviteJoinRoomReq struct {
		UId        int64   `json:"u_id"`
		InviteUIds []int64 `json:"invite_u_ids"`
		RoomId     string  `json:"room_id"`
		Msg        string  `json:"msg"`
		Duration   int64   `json:"duration"` // 单位s
	}

	RoomMemberLeaveReq struct {
		UId    int64  `json:"u_id"`
		RoomId string `json:"room_id"`
		Msg    string `json:"msg"`
	}

	KickoffMemberReq struct {
		UId         int64   `json:"u_id"`
		RoomId      string  `json:"room_id"`
		Msg         string  `json:"msg"`
		KickoffUIds []int64 `json:"kickoff_u_ids"`
	}
)

const (
	ModeChat      = 1
	ModeAudio     = 2
	ModeVideo     = 3
	ModeVoiceRoom = 4
	ModeVideoRoom = 5
)

// Room 房间
type Room struct {
	Id           string         `json:"id"`                     // 房间id
	Mode         int            `json:"mode"`                   // 模式， 1普通聊天 2语音电话 3视频电话 4语音房 5视频房
	OwnerId      int64          `json:"owner_id"`               // 房间创建者id
	CreateTime   int64          `json:"create_time"`            // 房间创建时间
	MediaParams  *MediaParams   `json:"media_params"`           // 媒体参数
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
