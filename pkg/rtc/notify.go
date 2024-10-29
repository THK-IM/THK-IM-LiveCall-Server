package rtc

import "encoding/json"

const (
	NotifyTypeNewStream    = "NewStream"
	NotifyTypeRemoveStream = "RemoveStream"
	NotifyTypeDataMsg      = "DataChannelMsg"
)

// Notify 下发给客户端的通知结构
type Notify struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// PublishEvent 流发布事件结构 房间有流加入/退出触发，下发给客户端
type PublishEvent struct {
	RoomId    string `json:"room_id"`    // 房间id
	StreamKey string `json:"stream_key"` // 流key
	UId       int64  `json:"u_id"`       // 用户id
	Role      int    `json:"role"`       // 角色
}

func NewStreamNotify(msg string) *string {
	notify := &Notify{
		Type:    NotifyTypeNewStream,
		Message: msg,
	}
	if js, err := json.Marshal(notify); err != nil {
		return nil
	} else {
		jsStr := string(js)
		return &jsStr
	}
}

func RemoveStreamNotify(msg string) *string {
	notify := &Notify{
		Type:    NotifyTypeRemoveStream,
		Message: msg,
	}
	if js, err := json.Marshal(notify); err != nil {
		return nil
	} else {
		jsStr := string(js)
		return &jsStr
	}
}

func DataChanelMsg(msg string) *string {
	notify := &Notify{
		Type:    NotifyTypeDataMsg,
		Message: msg,
	}
	if js, err := json.Marshal(notify); err != nil {
		return nil
	} else {
		jsStr := string(js)
		return &jsStr
	}
}
