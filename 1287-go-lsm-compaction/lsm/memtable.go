package lsm

import (
	"sort"
	"sync"
)

type MemTable struct {
	data    map[string]*Entry
	size    int
	maxSize int
	mu      sync.RWMutex
}

func NewMemTable(maxSize int) *MemTable {
	return &MemTable{
		data:    make(map[string]*Entry),
		size:    0,
		maxSize: maxSize,
	}
}

func (m *MemTable) Put(key, value string, seqNum uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if old, exists := m.data[key]; exists {
		m.size -= old.Size()
	}
	
	entry := &Entry{
		Key:     key,
		Value:   value,
		Deleted: false,
		SeqNum:  seqNum,
	}
	
	m.data[key] = entry
	m.size += entry.Size()
}

func (m *MemTable) Delete(key string, seqNum uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if old, exists := m.data[key]; exists {
		m.size -= old.Size()
	}
	
	entry := &Entry{
		Key:     key,
		Value:   "",
		Deleted: true,
		SeqNum:  seqNum,
	}
	
	m.data[key] = entry
	m.size += entry.Size()
}

func (m *MemTable) Get(key string) (*Entry, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	entry, exists := m.data[key]
	return entry, exists
}

func (m *MemTable) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.size
}

func (m *MemTable) IsFull() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.size >= m.maxSize
}

func (m *MemTable) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

func (m *MemTable) Entries() []*Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	entries := make([]*Entry, 0, len(m.data))
	for _, e := range m.data {
		entries = append(entries, e)
	}
	
	sort.Sort(EntryList(entries))
	return entries
}

func (m *MemTable) Range(start, end string) []*Entry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	entries := make([]*Entry, 0)
	for _, e := range m.data {
		if e.Key >= start && (end == "" || e.Key <= end) {
			entries = append(entries, e)
		}
	}
	
	sort.Sort(EntryList(entries))
	return entries
}
