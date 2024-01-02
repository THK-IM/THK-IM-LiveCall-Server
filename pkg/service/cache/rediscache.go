package cache

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

type RedisCache struct {
	client    *redis.Client
	logger    *logrus.Entry
	rwMutex   *sync.RWMutex
	pubSubMap map[string]*redis.PubSub
}

func (r *RedisCache) Expire(key string, expire time.Duration) error {
	ctx := context.Background()
	return r.client.Expire(ctx, key, expire).Err()
}

func (r *RedisCache) SetEx(key string, value interface{}, expire time.Duration) error {
	ctx := context.Background()
	return r.client.SetEx(ctx, key, value, expire).Err()
}

func (r *RedisCache) Get(key string) (value interface{}, err error) {
	ctx := context.Background()
	cmd := r.client.Get(ctx, key)
	return cmd.Result()
}

func (r *RedisCache) HSet(key, field string, value interface{}, expire time.Duration) error {
	ctx := context.Background()
	return r.client.HSet(ctx, key, field, value).Err()
}

func (r *RedisCache) HGet(key, field string) (interface{}, error) {
	ctx := context.Background()
	cmd := r.client.HGet(ctx, key, field)
	return cmd.Result()
}

func (r *RedisCache) HDel(key, field string) error {
	ctx := context.Background()
	return r.client.HDel(ctx, key, field).Err()
}

func (r *RedisCache) HValues(key string) ([]string, error) {
	ctx := context.Background()
	cmd := r.client.HVals(ctx, key)
	return cmd.Result()
}

func (r *RedisCache) HFields(key string) ([]string, error) {
	ctx := context.Background()
	cmd := r.client.HKeys(ctx, key)
	return cmd.Result()
}

func (r *RedisCache) Pub(key, msg string) error {
	r.rwMutex.RLock()
	defer r.rwMutex.RUnlock()
	ctx := context.Background()
	return r.client.Publish(ctx, key, msg).Err()
}

func (r *RedisCache) Sub(key string, onNewMessage OnNewMessage) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	ctx := context.Background()
	pubSub := r.pubSubMap[key]
	if pubSub == nil {
		pubSub = r.client.Subscribe(ctx, key)
		r.pubSubMap[key] = pubSub
	}
	go func() {
		for {
			select {
			case msg, ok := <-(pubSub.Channel()):
				if !ok {
					return
				} else {
					onNewMessage(msg.Payload)
				}
			}
		}
	}()
}

func (r *RedisCache) Unsub(key string) {
	r.rwMutex.Lock()
	defer r.rwMutex.Unlock()
	pubSub := r.pubSubMap[key]
	if pubSub != nil {
		if err := pubSub.Unsubscribe(context.Background(), key); err == nil {
			delete(r.pubSubMap, key)
		}
	}
}

func (r *RedisCache) SAdd(key string, members ...string) (int64, error) {
	ctx := context.Background()
	return r.client.SAdd(ctx, key, members).Result()
}

func (r *RedisCache) SMembers(key string) ([]string, error) {
	ctx := context.Background()
	return r.client.SMembers(ctx, key).Result()
}

func (r *RedisCache) SRem(key string, members ...string) (int64, error) {
	ctx := context.Background()
	return r.client.SRem(ctx, key, members).Result()
}

func (r *RedisCache) Del(key string) error {
	ctx := context.Background()
	return r.client.Del(ctx, key).Err()
}

func MakeRedisCache(client *redis.Client, logger *logrus.Entry) Service {
	return &RedisCache{
		client:    client,
		logger:    logger.WithField("search_index", "redis_cache"),
		rwMutex:   &sync.RWMutex{},
		pubSubMap: make(map[string]*redis.PubSub),
	}
}
