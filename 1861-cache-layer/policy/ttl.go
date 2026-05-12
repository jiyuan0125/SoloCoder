package policy

import "time"

type TTLCache interface {
	Put(key string, ttl time.Duration)
	Remove(key string)
	ExpiredAt(key string) (time.Time, bool)
	IsExpired(key string) bool
}

type ttlCache struct {
	expireMap map[string]time.Time
}

func NewTTLCache() TTLCache {
	return &ttlCache{
		expireMap: make(map[string]time.Time),
	}
}

func (t *ttlCache) Put(key string, ttl time.Duration) {
	if ttl <= 0 {
		delete(t.expireMap, key)
		return
	}
	t.expireMap[key] = time.Now().Add(ttl)
}

func (t *ttlCache) Remove(key string) {
	delete(t.expireMap, key)
}

func (t *ttlCache) ExpiredAt(key string) (time.Time, bool) {
	exp, ok := t.expireMap[key]
	return exp, ok
}

func (t *ttlCache) IsExpired(key string) bool {
	exp, ok := t.expireMap[key]
	if !ok {
		return false
	}
	return time.Now().After(exp)
}
