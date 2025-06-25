package loader

import (
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-base-server/conf"
	msgSdk "github.com/thk-im/thk-im-msgapi-server/pkg/sdk"
)

func LoadSdks(sdkConfigs []conf.Sdk, logger *logrus.Entry) map[string]interface{} {
	sdkMap := make(map[string]interface{})
	for _, c := range sdkConfigs {
		if c.Name == "login_api" {
			loginApi := msgSdk.NewLoginApi(c, logger)
			sdkMap[c.Name] = loginApi
		} else if c.Name == "msg_api" {
			msgApi := msgSdk.NewMsgApi(c, logger)
			sdkMap[c.Name] = msgApi
		}
	}
	return sdkMap
}
