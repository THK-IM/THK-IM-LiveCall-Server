package rtc

import (
	"github.com/pion/webrtc/v4"
	baseDto "github.com/thk-im/thk-im-base-server/dto"
	"github.com/thk-im/thk-im-livecall-server/pkg/dto"
)

const (
	// RequestSubscribeEventKey 请求订阅事件
	RequestSubscribeEventKey  = "RequestSubscribeEvent"
	ResponseSubscribeEventKey = "ResponseSubscribeEvent"

	// NotifyClientNewStreamEventKey 用于通知客户端有新流加入
	NotifyClientNewStreamEventKey = "NotifyClientNewStreamEvent"
	// NotifyClientRemoveStreamEventKey 用于通知客户端有流退出
	NotifyClientRemoveStreamEventKey = "NotifyClientRemoveStreamEvent"

	DataChannelEventKey   = "DataChannelEvent"
	DataChannelNewEvent   = "DataChannelNewEvent"
	DataChannelMsgEvent   = "DataChannelMsgEvent"
	DataChannelCloseEvent = "DataChannelCloseEvent"
)

// RequestSubscribeEvent 请求订阅事件结构
type RequestSubscribeEvent struct {
	Req    *dto.PlayReq      `json:"req"`
	Claims baseDto.ThkClaims `json:"claims"`
}

// ResponseSubscribeEvent 响应订阅事件结构
type ResponseSubscribeEvent struct {
	Answer    string `json:"answer"`     // rtc建立连接的answer
	StreamKey string `json:"stream_key"` // 流key
	Uid       int64  `json:"uid"`        // 用户id
}

// DataChannelEvent 数据通道事件结构
type DataChannelEvent struct {
	StreamKey string                     `json:"stream_key"` // 流key
	Label     string                     `json:"label"`      // 标签
	RoomId    string                     `json:"room_id"`    // 房间id
	Uid       int64                      `json:"uid"`        // 用户id
	EventType string                     `json:"event_type"` // 事件类型，打开/关闭/发送消息:open/close/msg
	Message   *webrtc.DataChannelMessage `json:"message"`    // 消息内容
	Channel   *webrtc.DataChannelInit    `json:"channel"`    // 数据通道参数
}

//type RoomPusherEvent struct {
//	Type       string `json:"type"`
//	RoomId     string `json:"room_id"`
//	PublishKey string `json:"publish_key"`
//	Uid        int64  `json:"uid"`
//}
