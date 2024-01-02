package cache

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/zoumo/goset"
	"sync"
	"time"
)

type LocalCache struct {
	logger    *logrus.Entry
	data      map[string]interface{}
	keyExpire map[string]*time.Time
	rwMutex   *sync.RWMutex
}

func (l *LocalCache) SetEx(key string, value interface{}, expire time.Duration) error {
	l.rwMutex.Lock()
	defer l.rwMutex.Unlock()
	l.data[key] = value
	expireTime := time.Now().Add(expire)
	l.keyExpire[key] = &expireTime
	return nil
}

func (l *LocalCache) Expire(key string, expire time.Duration) error {
	l.rwMutex.Lock()
	defer l.rwMutex.Unlock()
	expireTime := time.Now().Add(expire)
	l.keyExpire[key] = &expireTime
	return nil
}

func (l *LocalCache) Get(key string) (value interface{}, err error) {
	l.rwMutex.RLock()
	defer l.rwMutex.RUnlock()
	expireTime := l.keyExpire[key]
	if expireTime == nil {
		return l.data[key], nil
	} else if expireTime.After(time.Now()) {
		return l.data[key], nil
	} else {
		delete(l.data, key)
		return nil, nil
	}
}

func (l *LocalCache) HSet(key, field string, value interface{}, expire time.Duration) error {
	l.rwMutex.Lock()
	defer l.rwMutex.Unlock()
	data := l.data[key]
	if data == nil {
		dataMap := make(map[string]interface{})
		dataMap[field] = value
		l.data[key] = dataMap
		expireTime := time.Now().Add(expire)
		l.keyExpire[key] = &expireTime
		return nil
	}
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return errors.New("type err")
	} else {
		dataMap[field] = value
		expireTime := time.Now().Add(expire)
		l.keyExpire[key] = &expireTime
		return nil
	}
}

func (l *LocalCache) HGet(key, field string) (interface{}, error) {
	l.rwMutex.Lock()
	defer l.rwMutex.Unlock()
	data := l.data[key]
	expireTime := l.keyExpire[key]
	if expireTime == nil || expireTime.After(time.Now()) {
		if data == nil {
			return nil, nil
		} else {
			dataMap, ok := data.(map[string]interface{})
			if !ok {
				return nil, nil
			} else {
				return dataMap[field], nil
			}
		}
	} else {
		delete(l.data, key)
		return nil, nil
	}
}

func (l *LocalCache) HDel(key, field string) error {
	l.rwMutex.Lock()
	defer l.rwMutex.Unlock()
	data := l.data[key]
	if data == nil {
		return nil
	} else {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return nil
		} else {
			delete(dataMap, field)
		}
	}
	return nil
}

func (l *LocalCache) HValues(key string) ([]string, error) {
	l.rwMutex.RLock()
	defer l.rwMutex.RUnlock()
	data := l.data[key]
	expireTime := l.keyExpire[key]
	if expireTime == nil || expireTime.After(time.Now()) {
		if data == nil {
			return nil, nil
		} else {
			dataMap, ok := data.(map[string]interface{})
			if !ok {
				return nil, nil
			} else {
				values := make([]string, 0)
				for _, v := range dataMap {
					val, _ := v.(string)
					values = append(values, val)
				}
				return values, nil
			}
		}
	} else {
		delete(l.data, key)
		return nil, nil
	}
}

func (l *LocalCache) HFields(key string) ([]string, error) {
	l.rwMutex.RLock()
	defer l.rwMutex.RUnlock()
	data := l.data[key]
	expireTime := l.keyExpire[key]
	if expireTime == nil || expireTime.After(time.Now()) {
		if data == nil {
			return nil, nil
		} else {
			dataMap, ok := data.(map[string]interface{})
			if !ok {
				return nil, nil
			} else {
				keys := make([]string, 0)
				for k, _ := range dataMap {
					keys = append(keys, k)
				}
				return keys, nil
			}
		}
	} else {
		delete(l.data, key)
		return nil, nil
	}
}

func (l *LocalCache) Pub(key, msg string) error {
	l.rwMutex.RLock()
	defer l.rwMutex.RUnlock()
	var msgChannel chan string
	if l.data[key] == nil {
		msgChannel = make(chan string)
		l.data[key] = msgChannel
	} else {
		msgChannel = l.data[key].(chan string)
	}
	msgChannel <- msg
	return nil
}

func (l *LocalCache) Sub(key string, onNewMessage OnNewMessage) {
	l.rwMutex.Lock()
	defer l.rwMutex.Unlock()
	var msgChannel chan string
	if l.data[key] == nil {
		msgChannel = make(chan string)
		l.data[key] = msgChannel
	} else {
		msgChannel = l.data[key].(chan string)
	}
	go func() {
		for {
			select {
			case msg, ok := <-msgChannel:
				if !ok {
					return
				} else {
					onNewMessage(msg)
				}
			}
		}
	}()
}

func (l *LocalCache) Unsub(key string) {
	l.rwMutex.Lock()
	defer l.rwMutex.Unlock()
	if l.data[key] != nil {
		if msgChannel, ok := l.data[key].(chan string); ok {
			close(msgChannel)
			delete(l.data, key)
		}
	}
}

func (l *LocalCache) SAdd(key string, members ...string) (int64, error) {
	l.rwMutex.Lock()
	defer l.rwMutex.Unlock()
	if l.data[key] == nil {
		l.data[key] = goset.NewSetFrom(members)
		return int64(len(members)), nil
	} else {
		if sets, ok := l.data[key].(goset.Set); ok {
			err := sets.Extend(members)
			return int64(len(members)), err
		} else {
			return 0, errors.New("key error")
		}
	}
}

func (l *LocalCache) SMembers(key string) ([]string, error) {
	l.rwMutex.RLock()
	defer l.rwMutex.RUnlock()
	if l.data[key] == nil {
		return nil, nil
	} else {
		if sets, ok := l.data[key].(goset.Set); ok {
			elements := sets.Elements()
			results := make([]string, 0)
			for _, e := range elements {
				if r, o := e.(string); o {
					results = append(results, r)
				}
			}
			return results, nil
		} else {
			return nil, errors.New("key error")
		}
	}
}

func (l *LocalCache) SRem(key string, members ...string) (int64, error) {
	l.rwMutex.Lock()
	defer l.rwMutex.Unlock()
	if l.data[key] == nil {
		return 0, errors.New("key error")
	} else {
		if sets, ok := l.data[key].(goset.Set); ok {
			c := 0
			for _, e := range members {
				if sets.Contains(e) {
					c++
					sets.Remove(e)
				}
			}
			return int64(c), nil
		} else {
			return 0, errors.New("key error")
		}
	}
}

func (l *LocalCache) Del(key string) error {
	l.rwMutex.Lock()
	defer l.rwMutex.Unlock()
	delete(l.data, key)
	return nil
}

func MakeLocalCache(logger *logrus.Entry, mode string) Service {
	cache := &LocalCache{
		logger:    logger.WithField("search_index", "local_cache"),
		data:      make(map[string]interface{}),
		rwMutex:   &sync.RWMutex{},
		keyExpire: make(map[string]*time.Time),
	}
	return cache
}
