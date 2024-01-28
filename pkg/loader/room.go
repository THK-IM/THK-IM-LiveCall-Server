package loader

import (
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/cache"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/room"
)

func LoadRoomService(cacheService cache.Service, logger *logrus.Entry) room.Service {
	return room.NewService(cacheService, logger)
}
