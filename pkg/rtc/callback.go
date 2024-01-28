package rtc

type (
	// 连接已建立
	onConnConnected func(roomId, streamKey, subKey string, uId int64)

	// 连接已关闭
	onConnClosed func(roomId, streamKey, subKey string, uId int64)

	// DataChannel事件
	onDataChannelEvent func(event *DataChannelEvent)

	// 房间内进新人请求拉流请求事件，各子节点根据请求拉流的key给出响应
	onPullRequestEvent func(resp *ResponseSubscribeEvent)

	// 写关键帧
	writeKeyFrame func(roomId string, subKey string)
)
