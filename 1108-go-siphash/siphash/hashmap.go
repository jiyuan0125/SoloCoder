package siphash

import (
	"sync"
	"sync/atomic"
)

const (
	loadFactor = 0.75
	initSize   = 16
)

type entry struct {
	key   []byte
	value []byte
	next  *entry
}

type HashMap struct {
	mu      sync.RWMutex
	resize  int32
	buckets []*entry
	count   int
	key     *Key
	c, d    int
}

func NewHashMap(key *Key) *HashMap {
	return NewHashMapWithRounds(key, DefaultC, DefaultD)
}

func NewHashMap13(key *Key) *HashMap {
	return NewHashMapWithRounds(key, 1, 3)
}

func NewHashMap24(key *Key) *HashMap {
	return NewHashMapWithRounds(key, 2, 4)
}

func NewHashMapWithRounds(key *Key, c, d int) *HashMap {
	return &HashMap{
		buckets: make([]*entry, initSize),
		key:     key,
		c:       c,
		d:       d,
	}
}

func (h *HashMap) hash(key []byte) uint64 {
	return Sum64WithRounds(h.key, key, h.c, h.d)
}

func (h *HashMap) index(hash uint64, bucketCount int) int {
	return int(hash & uint64(bucketCount-1))
}

func (h *HashMap) needResize() bool {
	bucketCount := len(h.buckets)
	return float64(h.count)/float64(bucketCount) >= loadFactor
}

func (h *HashMap) Put(key, value []byte) {
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)

	h.mu.Lock()
	defer h.mu.Unlock()

	hash := h.hash(key)
	idx := h.index(hash, len(h.buckets))

	for e := h.buckets[idx]; e != nil; e = e.next {
		if bytesEqual(e.key, key) {
			e.value = valueCopy
			return
		}
	}

	h.buckets[idx] = &entry{
		key:   keyCopy,
		value: valueCopy,
		next:  h.buckets[idx],
	}
	h.count++

	if h.needResize() {
		h.resizeBucket()
	}
}

func (h *HashMap) Get(key []byte) ([]byte, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	hash := h.hash(key)
	idx := h.index(hash, len(h.buckets))

	for e := h.buckets[idx]; e != nil; e = e.next {
		if bytesEqual(e.key, key) {
			result := make([]byte, len(e.value))
			copy(result, e.value)
			return result, true
		}
	}

	return nil, false
}

func (h *HashMap) Delete(key []byte) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	hash := h.hash(key)
	idx := h.index(hash, len(h.buckets))

	var prev *entry
	for e := h.buckets[idx]; e != nil; e = e.next {
		if bytesEqual(e.key, key) {
			if prev == nil {
				h.buckets[idx] = e.next
			} else {
				prev.next = e.next
			}
			h.count--
			return true
		}
		prev = e
	}

	return false
}

func (h *HashMap) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}

func (h *HashMap) Keys() [][]byte {
	h.mu.RLock()
	defer h.mu.RUnlock()

	keys := make([][]byte, 0, h.count)
	for _, bucket := range h.buckets {
		for e := bucket; e != nil; e = e.next {
			keyCopy := make([]byte, len(e.key))
			copy(keyCopy, e.key)
			keys = append(keys, keyCopy)
		}
	}
	return keys
}

func (h *HashMap) RotateKey(newKey *Key) {
	h.mu.Lock()
	defer h.mu.Unlock()

	oldBuckets := h.buckets
	h.key = newKey
	h.buckets = make([]*entry, len(oldBuckets))
	h.count = 0

	for _, bucket := range oldBuckets {
		for e := bucket; e != nil; e = e.next {
			hash := h.hash(e.key)
			idx := h.index(hash, len(h.buckets))

			found := false
			for e2 := h.buckets[idx]; e2 != nil; e2 = e2.next {
				if bytesEqual(e2.key, e.key) {
					e2.value = e.value
					found = true
					break
				}
			}

			if !found {
				h.buckets[idx] = &entry{
					key:   e.key,
					value: e.value,
					next:  h.buckets[idx],
				}
				h.count++
			}
		}
	}

	if h.needResize() {
		h.resizeBucket()
	}
}

func (h *HashMap) resizeBucket() {
	if !atomic.CompareAndSwapInt32(&h.resize, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&h.resize, 0)

	newSize := len(h.buckets) * 2
	if newSize < initSize {
		newSize = initSize
	}

	newBuckets := make([]*entry, newSize)

	for _, bucket := range h.buckets {
		for e := bucket; e != nil; {
			next := e.next
			hash := h.hash(e.key)
			idx := h.index(hash, newSize)
			e.next = newBuckets[idx]
			newBuckets[idx] = e
			e = next
		}
	}

	h.buckets = newBuckets
}

func (h *HashMap) Rounds() (c, d int) {
	return h.c, h.d
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
