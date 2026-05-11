package skiplist

import (
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomLevel(maxLevel int, p float64) int {
	level := 0
	for rng.Float64() < p && level < maxLevel-1 {
		level++
	}
	return level
}

type baseSkipList struct {
	header   *Node
	maxLevel int
	level    int
	length   int
	P        float64
}

func newBaseSkipList(maxLevel int) *baseSkipList {
	if maxLevel <= 0 {
		maxLevel = DefaultMaxLevel
	}
	header := newNode("", nil, maxLevel-1)
	return &baseSkipList{
		header:   header,
		maxLevel: maxLevel,
		level:    0,
		length:   0,
		P:        DefaultP,
	}
}

func (sl *baseSkipList) findInsertPath(key string, updates []*Node) (*Node, bool) {
	current := sl.header
	for i := sl.level; i >= 0; i-- {
		for current.Next[i] != nil && current.Next[i].Key < key {
			current = current.Next[i]
		}
		updates[i] = current
	}
	current = current.Next[0]
	if current != nil && current.Key == key {
		return current, true
	}
	return nil, false
}

func (sl *baseSkipList) getUpdateArray() []*Node {
	return make([]*Node, sl.maxLevel)
}
