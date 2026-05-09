package lsm

import (
	"math/rand"
	"time"
)

const maxLevel = 16
const p = 0.5

type SkipListNode struct {
	key      string
	value    []byte
	tombstone bool
	forward  []*SkipListNode
}

type SkipList struct {
	head  *SkipListNode
	level int
	size  int
	rng   *rand.Rand
}

func newNode(level int, key string, value []byte, tombstone bool) *SkipListNode {
	return &SkipListNode{
		key:      key,
		value:    value,
		tombstone: tombstone,
		forward:  make([]*SkipListNode, level),
	}
}

func NewSkipList() *SkipList {
	return &SkipList{
		head:  newNode(maxLevel, "", nil, false),
		level: 0,
		size:  0,
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (sl *SkipList) randomLevel() int {
	level := 1
	for sl.rng.Float64() < p && level < maxLevel {
		level++
	}
	return level
}

func (sl *SkipList) Insert(key string, value []byte, tombstone bool) {
	update := make([]*SkipListNode, maxLevel)
	current := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < key {
			current = current.forward[i]
		}
		update[i] = current
	}

	current = current.forward[0]

	if current != nil && current.key == key {
		current.value = value
		current.tombstone = tombstone
		return
	}

	sl.size++
	newLevel := sl.randomLevel()
	if newLevel > sl.level {
		for i := sl.level; i < newLevel; i++ {
			update[i] = sl.head
		}
		sl.level = newLevel
	}

	newNode := newNode(newLevel, key, value, tombstone)
	for i := 0; i < newLevel; i++ {
		newNode.forward[i] = update[i].forward[i]
		update[i].forward[i] = newNode
	}
}

func (sl *SkipList) Get(key string) ([]byte, bool, bool) {
	current := sl.head

	for i := sl.level - 1; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < key {
			current = current.forward[i]
		}
	}

	current = current.forward[0]
	if current != nil && current.key == key {
		return current.value, current.tombstone, true
	}
	return nil, false, false
}

func (sl *SkipList) Size() int {
	return sl.size
}

func (sl *SkipList) Iterator() *SkipListIterator {
	return &SkipListIterator{
		current: sl.head.forward[0],
	}
}

type SkipListIterator struct {
	current *SkipListNode
}

func (it *SkipListIterator) Valid() bool {
	return it.current != nil
}

func (it *SkipListIterator) Key() string {
	return it.current.key
}

func (it *SkipListIterator) Value() []byte {
	return it.current.value
}

func (it *SkipListIterator) Tombstone() bool {
	return it.current.tombstone
}

func (it *SkipListIterator) Next() {
	if it.current != nil {
		it.current = it.current.forward[0]
	}
}
