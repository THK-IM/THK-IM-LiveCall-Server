package rtc

type (
	// 连接已建立
	onConnConnected func(roomId, uId, streamKey, subKey string)

	// 连接已关闭
	onConnClosed func(roomId, uId, streamKey, subKey string)

	// DataChannel事件
	onDataChannelEvent func(event *DataChannelEvent)

	// 房间内进新人请求拉流请求事件，各子节点根据请求拉流的key给出响应
	onPullRequestEvent func(resp *ResponseSubscribeEvent)

	writeKeyFrame func(roomId string, subKey string)
)
