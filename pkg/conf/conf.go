package conf

import (
	baseConf "github.com/thk-im/thk-im-base-server/conf"
)

type Rtc struct {
	Timeout int64  `yaml:"Timeout"`
	NodeIp  string `yaml:"NodeIp"`
	UdpPort int    `yaml:"UdpPort"`
	TcpPort int    `yaml:"TcpPort"`
}

type Cache struct {
	Cluster string                `yaml:"Cluster"`
	Redis   *baseConf.RedisSource `yaml:"RedisSource"`
}

type LiveCallConfig struct {
	Rtc              *Rtc   `yaml:"Rtc"`
	Cache            *Cache `yaml:"Cache"`
	SignalType       int    `yaml:"SignalType"`
	*baseConf.Config `yaml:",inline"`
}
