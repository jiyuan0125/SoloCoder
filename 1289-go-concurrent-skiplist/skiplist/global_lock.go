package skiplist

import (
	"sync"
	"time"
)

type GlobalLockSkipList struct {
	sl    *baseSkipList
	mu    sync.RWMutex
	stats *SkipListStats
}

func NewGlobalLockSkipList(maxLevel int) *GlobalLockSkipList {
	return &GlobalLockSkipList{
		sl:    newBaseSkipList(maxLevel),
		stats: newSkipListStats(),
	}
}

func (gl *GlobalLockSkipList) Insert(key string, value interface{}) bool {
	start := time.Now()
	gl.mu.Lock()
	waitTime := time.Since(start).Nanoseconds()
	
	success := gl.doInsert(key, value)
	gl.mu.Unlock()
	
	gl.stats.recordOperation(success, waitTime, "insert")
	return success
}

func (gl *GlobalLockSkipList) doInsert(key string, value interface{}) bool {
	updates := gl.sl.getUpdateArray()
	current, found := gl.sl.findInsertPath(key, updates)
	
	if found {
		current.Value = value
		return true
	}
	
	newLevel := randomLevel(gl.sl.maxLevel, gl.sl.P)
	if newLevel > gl.sl.level {
		for i := gl.sl.level + 1; i <= newLevel; i++ {
			updates[i] = gl.sl.header
		}
		gl.sl.level = newLevel
	}
	
	newNode := newNode(key, value, newLevel)
	for i := 0; i <= newLevel; i++ {
		newNode.Next[i] = updates[i].Next[i]
		updates[i].Next[i] = newNode
	}
	
	if newNode.Next[0] != nil {
		newNode.Next[0].Backward = newNode
	}
	newNode.Backward = updates[0]
	
	gl.sl.length++
	return true
}

func (gl *GlobalLockSkipList) Delete(key string) bool {
	start := time.Now()
	gl.mu.Lock()
	waitTime := time.Since(start).Nanoseconds()
	
	success := gl.doDelete(key)
	gl.mu.Unlock()
	
	gl.stats.recordOperation(success, waitTime, "delete")
	return success
}

func (gl *GlobalLockSkipList) doDelete(key string) bool {
	updates := gl.sl.getUpdateArray()
	current, found := gl.sl.findInsertPath(key, updates)
	
	if !found {
		return false
	}
	
	for i := 0; i <= current.Level; i++ {
		if updates[i].Next[i] != current {
			break
		}
		updates[i].Next[i] = current.Next[i]
	}
	
	if current.Next[0] != nil {
		current.Next[0].Backward = updates[0]
	}
	
	for gl.sl.level > 0 && gl.sl.header.Next[gl.sl.level] == nil {
		gl.sl.level--
	}
	
	gl.sl.length--
	return true
}

func (gl *GlobalLockSkipList) Get(key string) (interface{}, bool) {
	start := time.Now()
	gl.mu.RLock()
	waitTime := time.Since(start).Nanoseconds()
	
	value, found := gl.doGet(key)
	gl.mu.RUnlock()
	
	gl.stats.recordOperation(found, waitTime, "get")
	return value, found
}

func (gl *GlobalLockSkipList) doGet(key string) (interface{}, bool) {
	current := gl.sl.header
	for i := gl.sl.level; i >= 0; i-- {
		for current.Next[i] != nil && current.Next[i].Key < key {
			current = current.Next[i]
		}
	}
	current = current.Next[0]
	if current != nil && current.Key == key {
		return current.Value, true
	}
	return nil, false
}

func (gl *GlobalLockSkipList) Range(start, end string) []*KeyValuePair {
	startTime := time.Now()
	gl.mu.RLock()
	waitTime := time.Since(startTime).Nanoseconds()
	
	results := gl.doRange(start, end)
	gl.mu.RUnlock()
	
	gl.stats.recordOperation(true, waitTime, "range")
	return results
}

func (gl *GlobalLockSkipList) doRange(start, end string) []*KeyValuePair {
	var results []*KeyValuePair
	
	current := gl.sl.header.Next[0]
	for current != nil && current.Key < start {
		current = current.Next[0]
	}
	
	for current != nil && current.Key <= end {
		results = append(results, &KeyValuePair{
			Key:   current.Key,
			Value: current.Value,
		})
		current = current.Next[0]
	}
	
	return results
}

func (gl *GlobalLockSkipList) GetStats() SkipListStatsSnapshot {
	return gl.stats.getStats()
}

func (gl *GlobalLockSkipList) ResetStats() {
	gl.stats.reset()
}

func (gl *GlobalLockSkipList) Close() {
}
