package targets

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

	"log-shipper/internal/types"
)

type MemoryTarget struct {
	id         string
	config     types.MemoryTargetConfig
	entries    *list.List
	maxEntries int

	lastWriteTime time.Time
	successCount  int64
	failureCount  int64
	writeCount    int64
	startTime     time.Time

	mu  sync.Mutex
}

func NewMemoryTarget(id string, config types.MemoryTargetConfig) *MemoryTarget {
	maxEntries := config.MaxEntries
	if maxEntries <= 0 {
		maxEntries = 1000
	}
	return &MemoryTarget{
		id:         id,
		config:     config,
		entries:    list.New(),
		maxEntries: maxEntries,
		startTime:  time.Now(),
	}
}

func (t *MemoryTarget) ID() string { return t.id }
func (t *MemoryTarget) Type() string { return "memory" }
func (t *MemoryTarget) Config() interface{} { return t.config }

func (t *MemoryTarget) Status() types.TargetStatus {
	elapsed := time.Since(t.startTime).Seconds()
	var rate float64
	if elapsed > 0 {
		rate = float64(atomic.LoadInt64(&t.writeCount)) / elapsed
	}
	return types.TargetStatus{
		ID:            t.id,
		Type:          "memory",
		Config:        t.config,
		LastWriteTime: t.lastWriteTime,
		SuccessCount:  atomic.LoadInt64(&t.successCount),
		FailureCount:  atomic.LoadInt64(&t.failureCount),
		WriteRate:     rate,
	}
}

func (t *MemoryTarget) Start() error { return nil }
func (t *MemoryTarget) Stop() error  { return nil }

func (t *MemoryTarget) Write(entries []*types.LogEntry) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, entry := range entries {
		t.entries.PushBack(entry)
		atomic.AddInt64(&t.writeCount, 1)
		if t.entries.Len() > t.maxEntries {
			t.entries.Remove(t.entries.Front())
		}
	}

	t.lastWriteTime = time.Now()
	atomic.AddInt64(&t.successCount, 1)
	return nil
}

func (t *MemoryTarget) GetEntries() []*types.LogEntry {
	t.mu.Lock()
	defer t.mu.Unlock()
	result := make([]*types.LogEntry, 0, t.entries.Len())
	for e := t.entries.Front(); e != nil; e = e.Next() {
		result = append(result, e.Value.(*types.LogEntry))
	}
	return result
}
