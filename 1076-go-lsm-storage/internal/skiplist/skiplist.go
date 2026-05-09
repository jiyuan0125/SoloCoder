package skiplist

import (
	"math/rand"
	"time"
)

const (
	maxLevel = 16
	pFactor  = 0.5
)

type Node struct {
	Key       []byte
	Value     []byte
	Tombstone bool
	Next      []*Node
}

type SkipList struct {
	head     *Node
	level    int
	length   int
	rng      *rand.Rand
}

func New() *SkipList {
	return &SkipList{
		head:  &Node{Next: make([]*Node, maxLevel)},
		level: 1,
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *SkipList) randomLevel() int {
	level := 1
	for level < maxLevel && s.rng.Float64() < pFactor {
		level++
	}
	return level
}

func (s *SkipList) Put(key, value []byte, tombstone bool) {
	update := make([]*Node, maxLevel)
	current := s.head

	for i := s.level - 1; i >= 0; i-- {
		for current.Next[i] != nil && compare(current.Next[i].Key, key) < 0 {
			current = current.Next[i]
		}
		update[i] = current
	}

	if current.Next[0] != nil && compare(current.Next[0].Key, key) == 0 {
		current.Next[0].Value = value
		current.Next[0].Tombstone = tombstone
		return
	}

	newLevel := s.randomLevel()
	if newLevel > s.level {
		for i := s.level; i < newLevel; i++ {
			update[i] = s.head
		}
		s.level = newLevel
	}

	newNode := &Node{
		Key:       key,
		Value:     value,
		Tombstone: tombstone,
		Next:      make([]*Node, newLevel),
	}

	for i := 0; i < newLevel; i++ {
		newNode.Next[i] = update[i].Next[i]
		update[i].Next[i] = newNode
	}

	s.length++
}

func (s *SkipList) Get(key []byte) ([]byte, bool, bool) {
	current := s.head
	for i := s.level - 1; i >= 0; i-- {
		for current.Next[i] != nil && compare(current.Next[i].Key, key) < 0 {
			current = current.Next[i]
		}
	}

	if current.Next[0] != nil && compare(current.Next[0].Key, key) == 0 {
		return current.Next[0].Value, current.Next[0].Tombstone, true
	}
	return nil, false, false
}

func (s *SkipList) Delete(key []byte) {
	s.Put(key, nil, true)
}

func (s *SkipList) Length() int {
	return s.length
}

func (s *SkipList) Iterator() *Iterator {
	return &Iterator{current: s.head.Next[0]}
}

type Iterator struct {
	current *Node
}

func (it *Iterator) Valid() bool {
	return it.current != nil
}

func (it *Iterator) Next() {
	it.current = it.current.Next[0]
}

func (it *Iterator) Key() []byte {
	return it.current.Key
}

func (it *Iterator) Value() []byte {
	return it.current.Value
}

func (it *Iterator) Tombstone() bool {
	return it.current.Tombstone
}

func compare(a, b []byte) int {
	la, lb := len(a), len(b)
	minLen := la
	if lb < minLen {
		minLen = lb
	}
	for i := 0; i < minLen; i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	if la < lb {
		return -1
	} else if la > lb {
		return 1
	}
	return 0
}
