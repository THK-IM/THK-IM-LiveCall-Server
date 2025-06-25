package app

import (
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-base-server/server"
	"github.com/thk-im/thk-im-livecall-server/pkg/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/loader"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/room"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/room/cache"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/stat"
	msgSdk "github.com/thk-im/thk-im-msgapi-server/pkg/sdk"
)

type Context struct {
	startTime   int64
	statService stat.Service
	roomCache   cache.RoomCache
	roomService room.Service
	logger      *logrus.Entry
	signalType  int
	*server.Context
}

func (c *Context) RoomService() room.Service {
	return c.roomService
}

func (c *Context) RoomCache() cache.RoomCache {
	return c.roomCache
}

func (c *Context) StatService() stat.Service {
	return c.statService
}

func (c *Context) SignalType() int {
	return c.signalType
}

func (c *Context) Init(config *conf.LiveCallConfig) {
	c.Context = &server.Context{}
	c.Context.Init(config.Config)
	c.Context.SdkMap = loader.LoadSdks(c.Config().Sdks, c.Logger())
	logger := c.Context.Logger()
	cacheService := loader.LoadRoomCache(config.Cache, logger)
	c.roomService = loader.LoadRoomService(c.SnowflakeNode(), cacheService, logger)
	c.statService = loader.LoadStatService(config.Stat, logger)
	c.roomCache = cacheService
	c.signalType = config.SignalType
}

func (c *Context) LoginApi() msgSdk.LoginApi {
	return c.Context.SdkMap["login_api"].(msgSdk.LoginApi)
}

func (c *Context) MsgApi() msgSdk.MsgApi {
	return c.Context.SdkMap["msg_api"].(msgSdk.MsgApi)
}

func (c *Context) StartServe() {
	c.Context.StartServe()
}
