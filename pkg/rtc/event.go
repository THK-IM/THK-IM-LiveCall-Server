package rtc

import (
	"github.com/pion/webrtc/v3"
)

const (
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

type RequestSubscribeEvent struct {
	RoomId    string `json:"room_id"`
	Uid       string `json:"uid"`
	OfferSdp  string `json:"offer_sdp"`
	StreamKey string `json:"stream_key"`
}

type ResponseSubscribeEvent struct {
	Answer    string `json:"answer"`
	StreamKey string `json:"stream_key"`
	Uid       string `json:"uid"`
}

type PublishEvent struct {
	RoomId    string `json:"room_id"`
	Uid       string `json:"uid"`
	StreamKey string `json:"stream_key"`
	Role      int    `json:"role"`
}

type DataChannelEvent struct {
	StreamKey string                     `json:"stream_key"`
	Label     string                     `json:"label"`
	RoomId    string                     `json:"room_id"`
	UId       string                     `json:"u_id"`
	EventType string                     `json:"event_type"` // open/close/msg
	Message   *webrtc.DataChannelMessage `json:"message"`
	Channel   *webrtc.DataChannelInit    `json:"channel"`
}

type RoomPusherEvent struct {
	Type       string `json:"type"`
	RoomId     string `json:"room_id"`
	Uid        string `json:"uid"`
	PublishKey string `json:"publish_key"`
}
