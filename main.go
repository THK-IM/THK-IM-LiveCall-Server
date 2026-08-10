package main

import (
	baseConf "github.com/thk-im/thk-im-base-server/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/app"
	"github.com/thk-im/thk-im-livecall-server/pkg/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/handler"
)

func main() {
	configPath := "etc/msg_live_call_server.yaml"
	config := &conf.LiveCallConfig{}
	if err := baseConf.LoadConfig(configPath, config); err != nil {
		panic(err)
	}

	appCtx := &app.Context{}
	appCtx.Init(config)
	handler.RegisterRtcHandler(appCtx)

	appCtx.StartServe()
}
