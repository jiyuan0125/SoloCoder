package skiplist

import (
	"sync/atomic"
	"time"
)

type NodeLevelLockSkipList struct {
	header   *Node
	maxLevel int
	level    int32
	length   int32
	P        float64
	stats    *SkipListStats
}

func NewNodeLevelLockSkipList(maxLevel int) *NodeLevelLockSkipList {
	if maxLevel <= 0 {
		maxLevel = DefaultMaxLevel
	}
	header := newNode("", nil, maxLevel-1)
	return &NodeLevelLockSkipList{
		header:   header,
		maxLevel: maxLevel,
		level:    0,
		length:   0,
		P:        DefaultP,
		stats:    newSkipListStats(),
	}
}

func (nll *NodeLevelLockSkipList) getCurrentLevel() int {
	return int(atomic.LoadInt32(&nll.level))
}

func (nll *NodeLevelLockSkipList) setCurrentLevel(level int) {
	atomic.StoreInt32(&nll.level, int32(level))
}

func (nll *NodeLevelLockSkipList) Insert(key string, value interface{}) bool {
	start := time.Now()
	success := nll.doInsert(key, value)
	waitTime := time.Since(start).Nanoseconds()
	nll.stats.recordOperation(success, waitTime, "insert")
	return success
}

func (nll *NodeLevelLockSkipList) doInsert(key string, value interface{}) bool {
	for {
		updates := make([]*Node, nll.maxLevel)
		currentLevel := nll.getCurrentLevel()
		current := nll.header

		for i := currentLevel; i >= 0; i-- {
			for current.Next[i] != nil && current.Next[i].Key < key {
				current = current.Next[i]
			}
			updates[i] = current
		}

		next := current.Next[0]

		if next != nil && next.Key == key {
			next.mu.Lock()
			if next.isDeleted() {
				next.mu.Unlock()
				continue
			}
			next.Value = value
			next.mu.Unlock()
			return true
		}

		lockedNodes := make(map[*Node]bool)
		allLocked := true

		for i := 0; i <= currentLevel; i++ {
			node := updates[i]
			if node == nil {
				continue
			}
			if !lockedNodes[node] {
				node.mu.Lock()
				lockedNodes[node] = true
			}

			if node.Next[i] != nil && node.Next[i].Key < key {
				allLocked = false
				break
			}

			nextAfterLock := node.Next[0]
			if i == 0 && nextAfterLock != nil && nextAfterLock.Key == key {
				nextAfterLock.mu.Lock()
				if nextAfterLock.isDeleted() {
					nextAfterLock.mu.Unlock()
					for n := range lockedNodes {
						n.mu.Unlock()
					}
					continue
				}
				nextAfterLock.Value = value
				nextAfterLock.mu.Unlock()
				for n := range lockedNodes {
					n.mu.Unlock()
				}
				return true
			}

			if i == 0 && nextAfterLock != nil && nextAfterLock.Key < key {
				allLocked = false
				break
			}
		}

		if !allLocked {
			for n := range lockedNodes {
				n.mu.Unlock()
			}
			continue
		}

		newLevel := randomLevel(nll.maxLevel, nll.P)
		if newLevel > currentLevel {
			for i := currentLevel + 1; i <= newLevel; i++ {
				updates[i] = nll.header
				if !lockedNodes[nll.header] {
					nll.header.mu.Lock()
					lockedNodes[nll.header] = true
				}
			}
			nll.setCurrentLevel(newLevel)
		}

		newNode := newNode(key, value, newLevel)
		newNode.mu.Lock()

		for i := 0; i <= newLevel; i++ {
			newNode.Next[i] = updates[i].Next[i]
			updates[i].Next[i] = newNode
		}

		if newNode.Next[0] != nil {
			newNode.Next[0].Backward = newNode
		}
		newNode.Backward = updates[0]

		newNode.mu.Unlock()

		for n := range lockedNodes {
			n.mu.Unlock()
		}

		atomic.AddInt32(&nll.length, 1)
		return true
	}
}

func (nll *NodeLevelLockSkipList) Delete(key string) bool {
	start := time.Now()
	success := nll.doDelete(key)
	waitTime := time.Since(start).Nanoseconds()
	nll.stats.recordOperation(success, waitTime, "delete")
	return success
}

func (nll *NodeLevelLockSkipList) doDelete(key string) bool {
	for {
		updates := make([]*Node, nll.maxLevel)
		currentLevel := nll.getCurrentLevel()
		current := nll.header

		for i := currentLevel; i >= 0; i-- {
			for current.Next[i] != nil && current.Next[i].Key < key {
				current = current.Next[i]
			}
			updates[i] = current
		}

		current = current.Next[0]

		if current == nil || current.Key != key {
			return false
		}

		current.mu.Lock()
		if current.isDeleted() {
			current.mu.Unlock()
			continue
		}

		if current.Level > currentLevel {
			current.mu.Unlock()
			continue
		}

		lockedNodes := make(map[*Node]bool)
		lockedNodes[current] = true

		validationPassed := true
		for i := 0; i <= current.Level; i++ {
			node := updates[i]
			if node == nil {
				continue
			}
			if !lockedNodes[node] {
				node.mu.Lock()
				lockedNodes[node] = true
			}

			if node.Next[i] != current {
				validationPassed = false
				break
			}
		}

		if !validationPassed {
			for n := range lockedNodes {
				n.mu.Unlock()
			}
			nll.stats.recordConflict()
			continue
		}

		current.markDeleted()

		for i := 0; i <= current.Level; i++ {
			updates[i].Next[i] = current.Next[i]
		}

		if current.Next[0] != nil {
			current.Next[0].Backward = updates[0]
		}

		for n := range lockedNodes {
			n.mu.Unlock()
		}

		for nll.getCurrentLevel() > 0 && nll.header.Next[nll.getCurrentLevel()] == nil {
			nll.setCurrentLevel(nll.getCurrentLevel() - 1)
		}

		atomic.AddInt32(&nll.length, -1)
		return true
	}
}

func (nll *NodeLevelLockSkipList) Get(key string) (interface{}, bool) {
	start := time.Now()
	value, found := nll.doGet(key)
	waitTime := time.Since(start).Nanoseconds()
	nll.stats.recordOperation(found, waitTime, "get")
	return value, found
}

func (nll *NodeLevelLockSkipList) doGet(key string) (interface{}, bool) {
	for {
		current := nll.header
		currentLevel := nll.getCurrentLevel()

		for i := currentLevel; i >= 0; i-- {
			for current.Next[i] != nil && current.Next[i].Key < key {
				current = current.Next[i]
			}
		}

		current = current.Next[0]
		if current == nil || current.Key != key {
			return nil, false
		}

		current.mu.Lock()
		if current.isDeleted() {
			current.mu.Unlock()
			continue
		}
		value := current.Value
		current.mu.Unlock()
		return value, true
	}
}

func (nll *NodeLevelLockSkipList) Range(start, end string) []*KeyValuePair {
	startTime := time.Now()
	results := nll.doRange(start, end)
	waitTime := time.Since(startTime).Nanoseconds()
	nll.stats.recordOperation(true, waitTime, "range")
	return results
}

func (nll *NodeLevelLockSkipList) doRange(start, end string) []*KeyValuePair {
	var results []*KeyValuePair

	current := nll.header.Next[0]
	for current != nil && current.Key < start {
		current = current.Next[0]
	}

	for current != nil && current.Key <= end {
		current.mu.Lock()
		if !current.isDeleted() {
			results = append(results, &KeyValuePair{
				Key:   current.Key,
				Value: current.Value,
			})
		}
		next := current.Next[0]
		current.mu.Unlock()
		current = next
	}

	return results
}

func (nll *NodeLevelLockSkipList) GetStats() SkipListStatsSnapshot {
	return nll.stats.getStats()
}

func (nll *NodeLevelLockSkipList) ResetStats() {
	nll.stats.reset()
}

func (nll *NodeLevelLockSkipList) Close() {
}
