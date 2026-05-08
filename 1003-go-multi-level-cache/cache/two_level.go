package cache

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/singleflight"
)

type Stats struct {
	MemoryHits   int64
	MemoryMisses int64
	FileHits     int64
	FileMisses   int64
}

type TwoLevelCache struct {
	memory *MemoryCache
	file   *FileCache
	group  singleflight.Group
	stats  Stats
}

type TwoLevelCacheConfig struct {
	MemoryCapacity  int
	FileCacheDir    string
	CleanInterval   time.Duration
}

func NewTwoLevelCache(config TwoLevelCacheConfig) (*TwoLevelCache, error) {
	if config.MemoryCapacity <= 0 {
		config.MemoryCapacity = 1000
	}
	if config.CleanInterval <= 0 {
		config.CleanInterval = 60 * time.Second
	}
	if config.FileCacheDir == "" {
		config.FileCacheDir = "./file_cache"
	}

	tc := &TwoLevelCache{
		memory: NewMemoryCache(config.MemoryCapacity),
		file:   NewFileCache(config.FileCacheDir, config.CleanInterval),
	}

	if err := tc.file.Start(); err != nil {
		return nil, err
	}

	return tc, nil
}

func (tc *TwoLevelCache) Stop() {
	tc.file.Stop()
}

func (tc *TwoLevelCache) Get(key string) (interface{}, time.Time, bool) {
	value, expire, found := tc.memory.Get(key)
	if found {
		atomic.AddInt64(&tc.stats.MemoryHits, 1)
		return value, expire, true
	}
	atomic.AddInt64(&tc.stats.MemoryMisses, 1)

	result, err, _ := tc.group.Do(key, func() (interface{}, error) {
		return tc.loadFromFile(key)
	})

	if err != nil {
		return nil, time.Time{}, false
	}

	loadResult := result.(*loadFromFileResult)
	if !loadResult.found {
		return nil, time.Time{}, false
	}

	tc.memory.Set(key, loadResult.value, 0)
	return loadResult.value, loadResult.expire, true
}

type loadFromFileResult struct {
	value  interface{}
	expire time.Time
	found  bool
}

func (tc *TwoLevelCache) loadFromFile(key string) (*loadFromFileResult, error) {
	raw, expire, found := tc.file.Get(key)
	if !found {
		atomic.AddInt64(&tc.stats.FileMisses, 1)
		return &loadFromFileResult{found: false}, nil
	}
	atomic.AddInt64(&tc.stats.FileHits, 1)

	var value interface{}
	if err := json.Unmarshal(raw, &value); err != nil {
		tc.file.Delete(key)
		return &loadFromFileResult{found: false}, nil
	}

	return &loadFromFileResult{
		value:  value,
		expire: expire,
		found:  true,
	}, nil
}

func (tc *TwoLevelCache) Set(key string, value interface{}, memoryTTL, fileTTL time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := tc.file.Set(key, raw, fileTTL); err != nil {
		return err
	}

	tc.memory.Set(key, value, memoryTTL)
	return nil
}

func (tc *TwoLevelCache) Delete(key string) bool {
	fileDeleted := tc.file.Delete(key)
	memDeleted := tc.memory.Delete(key)
	return fileDeleted || memDeleted
}

func (tc *TwoLevelCache) DeletePrefix(prefix string) int {
	fileCount := tc.file.DeletePrefix(prefix)
	memCount := tc.memory.DeletePrefix(prefix)
	if fileCount > memCount {
		return fileCount
	}
	return memCount
}

func (tc *TwoLevelCache) Clear() {
	tc.file.Clear()
	tc.memory.Clear()
}

func (tc *TwoLevelCache) Stats() Stats {
	return Stats{
		MemoryHits:   atomic.LoadInt64(&tc.stats.MemoryHits),
		MemoryMisses: atomic.LoadInt64(&tc.stats.MemoryMisses),
		FileHits:     atomic.LoadInt64(&tc.stats.FileHits),
		FileMisses:   atomic.LoadInt64(&tc.stats.FileMisses),
	}
}

func (tc *TwoLevelCache) MemoryCount() int {
	return tc.memory.Count()
}

func (tc *TwoLevelCache) FileCount() int {
	return tc.file.Count()
}

func (tc *TwoLevelCache) MemoryCapacity() int {
	return tc.memory.Capacity()
}

func (tc *TwoLevelCache) MemoryUsage() float64 {
	return tc.memory.Usage()
}

func (tc *TwoLevelCache) WarmUp(maxMemoryItems int) {
	if maxMemoryItems <= 0 {
		maxMemoryItems = tc.memory.Capacity()
	}

	warmed := 0
	tc.file.WarmUp(tc.memory, func(key string, value json.RawMessage) {
		if warmed >= maxMemoryItems {
			return
		}

		var obj interface{}
		if err := json.Unmarshal(value, &obj); err != nil {
			return
		}

		tc.memory.Set(key, obj, 0)
		warmed++
	})
}

var _ = sync.Mutex{}
