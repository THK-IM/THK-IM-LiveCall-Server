package errorx

import "github.com/thk-im/thk-im-base-server/errorx"

var (
	ErrRoomNotExisted = errorx.NewErrorX(4004001, "RoomNotExisted")
	ErrNoPermission   = errorx.NewErrorX(4004002, "NoPermission")
)
