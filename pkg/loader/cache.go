package loader

import (
	"errors"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	baseConf "github.com/thk-im/thk-im-base-server/conf"
	baseLoader "github.com/thk-im/thk-im-base-server/loader"
	"github.com/thk-im/thk-im-livecall-server/pkg/conf"
	"github.com/thk-im/thk-im-livecall-server/pkg/service/cache"
)

func LoadCacheService(source *conf.Cache, logger *logrus.Entry) cache.Service {
	if source.Cluster == "Standalone" {
		return cache.MakeLocalCache(logger)
	} else if source.Cluster == "Cluster" {
		redisClient := loadRedis(source.Redis)
		return cache.MakeRedisCache(redisClient, logger)
	} else {
		panic(errors.New("cache config err: " + source.Cluster))
	}
}

func loadRedis(source *baseConf.RedisSource) *redis.Client {
	return baseLoader.LoadRedis(source)
}
