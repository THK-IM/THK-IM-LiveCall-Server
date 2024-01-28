package loader

import (
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-base-server/conf"
	userSdk "github.com/thk-im/thk-im-user-server/pkg/sdk"
)

func LoadSdks(sdkConfigs []conf.Sdk, logger *logrus.Entry) map[string]interface{} {
	sdkMap := make(map[string]interface{})
	for _, c := range sdkConfigs {
		if c.Name == "user_api" {
			userApi := userSdk.NewUserApi(c, logger)
			sdkMap[c.Name] = userApi
		}
	}
	return sdkMap
}
