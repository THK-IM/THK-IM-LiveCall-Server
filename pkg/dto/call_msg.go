package dto

import "time"

type CallMsg struct {
	RoomId      string  `json:"room_id"`
	RoomOwnerId int64   `json:"room_owner_id"`
	RoomMode    int     `json:"room_mode"`
	CreateTime  int64   `json:"create_time"`
	Accepted    int     `json:"accepted"` // 0未接听 1被挂断 2已接通 3通话中被挂断
	AcceptTime  int64   `json:"accept_time"`
	Duration    int64   `json:"duration"`
	JoinedUIds  []int64 `json:"joined_u_ids"`
}

func BuildCallMsg(room *Room) CallMsg {
	accepted := 0
	acceptTime := int64(0)
	duration := int64(0)
	leaveTime := time.Now().UnixMilli()
	joinedUIds := make([]int64, 0)
	for _, p := range room.Participants {
		if p.UId != room.OwnerId {
			if p.JoinTime > 0 {
				accepted = 2
				// 不算房主最早加入的人
				if p.JoinTime < acceptTime || acceptTime == 0 {
					acceptTime = p.JoinTime
				}
			}
			if p.Refuse == 1 {
				accepted = 1
			} else if p.Refuse == 2 {
				accepted = 3
			}
		}
		if p.JoinTime > 0 {
			joinedUIds = append(joinedUIds, p.UId)
		}
	}
	if accepted > 0 {
		duration = leaveTime - acceptTime
	}
	// 房主已经离开了，其他人在进来的
	if duration < 0 {
		accepted = 0
	}
	return CallMsg{
		RoomId:      room.Id,
		RoomOwnerId: room.OwnerId,
		RoomMode:    room.Mode,
		CreateTime:  room.CreateTime,
		Accepted:    accepted,
		AcceptTime:  acceptTime,
		Duration:    duration,
		JoinedUIds:  joinedUIds,
	}
}
