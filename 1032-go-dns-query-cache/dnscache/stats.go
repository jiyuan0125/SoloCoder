package dnscache

import (
	"dns-cache-service/common"
	"sync/atomic"
)

type Stats struct {
	totalQueries atomic.Int64
	cacheHits    atomic.Int64
	cacheMisses  atomic.Int64
}

func NewStats() *Stats {
	return &Stats{}
}

func (s *Stats) RecordQuery(hit bool) {
	s.totalQueries.Add(1)
	if hit {
		s.cacheHits.Add(1)
	} else {
		s.cacheMisses.Add(1)
	}
}

func (s *Stats) Reset() {
	s.totalQueries.Store(0)
	s.cacheHits.Store(0)
	s.cacheMisses.Store(0)
}

func (s *Stats) GetStats(cache *Cache) *common.StatsResponse {
	total := s.totalQueries.Load()
	hits := s.cacheHits.Load()
	misses := s.cacheMisses.Load()

	hitRate := 0.0
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	return &common.StatsResponse{
		TotalEntries: cache.Count(),
		TotalQueries: total,
		CacheHits:    hits,
		CacheMisses:  misses,
		HitRate:      hitRate,
	}
}
