package memtable

import "lsm-storage/internal/skiplist"

const DefaultThreshold = 4 * 1024 * 1024

type MemTable struct {
	sl        *skiplist.SkipList
	size      int
	threshold int
}

func New(threshold int) *MemTable {
	if threshold <= 0 {
		threshold = DefaultThreshold
	}
	return &MemTable{
		sl:        skiplist.New(),
		threshold: threshold,
	}
}

func (m *MemTable) Put(key, value []byte) {
	m.sl.Put(key, value, false)
	m.size += len(key) + len(value) + 1
}

func (m *MemTable) Delete(key []byte) {
	m.sl.Delete(key)
	m.size += len(key) + 1
}

func (m *MemTable) Get(key []byte) ([]byte, bool, bool) {
	return m.sl.Get(key)
}

func (m *MemTable) ShouldFlush() bool {
	return m.size >= m.threshold
}

func (m *MemTable) Size() int {
	return m.size
}

func (m *MemTable) Iterator() *skiplist.Iterator {
	return m.sl.Iterator()
}
