package dto

import "encoding/json"

const (
	InviteLiveCall = 1
	HangupLiveCall = 2
	EndLiveCall    = 3

	// BeingRequested 正在被请求通话
	BeingRequested = 1
	// CancelRequested 取消被请求通话
	CancelRequested = 2
	// RejectRequest 拒绝请求通话
	RejectRequest = 3
	// AcceptRequest 接受请求通话
	AcceptRequest = 4
	// Hangup 挂断电话
	Hangup = 5
	// EndCall 结束通话
	EndCall = 6
)

type (
	LiveCallSignal struct {
		Type int    `json:"type"`
		Body string `json:"body"`
	}

	BeingRequestedSignal struct {
		RoomId      string  `json:"room_id"`
		Members     []int64 `json:"members"`
		RequestId   int64   `json:"request_id"`
		Mode        int     `json:"mode"`
		Msg         string  `json:"msg"`
		CreateTime  int64   `json:"create_time"`
		TimeoutTime int64   `json:"timeout_time"`
	}

	CancelRequestedSignal struct {
		RoomId     string `json:"room_id"`
		Msg        string `json:"msg"`
		CreateTime int64  `json:"create_time"`
		CancelTime int64  `json:"cancel_time"`
	}

	RejectRequestSignal struct {
		RoomId     string `json:"room_id"`
		UId        int64  `json:"u_id"`
		Msg        string `json:"msg"`
		RejectTime int64  `json:"reject_time"`
	}

	AcceptRequestSignal struct {
		RoomId     string `json:"room_id"`
		UId        int64  `json:"u_id"`
		Msg        string `json:"msg"`
		AcceptTime int64  `json:"accept_time"`
	}

	HangupSignal struct {
		RoomId     string `json:"room_id"`
		UId        int64  `json:"u_id"`
		Msg        string `json:"msg"`
		HangupTime int64  `json:"hangup_time"`
	}

	EndCallSignal struct {
		RoomId      string `json:"room_id"`
		UId         int64  `json:"u_id"`
		Msg         string `json:"msg"`
		EndCallTime int64  `json:"end_call_time"`
	}
)

func MakeBeingRequestedSignal(roomId string, members []int64, mode int, msg string, uId, createTime, timeoutTime int64) *LiveCallSignal {
	signal := &BeingRequestedSignal{
		RoomId:      roomId,
		Members:     members,
		RequestId:   uId,
		Mode:        mode,
		Msg:         msg,
		CreateTime:  createTime,
		TimeoutTime: timeoutTime,
	}
	signalJson, err := json.Marshal(signal)
	if err != nil {
		return nil
	}
	return &LiveCallSignal{Type: BeingRequested, Body: string(signalJson)}
}

func MakeCancelRequestedSignal(roomId string, msg string, createTime, cancelTime int64) *LiveCallSignal {
	signal := &CancelRequestedSignal{
		RoomId:     roomId,
		Msg:        msg,
		CreateTime: createTime,
		CancelTime: cancelTime,
	}
	signalJson, err := json.Marshal(signal)
	if err != nil {
		return nil
	}
	return &LiveCallSignal{Type: CancelRequested, Body: string(signalJson)}
}

func MakeRejectRequestSignal(roomId string, msg string, uId, rejectTime int64) *LiveCallSignal {
	signal := &RejectRequestSignal{
		RoomId:     roomId,
		UId:        uId,
		Msg:        msg,
		RejectTime: rejectTime,
	}
	signalJson, err := json.Marshal(signal)
	if err != nil {
		return nil
	}
	return &LiveCallSignal{Type: RejectRequest, Body: string(signalJson)}
}

func MakeAcceptRequestSignal(roomId string, msg string, uId, acceptTime int64) *LiveCallSignal {
	signal := &AcceptRequestSignal{
		RoomId:     roomId,
		UId:        uId,
		Msg:        msg,
		AcceptTime: acceptTime,
	}
	signalJson, err := json.Marshal(signal)
	if err != nil {
		return nil
	}
	return &LiveCallSignal{Type: AcceptRequest, Body: string(signalJson)}
}

func MakeHangupSignal(roomId string, msg string, uId, hangupTime int64) *LiveCallSignal {
	signal := &HangupSignal{
		RoomId:     roomId,
		UId:        uId,
		Msg:        msg,
		HangupTime: hangupTime,
	}
	signalJson, err := json.Marshal(signal)
	if err != nil {
		return nil
	}
	return &LiveCallSignal{Type: Hangup, Body: string(signalJson)}
}

func MakeEndCallSignal(roomId string, msg string, uId, endCallTime int64) *LiveCallSignal {
	signal := &EndCallSignal{
		RoomId:      roomId,
		UId:         uId,
		Msg:         msg,
		EndCallTime: endCallTime,
	}
	signalJson, err := json.Marshal(signal)
	if err != nil {
		return nil
	}
	return &LiveCallSignal{Type: EndCall, Body: string(signalJson)}
}

func (l LiveCallSignal) JsonString() string {
	d, err := json.Marshal(l)
	if err != nil {
		return ""
	} else {
		return string(d)
	}
}
