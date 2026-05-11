package shardedmap

import (
	"sync"
	"sync/atomic"
)

const (
	DefaultShardCount   = 16
	MinShardCount       = 1
	MaxShardCount       = 1024
	DefaultShardMaxSize = 256
	DefaultTotalMinSize = 64
	UnevenFactor        = 3
)

type ShardInfo struct {
	ShardID  int
	Count    int
	IsUneven bool
}

type ShardedMap struct {
	shards       atomic.Value
	shardCount   int
	shardMaxSize int
	totalMinSize int
	resizeMu     sync.Mutex
}

type shardArray []*shard

func NewShardedMap(initialShardCount int) *ShardedMap {
	return NewShardedMapWithParams(initialShardCount, DefaultShardMaxSize, DefaultTotalMinSize)
}

func NewShardedMapWithParams(initialShardCount, shardMaxSize, totalMinSize int) *ShardedMap {
	if initialShardCount < MinShardCount {
		initialShardCount = MinShardCount
	}
	if initialShardCount > MaxShardCount {
		initialShardCount = MaxShardCount
	}
	shardCount := nextPowerOfTwo(initialShardCount)
	shards := make(shardArray, shardCount)
	for i := range shards {
		shards[i] = newShard()
	}
	sm := &ShardedMap{
		shardCount:   shardCount,
		shardMaxSize: shardMaxSize,
		totalMinSize: totalMinSize,
	}
	sm.shards.Store(&shards)
	return sm
}

func (sm *ShardedMap) getShards() shardArray {
	return *sm.shards.Load().(*shardArray)
}

func (sm *ShardedMap) getShardIndex(key string, shardCount int) int {
	hash := fnv1aHash(key)
	return int(hash & uint64(shardCount-1))
}

func (sm *ShardedMap) Put(key, value string) {
	for {
		shards := sm.getShards()
		shardCount := len(shards)
		idx := sm.getShardIndex(key, shardCount)
		shards[idx].put(key, value)

		if shards[idx].len() > sm.shardMaxSize {
			if sm.tryExpand(shardCount) {
				continue
			}
		}
		return
	}
}

func (sm *ShardedMap) Get(key string) (string, bool) {
	shards := sm.getShards()
	shardCount := len(shards)
	idx := sm.getShardIndex(key, shardCount)
	return shards[idx].get(key)
}

func (sm *ShardedMap) Delete(key string) {
	shards := sm.getShards()
	shardCount := len(shards)
	idx := sm.getShardIndex(key, shardCount)
	shards[idx].delete(key)

	total := sm.Len()
	if total < sm.totalMinSize && shardCount > MinShardCount {
		sm.tryShrink(shardCount)
	}
}

func (sm *ShardedMap) Len() int {
	shards := sm.getShards()
	total := 0
	for _, s := range shards {
		total += s.len()
	}
	return total
}

func (sm *ShardedMap) Keys() []string {
	shards := sm.getShards()
	total := 0
	for _, s := range shards {
		total += s.len()
	}
	keys := make([]string, 0, total)
	for _, s := range shards {
		keys = append(keys, s.keys()...)
	}
	return keys
}

func (sm *ShardedMap) ShardCount() int {
	return len(sm.getShards())
}

func (sm *ShardedMap) Stats() []ShardInfo {
	shards := sm.getShards()
	shardCount := len(shards)
	total := 0
	counts := make([]int, shardCount)
	for i, s := range shards {
		c := s.len()
		counts[i] = c
		total += c
	}

	avg := 0
	if shardCount > 0 {
		avg = total / shardCount
	}

	stats := make([]ShardInfo, shardCount)
	for i, c := range counts {
		stats[i] = ShardInfo{
			ShardID:  i,
			Count:    c,
			IsUneven: avg > 0 && c > avg*UnevenFactor,
		}
	}
	return stats
}

func (sm *ShardedMap) tryExpand(currentCount int) bool {
	newCount := currentCount * 2
	if newCount > MaxShardCount {
		return false
	}
	return sm.resize(newCount)
}

func (sm *ShardedMap) tryShrink(currentCount int) bool {
	newCount := currentCount / 2
	if newCount < MinShardCount {
		return false
	}
	return sm.resize(newCount)
}

func (sm *ShardedMap) resize(newCount int) bool {
	sm.resizeMu.Lock()
	defer sm.resizeMu.Unlock()

	oldShards := sm.getShards()
	oldCount := len(oldShards)

	if oldCount == newCount {
		return false
	}

	for _, s := range oldShards {
		s.mu.Lock()
		defer s.mu.Unlock()
	}

	newShards := make(shardArray, newCount)
	for i := range newShards {
		newShards[i] = newShard()
	}

	for _, oldShard := range oldShards {
		for k, v := range oldShard.m {
			idx := sm.getShardIndex(k, newCount)
			newShards[idx].m[k] = v
		}
	}

	sm.shards.Store(&newShards)
	return true
}
