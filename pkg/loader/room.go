package loader

import (
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-base-server/snowflake"
	"github.com/thk-im/thk-im-livecall-server/pkg/sdk"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/room"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/room/cache"
)

func LoadRoomService(node *snowflake.Node, cacheService cache.RoomCache, checkApi sdk.CheckApi, logger *logrus.Entry) room.Service {
	return room.NewService(node, cacheService, logger)
}
