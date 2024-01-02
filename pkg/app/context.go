package app

import (
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/cache"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/room"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/stat"
)

type Context struct {
	startTime    int64
	statService  stat.Service
	cacheService cache.Service
	roomService  room.Service
	logger       *logrus.Entry
}

func (a *Context) RoomService() room.Service {
	return a.roomService
}

func (a *Context) CacheService() cache.Service {
	return a.cacheService
}

func (a *Context) StatService() stat.Service {
	return a.statService
}

func (a *Context) StartTime() int64 {
	return a.startTime
}

func (a *Context) Logger() *logrus.Entry {
	return a.logger
}
