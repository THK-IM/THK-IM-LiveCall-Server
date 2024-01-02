package loader

import (
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/cache"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/room"
)

func LoadRoomService(logger *logrus.Entry, cacheService cache.Service) room.Service {
	return room.NewService(logger, cacheService)
}
