package rtc

import baseDto "github.com/thk-im/thk-im-base-server/dto"

type (
	// 房间内进新人请求拉流请求事件，各子节点根据请求拉流的key给出响应
	onPullRequestEvent func(resp *ResponseSubscribeEvent)
)

type Callback interface {
	// OnPusherConnected 连接
	OnPusherConnected(roomId, key, subKey string, uid int64, claims baseDto.ThkClaims)
	// OnPusherSteaming 正在推流
	OnPusherSteaming(roomId, key, subKey string, uid int64, time int64, claims baseDto.ThkClaims)
	// OnPusherClosed 关闭
	OnPusherClosed(roomId, key, subKey string, uid int64, claims baseDto.ThkClaims)
	// OnPullerConnected 关闭
	OnPullerConnected(roomId, key, subKey string, uid int64, claims baseDto.ThkClaims)
	// OnPullStreaming 正在拉流
	OnPullStreaming(roomId, key, subKey string, uid int64, time int64, claims baseDto.ThkClaims)
	// OnPullerClosed 关闭
	OnPullerClosed(roomId, key, subKey string, uid int64, claims baseDto.ThkClaims)
	// OnDataChannelEvent 数据通道事件
	OnDataChannelEvent(event *DataChannelEvent, claims baseDto.ThkClaims)
}
