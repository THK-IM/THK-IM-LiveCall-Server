package loader

import (
	"github.com/sirupsen/logrus"
	"github.com/thk-im/thk-im-livecall-server/pkg/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/sdk"
	msgSdk "github.com/thk-im/thk-im-msgapi-server/pkg/sdk"
	rtcSdk "github.com/thk-im/thk-im-rtc-server/pkg/sdk"
)

func LoadSdks(config *conf.LiveCallConfig, logger *logrus.Entry) map[string]interface{} {
	sdkMap := make(map[string]interface{})
	for _, c := range config.Sdks {
		logger.Infof("LoadSdks %s %s", c.Name, c.Endpoint)
		if c.Name == "login_api" {
			loginApi := msgSdk.NewLoginApi(c, logger)
			sdkMap[c.Name] = loginApi
		} else if c.Name == "msg_api" {
			msgApi := msgSdk.NewMsgApi(c, logger)
			sdkMap[c.Name] = msgApi
		} else if c.Name == "web_rtc_api" {
			rtcApi := rtcSdk.NewRTCChannelApi(c, logger)
			sdkMap[c.Name] = rtcApi
		} else if c.Name == "cloudflare_connect_api" {
			meetingApi := sdk.NewSfuApi(c, logger)
			sdkMap[c.Name] = meetingApi
		}
	}
	return sdkMap
}
