package hashtable

import (
	"errors"
	"fmt"
)

type ProbeMethod int

const (
	Linear ProbeMethod = iota
	Quadratic
	Double
)

const (
	LoadFactorThreshold = 0.7
)

type EntryStatus int

const (
	Empty EntryStatus = iota
	Occupied
	Tombstone
)

type Entry struct {
	Key    string
	Value  string
	Status EntryStatus
}

type HashTable struct {
	table         []Entry
	capacity      int
	size          int
	tombstones    int
	probeMethod   ProbeMethod
	maxProbeLen   int
}

var (
	ErrNotFound = errors.New("key not found")
	ErrTableFull = errors.New("hash table is full")
)

func (p ProbeMethod) String() string {
	switch p {
	case Linear:
		return "linear"
	case Quadratic:
		return "quadratic"
	case Double:
		return "double"
	default:
		return "unknown"
	}
}

func ParseProbeMethod(s string) (ProbeMethod, error) {
	switch s {
	case "linear":
		return Linear, nil
	case "quadratic":
		return Quadratic, nil
	case "double":
		return Double, nil
	default:
		return Linear, fmt.Errorf("invalid probe method: %s", s)
	}
}

func NewHashTable(method ProbeMethod, initialCapacity int) *HashTable {
	capacity := NextPrime(initialCapacity)
	return &HashTable{
		table:        make([]Entry, capacity),
		capacity:     capacity,
		size:         0,
		tombstones:   0,
		probeMethod:  method,
		maxProbeLen:  0,
	}
}

func (h *HashTable) Size() int {
	return h.size
}

func (h *HashTable) Tombstones() int {
	return h.tombstones
}

func (h *HashTable) Capacity() int {
	return h.capacity
}

func (h *HashTable) MaxProbeLen() int {
	return h.maxProbeLen
}

func (h *HashTable) Put(key, value string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	
	loadFactor := float64(h.size+h.tombstones) / float64(h.capacity)
	if loadFactor >= LoadFactorThreshold {
		h.rehash()
	}
	
	probeLen := 0
	idx := h.findInsertIndex(key, &probeLen)
	if idx < 0 {
		h.rehash()
		idx = h.findInsertIndex(key, &probeLen)
		if idx < 0 {
			return ErrTableFull
		}
	}
	
	if probeLen > h.maxProbeLen {
		h.maxProbeLen = probeLen
	}
	
	if h.table[idx].Status == Tombstone {
		h.tombstones--
	}
	h.table[idx].Key = key
	h.table[idx].Value = value
	h.table[idx].Status = Occupied
	h.size++
	
	return nil
}

func (h *HashTable) Get(key string) (string, error) {
	idx := h.findKeyIndex(key)
	if idx < 0 {
		return "", ErrNotFound
	}
	return h.table[idx].Value, nil
}

func (h *HashTable) Remove(key string) error {
	idx := h.findKeyIndex(key)
	if idx < 0 {
		return ErrNotFound
	}
	
	h.table[idx].Status = Tombstone
	h.table[idx].Key = ""
	h.table[idx].Value = ""
	h.size--
	h.tombstones++
	
	return nil
}

func (h *HashTable) findInsertIndex(key string, probeLen *int) int {
	hash := hash1(key)
	idx := hash % uint64(h.capacity)
	
	for i := 0; i < h.capacity; i++ {
		*probeLen = i + 1
		
		if h.table[idx].Status == Empty || h.table[idx].Status == Tombstone {
			return int(idx)
		}
		
		if h.table[idx].Status == Occupied && h.table[idx].Key == key {
			return int(idx)
		}
		
		idx = h.nextIndex(hash, key, i, idx)
	}
	
	return -1
}

func (h *HashTable) findKeyIndex(key string) int {
	hash := hash1(key)
	idx := hash % uint64(h.capacity)
	
	for i := 0; i < h.capacity; i++ {
		if h.table[idx].Status == Empty {
			return -1
		}
		
		if h.table[idx].Status == Occupied && h.table[idx].Key == key {
			return int(idx)
		}
		
		idx = h.nextIndex(hash, key, i, idx)
	}
	
	return -1
}

func (h *HashTable) nextIndex(hash1 uint64, key string, attempt int, currentIdx uint64) uint64 {
	switch h.probeMethod {
	case Linear:
		return (currentIdx + 1) % uint64(h.capacity)
		
	case Quadratic:
		return (hash1 + uint64(attempt)*uint64(attempt) + uint64(attempt)) % uint64(h.capacity)
		
	case Double:
		h2 := hash2(key, h.capacity)
		return (currentIdx + h2) % uint64(h.capacity)
		
	default:
		return (currentIdx + 1) % uint64(h.capacity)
	}
}

func (h *HashTable) rehash() {
	newCapacity := NextPrime(h.capacity * 2)
	oldTable := h.table
	h.table = make([]Entry, newCapacity)
	h.capacity = newCapacity
	h.size = 0
	h.tombstones = 0
	h.maxProbeLen = 0
	
	for _, entry := range oldTable {
		if entry.Status == Occupied {
			probeLen := 0
			idx := h.findInsertIndex(entry.Key, &probeLen)
			if idx >= 0 {
				h.table[idx].Key = entry.Key
				h.table[idx].Value = entry.Value
				h.table[idx].Status = Occupied
				h.size++
				if probeLen > h.maxProbeLen {
					h.maxProbeLen = probeLen
				}
			}
		}
	}
}
