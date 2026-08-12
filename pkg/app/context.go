package app

import (
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-base-server/server"
	"github.com/thk-im/thk-im-livecall-server/pkg/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/loader"
	"github.com/thk-im/thk-im-livecall-server/pkg/sdk"
	msgSdk "github.com/thk-im/thk-im-msgapi-server/pkg/sdk"
	rtcSdk "github.com/thk-im/thk-im-rtc-server/pkg/sdk"
)

type Context struct {
	startTime int64
	logger    *logrus.Entry
	*server.Context
}

func (c *Context) Init(config *conf.LiveCallConfig) {
	c.Context = &server.Context{}
	c.Context.Init(config.Config)
	c.Context.SdkMap = loader.LoadSdks(config, c.Logger())
	c.Context.ModelMap = loader.LoadModels(c.Config().Models, c.Database(), c.Logger(), c.SnowflakeNode())
	err := loader.LoadTables(c.Config().Models, c.Database())
	if err != nil {
		panic(err)
	}
}

func (c *Context) LoginApi() msgSdk.LoginApi {
	return c.Context.SdkMap["login_api"].(msgSdk.LoginApi)
}

func (c *Context) MsgApi() msgSdk.MsgApi {
	return c.Context.SdkMap["msg_api"].(msgSdk.MsgApi)
}

func (c *Context) WebRTCApi() rtcSdk.RTCApi {
	if c.Context.SdkMap["rtc_api"] == nil {
		return nil
	}
	return c.Context.SdkMap["rtc_api"].(rtcSdk.RTCApi)
}

func (c *Context) CloudflareConnectApi() sdk.SfuApi {
	if c.Context.SdkMap["cloudflare_connect_api"] == nil {
		return nil
	}
	return c.Context.SdkMap["cloudflare_connect_api"].(sdk.SfuApi)
}

func (c *Context) StartServe() {
	c.Context.StartServe()
}
