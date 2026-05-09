package skipmap

import (
	"math/rand"
	"sync/atomic"
	"unsafe"
)

const (
	maxLevel    = 32
	p           = 0.25
	pThreshold  = uint32(p * (1 << 32))
)

type node struct {
	key       string
	value     atomic.Pointer[string]
	deleted   atomic.Bool
	next      [maxLevel]atomic.Pointer[node]
	level     int
}

type SkipMap struct {
	head    *node
	tail    *node
	level   atomic.Int32
	size    atomic.Int64
}

type Iterator struct {
	sm    *SkipMap
	curr  *node
	end   string
	hasEnd bool
}

func randomLevel() int {
	level := 1
	for level < maxLevel && rand.Uint32() < pThreshold {
		level++
	}
	return level
}

func newNode(key string, value string, level int) *node {
	n := &node{
		key:   key,
		level: level,
	}
	n.value.Store(&value)
	return n
}

func New() *SkipMap {
	head := &node{
		level: maxLevel,
	}
	tail := &node{
		level: maxLevel,
	}
	for i := 0; i < maxLevel; i++ {
		head.next[i].Store(tail)
	}
	sm := &SkipMap{
		head: head,
		tail: tail,
	}
	sm.level.Store(1)
	return sm
}

func (sm *SkipMap) isTail(n *node) bool {
	return n == sm.tail
}

func (sm *SkipMap) isHead(n *node) bool {
	return n == sm.head
}

func (sm *SkipMap) findNode(key string, preds, succs []*node) *node {
	var pred, curr, succ *node
	var deleted bool

retry:
	pred = sm.head
	for i := int(sm.level.Load()) - 1; i >= 0; i-- {
		curr = pred.next[i].Load()
		for {
			succ = curr.next[i].Load()
			deleted = curr.deleted.Load()
			for deleted {
				if !pred.next[i].CompareAndSwap(curr, succ) {
					goto retry
				}
				curr = pred.next[i].Load()
				if sm.isTail(curr) {
					break
				}
				succ = curr.next[i].Load()
				deleted = curr.deleted.Load()
			}
			if !sm.isTail(curr) && curr.key < key {
				pred = curr
				curr = succ
			} else {
				break
			}
		}
		if preds != nil {
			preds[i] = pred
		}
		if succs != nil {
			succs[i] = curr
		}
	}

	for !sm.isTail(curr) && curr.deleted.Load() {
		curr = curr.next[0].Load()
	}

	if !sm.isTail(curr) && curr.key == key {
		return curr
	}
	return nil
}

func (sm *SkipMap) Put(key string, value string) (bool, error) {
	preds := make([]*node, maxLevel)
	succs := make([]*node, maxLevel)

	for {
		found := sm.findNode(key, preds, succs)
		if found != nil {
			found.value.Store(&value)
			return true, nil
		}

		newLevel := randomLevel()
		newNode := newNode(key, value, newLevel)

		for i := 0; i < newLevel; i++ {
			newNode.next[i].Store(succs[i])
		}

		if !preds[0].next[0].CompareAndSwap(succs[0], newNode) {
			continue
		}

		for i := 1; i < newLevel; i++ {
			for {
				pred := preds[i]
				succ := succs[i]
				if pred.next[i].CompareAndSwap(succ, newNode) {
					break
				}
				sm.findNode(key, preds, succs)
			}
		}

		for {
			currentLevel := sm.level.Load()
			if int32(newLevel) <= currentLevel {
				break
			}
			if sm.level.CompareAndSwap(currentLevel, int32(newLevel)) {
				break
			}
		}

		sm.size.Add(1)
		return true, nil
	}
}

func (sm *SkipMap) Get(key string) (string, bool) {
	var curr *node
	pred := sm.head
	for i := int(sm.level.Load()) - 1; i >= 0; i-- {
		curr = pred.next[i].Load()
		for !sm.isTail(curr) && curr.key < key {
			pred = curr
			curr = curr.next[i].Load()
		}
	}

	for !sm.isTail(curr) && curr.deleted.Load() {
		curr = curr.next[0].Load()
	}

	if !sm.isTail(curr) && curr.key == key {
		val := curr.value.Load()
		if val == nil {
			return "", false
		}
		return *val, true
	}
	return "", false
}

func (sm *SkipMap) Delete(key string) bool {
	preds := make([]*node, maxLevel)
	succs := make([]*node, maxLevel)

	found := sm.findNode(key, preds, succs)
	if found == nil {
		return false
	}

	if !found.deleted.CompareAndSwap(false, true) {
		return false
	}

	for i := found.level - 1; i >= 0; i-- {
		for {
			succ := found.next[i].Load()
			if preds[i].next[i].CompareAndSwap(found, succ) {
				break
			}
			sm.findNode(key, preds, succs)
		}
	}

	sm.size.Add(-1)
	return true
}

func (sm *SkipMap) Size() int64 {
	return sm.size.Load()
}

func (sm *SkipMap) RangeScan(start, end string) *Iterator {
	return &Iterator{
		sm:    sm,
		curr:  nil,
		end:   end,
		hasEnd: len(end) > 0,
	}
}

func (it *Iterator) findStart(start string) *node {
	var curr *node
	pred := it.sm.head
	for i := int(it.sm.level.Load()) - 1; i >= 0; i-- {
		curr = pred.next[i].Load()
		for !it.sm.isTail(curr) && curr.key < start {
			pred = curr
			curr = curr.next[i].Load()
		}
	}
	for !it.sm.isTail(curr) && curr.deleted.Load() {
		curr = curr.next[0].Load()
	}
	return curr
}

func (it *Iterator) Next(key *string, value *string) bool {
	if it.curr == nil {
		it.curr = it.sm.head.next[0].Load()
		for !it.sm.isTail(it.curr) && it.curr.deleted.Load() {
			it.curr = it.curr.next[0].Load()
		}
	}

	for !it.sm.isTail(it.curr) {
		if it.hasEnd && it.curr.key >= it.end {
			return false
		}
		*key = it.curr.key
		val := it.curr.value.Load()
		if val != nil {
			*value = *val
			next := it.curr.next[0].Load()
			for !it.sm.isTail(next) && next.deleted.Load() {
				next = next.next[0].Load()
			}
			it.curr = next
			return true
		}
		next := it.curr.next[0].Load()
		for !it.sm.isTail(next) && next.deleted.Load() {
			next = next.next[0].Load()
		}
		it.curr = next
	}
	return false
}

func (it *Iterator) Collect() []*KVPair {
	var pairs []*KVPair
	var key, value string
	for it.Next(&key, &value) {
		pairs = append(pairs, &KVPair{Key: key, Value: value})
	}
	return pairs
}

type KVPair struct {
	Key   string
	Value string
}

func init() {
	_ = unsafe.Sizeof(node{})
}
