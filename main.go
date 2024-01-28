package main

import (
	baseConf "github.com/thk-im/thk-im-base-server/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/handler"
	"github.com/thk-im/thk-im-livecall-server/pkg/rtc"
)

func main() {
	configPath := "etc/msg_live_call_server.yaml"
	config := &conf.LiveCallConfig{}
	if err := baseConf.LoadConfig(configPath, config); err != nil {
		panic(err)
	}

	appCtx := &app.Context{}
	appCtx.Init(config)
	rtcService := rtc.NewRtcService(config.Rtc, appCtx)
	rtcService.InitServer()
	handler.RegisterRtcHandler(appCtx, rtcService)

	appCtx.StartServe()
}
