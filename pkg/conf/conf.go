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

type ClickHouse struct {
	Url             string `yaml:"Url"`
	Db              string `yaml:"Db"`
	UserName        string `yaml:"Username"`
	Password        string `yaml:"Password"`
	MaxIdleConn     int    `yaml:"MaxIdleConn"`
	MaxOpenConn     int    `yaml:"MaxOpenConn"`
	Timeout         int64  `yaml:"Timeout"`         // 单位:秒
	ConnMaxLifeTime int64  `yaml:"ConnMaxLifeTime"` // 单位:秒
	ConnMaxIdleTime int64  `yaml:"ConnMaxIdleTime"` // 单位:秒
}

type Cache struct {
	Cluster string                `yaml:"Cluster"`
	Redis   *baseConf.RedisSource `yaml:"RedisSource"`
}

type Stat struct {
	ClickHouse *ClickHouse `yaml:"ClickHouse"`
}

type LiveCallConfig struct {
	Rtc              *Rtc   `yaml:"Rtc"`
	Cache            *Cache `yaml:"Cache"`
	Stat             *Stat  `yaml:"Stat"`
	*baseConf.Config `yaml:",inline"`
}
