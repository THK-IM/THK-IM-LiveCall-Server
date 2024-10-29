package cache

import "time"

type OnNewMessage func(msg string)

type RoomCache interface {
	SetEx(key string, value interface{}, expire time.Duration) error
	Expire(key string, expire time.Duration) error
	Get(key string) (value interface{}, err error)
	HSet(key, field string, value interface{}, expire time.Duration) error
	HDel(key, field string) error
	HGet(key, field string) (interface{}, error)
	HValues(key string) ([]string, error)
	HFields(key string) ([]string, error)
	Pub(key, msg string) error
	Sub(key string, onNewMessage OnNewMessage)
	Unsub(key string)
	SAdd(key string, members ...string) (int64, error)
	SMembers(key string) ([]string, error)
	SRem(key string, members ...string) (int64, error)
	Del(key string) error
}
