package multi

type Cache interface {
	Get(string) string
}

type RedisCache struct{}

func (r *RedisCache) Get(k string) string {
	return "redis:" + k
}

// +wire:provider
func NewRedisCache() Cache {
	return &RedisCache{}
}

type MemCache struct{}

func (m *MemCache) Get(k string) string {
	return "mem:" + k
}

func NewMemCache() Cache {
	return &MemCache{}
}

type Service struct {
	cache Cache
}

func NewService(cache Cache) *Service {
	return &Service{cache: cache}
}
