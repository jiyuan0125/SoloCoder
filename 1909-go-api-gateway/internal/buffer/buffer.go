package buffer

import (
	"sync"
	"time"
)

type LogEntry struct {
	Method    string
	Path      string
	Status    int
	Duration  int64
	Timestamp time.Time
}

type RingBuffer struct {
	mu     sync.Mutex
	buf    []LogEntry
	cap    int
	head   int
	count  int
}

func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		buf: make([]LogEntry, capacity),
		cap: capacity,
	}
}

func (r *RingBuffer) Push(entry LogEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.buf[r.head] = entry
	r.head = (r.head + 1) % r.cap
	if r.count < r.cap {
		r.count++
	}
}

func (r *RingBuffer) ReadAll() []LogEntry {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := make([]LogEntry, r.count)
	if r.count == 0 {
		return result
	}

	start := (r.head - r.count + r.cap) % r.cap
	for i := 0; i < r.count; i++ {
		result[i] = r.buf[(start+i)%r.cap]
	}
	return result
}
