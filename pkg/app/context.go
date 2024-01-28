package app

import (
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-base-server/server"
	"github.com/thk-im/thk-im-livecall-server/pkg/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/loader"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/cache"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/room"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/stat"
	userSdk "github.com/thk-im/thk-im-user-server/pkg/sdk"
)

type Context struct {
	startTime    int64
	statService  stat.Service
	cacheService cache.Service
	roomService  room.Service
	logger       *logrus.Entry
	*server.Context
}

func (c *Context) RoomService() room.Service {
	return c.roomService
}

func (c *Context) CacheService() cache.Service {
	return c.cacheService
}

func (c *Context) StatService() stat.Service {
	return c.statService
}

func (c *Context) Init(config *conf.LiveCallConfig) {
	c.Context = &server.Context{}
	c.Context.Init(config.Config)
	c.Context.SdkMap = loader.LoadSdks(c.Config().Sdks, c.Logger())
	logger := c.Context.Logger()
	cacheService := loader.LoadCacheService(config.Cache, logger)
	c.roomService = loader.LoadRoomService(cacheService, logger)
	c.statService = loader.LoadStatService(config.Stat, logger)
	c.cacheService = cacheService
}

func (c *Context) UserApi() userSdk.UserApi {
	return c.Context.SdkMap["user_api"].(userSdk.UserApi)
}

func (c *Context) StartServe() {
	c.Context.StartServe()
}
