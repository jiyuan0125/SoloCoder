package skiplist

import (
	"math/rand"
	"sync"
	"time"
)

type SkipList struct {
	header *node
	level  int
	length int
	mu     sync.RWMutex
	rand   *rand.Rand
}

func New() *SkipList {
	return &SkipList{
		header: newNode("", nil, maxLevel),
		level:  0,
		length: 0,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (sl *SkipList) randomLevel() int {
	level := 0
	for sl.rand.Float64() < probability && level < maxLevel-1 {
		level++
	}
	return level
}

func (sl *SkipList) Put(key string, value []byte) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	update := make([]*node, maxLevel)
	current := sl.header

	for i := sl.level; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < key {
			current = current.forward[i]
		}
		update[i] = current
	}

	current = current.forward[0]

	if current != nil && current.key == key {
		current.value = value
		return
	}

	newLevel := sl.randomLevel()

	if newLevel > sl.level {
		for i := sl.level + 1; i <= newLevel; i++ {
			update[i] = sl.header
		}
		sl.level = newLevel
	}

	newNode := newNode(key, value, newLevel)

	for i := 0; i <= newLevel; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}

	sl.length++
}

func (sl *SkipList) Get(key string) ([]byte, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	current := sl.header

	for i := sl.level; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < key {
			current = current.forward[i]
		}
	}

	current = current.forward[0]

	if current != nil && current.key == key {
		return current.value, true
	}

	return nil, false
}

func (sl *SkipList) Delete(key string) bool {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	update := make([]*node, maxLevel)
	current := sl.header

	for i := sl.level; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < key {
			current = current.forward[i]
		}
		update[i] = current
	}

	current = current.forward[0]

	if current == nil || current.key != key {
		return false
	}

	for i := 0; i <= sl.level; i++ {
		if update[i].forward[i] != current {
			break
		}
		update[i].forward[i] = current.forward[i]
	}

	for sl.level > 0 && sl.header.forward[sl.level] == nil {
		sl.level--
	}

	sl.length--
	return true
}

func (sl *SkipList) Range(from, to string) []*KVPair {
	if from > to {
		return []*KVPair{}
	}

	sl.mu.RLock()
	defer sl.mu.RUnlock()

	var result []*KVPair

	startNode := sl.findFirstGreaterOrEqual(from)
	if startNode == nil {
		return []*KVPair{}
	}

	current := startNode
	for current != nil && current.key <= to {
		result = append(result, &KVPair{
			Key:   current.key,
			Value: current.value,
		})
		current = current.forward[0]
	}

	return result
}

func (sl *SkipList) findFirstGreaterOrEqual(key string) *node {
	current := sl.header

	for i := sl.level; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < key {
			current = current.forward[i]
		}
	}

	return current.forward[0]
}

func (sl *SkipList) Rank(key string) (int, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()

	current := sl.header
	rank := 0

	for i := sl.level; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < key {
			rank += sl.countInLevel(current, i, key)
			current = current.forward[i]
		}
	}

	current = current.forward[0]

	if current != nil && current.key == key {
		return rank + 1, true
	}

	return 0, false
}

func (sl *SkipList) countInLevel(start *node, level int, key string) int {
	count := 0
	current := start.forward[level]

	for current != nil && current.key < key {
		count++
		current = current.forward[level]
	}

	return count
}

func (sl *SkipList) Length() int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.length
}

type KVPair struct {
	Key   string
	Value []byte
}
