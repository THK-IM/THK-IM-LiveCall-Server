package stat

import (
	"sync"
	"time"
)

var statPool = sync.Pool{
	New: func() interface{} {
		return new(Stat)
	},
}

type Stat struct {
	RoomId       string  `ch:"room_id"`        // 房间id
	Uid          string  `ch:"uid"`            // 用户id
	StreamId     string  `ch:"stream_id"`      // 流id
	StreamKey    string  `ch:"stream_key"`     // 流key
	SfuStreamKey string  `ch:"sfu_stream_key"` // 转发的流的key
	StreamType   int64   `ch:"stream_type"`    // 流类型 1音频/2视频
	CTime        int64   `ch:"c_time"`         // 创建时间
	PSize        int64   `ch:"p_size"`         // 已经收到包大小
	HSize        int64   `ch:"h_size"`         // 已经收到头大小
	PCount       int64   `ch:"p_count"`        // 已经收到包数量
	PLostCount   int64   `ch:"p_lost_count"`   // 丢包数量
	Jitter       float64 `ch:"jitter"`         // 延迟
}

func NewStat(streamId string, streamKey string, streamType int64, sfuStreamKey, roomId, uid string) *Stat {
	st := statPool.Get().(*Stat)
	st.StreamId = streamId
	st.StreamKey = streamKey
	st.StreamType = streamType
	st.SfuStreamKey = sfuStreamKey
	st.RoomId = roomId
	st.Uid = uid
	st.CTime = time.Now().UnixNano() / int64(time.Millisecond)
	return st
}

func ReNewStat(old *Stat, pSize, hSize, pCount, pLostCount int64, jitter float64) *Stat {
	st := statPool.Get().(*Stat)
	st.StreamId = old.StreamId
	st.StreamKey = old.StreamKey
	st.StreamType = old.StreamType
	st.SfuStreamKey = old.SfuStreamKey
	st.RoomId = old.RoomId
	st.Uid = old.Uid
	st.PSize = pSize
	st.HSize = hSize
	st.PCount = pCount
	st.PLostCount = pLostCount
	st.Jitter = jitter
	st.CTime = time.Now().UnixNano() / int64(time.Millisecond)
	return st
}

func Store(s *Stat) {
	statPool.Put(s)
}
