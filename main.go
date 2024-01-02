package main

import (
	baseConf "github.com/thk-im/thk-im-base-server/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/conf"
)

func main() {
	configPath := "etc/msg_live_call_server.yaml"
	config := &conf.LiveCallConfig{}
	if err := baseConf.LoadConfig(configPath, config); err != nil {
		panic(err)
	}

}
