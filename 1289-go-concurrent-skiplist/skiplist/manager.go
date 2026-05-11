package skiplist

import (
	"sync"
)

type SkipListManager struct {
	mu        sync.RWMutex
	skipList  ConcurrentSkipList
	strategy  ConcurrentStrategy
	maxLevel  int
}

func NewSkipListManager(strategy ConcurrentStrategy, maxLevel int) *SkipListManager {
	if maxLevel <= 0 {
		maxLevel = DefaultMaxLevel
	}
	
	var sl ConcurrentSkipList
	switch strategy {
	case StrategyNodeLevelLock:
		sl = NewNodeLevelLockSkipList(maxLevel)
	default:
		sl = NewGlobalLockSkipList(maxLevel)
	}
	
	return &SkipListManager{
		skipList: sl,
		strategy: strategy,
		maxLevel: maxLevel,
	}
}

func (m *SkipListManager) Insert(key string, value interface{}) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.skipList.Insert(key, value)
}

func (m *SkipListManager) Delete(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.skipList.Delete(key)
}

func (m *SkipListManager) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.skipList.Get(key)
}

func (m *SkipListManager) Range(start, end string) []*KeyValuePair {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.skipList.Range(start, end)
}

func (m *SkipListManager) GetStats() SkipListStatsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.skipList.GetStats()
}

func (m *SkipListManager) ResetStats() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	m.skipList.ResetStats()
}

func (m *SkipListManager) GetStrategy() ConcurrentStrategy {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.strategy
}

func (m *SkipListManager) SetStrategy(strategy ConcurrentStrategy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if strategy == m.strategy {
		return
	}
	
	data := m.skipList.Range("", "\xff")
	m.skipList.Close()
	
	var newSkipList ConcurrentSkipList
	switch strategy {
	case StrategyNodeLevelLock:
		newSkipList = NewNodeLevelLockSkipList(m.maxLevel)
	default:
		newSkipList = NewGlobalLockSkipList(m.maxLevel)
	}
	
	for _, kv := range data {
		newSkipList.Insert(kv.Key, kv.Value)
	}
	
	m.skipList = newSkipList
	m.strategy = strategy
}

func (m *SkipListManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.skipList != nil {
		m.skipList.Close()
		m.skipList = nil
	}
}
