package skiplist

import (
	"sync"
	"sync/atomic"
)

const (
	DefaultMaxLevel = 32
	DefaultP        = 0.25
)

type ConcurrentStrategy int

const (
	StrategyGlobalLock ConcurrentStrategy = iota
	StrategyNodeLevelLock
)

func (s ConcurrentStrategy) String() string {
	switch s {
	case StrategyGlobalLock:
		return "global_lock"
	case StrategyNodeLevelLock:
		return "node_level_lock"
	default:
		return "unknown"
	}
}

type Node struct {
	Key      string
	Value    interface{}
	Next     []*Node
	Backward *Node
	Level    int
	mu       sync.Mutex
	deleted  int32
}

func newNode(key string, value interface{}, level int) *Node {
	return &Node{
		Key:     key,
		Value:   value,
		Next:    make([]*Node, level+1),
		Level:   level,
		deleted: 0,
	}
}

func (n *Node) isDeleted() bool {
	return atomic.LoadInt32(&n.deleted) == 1
}

func (n *Node) markDeleted() {
	atomic.StoreInt32(&n.deleted, 1)
}

type SkipListStats struct {
	mu                sync.RWMutex
	totalOperations   int64
	successOperations int64
	failedOperations  int64
	conflictCount     int64
	totalWaitTime     int64
	insertCount       int64
	deleteCount       int64
	getCount          int64
	rangeCount        int64
}

func newSkipListStats() *SkipListStats {
	return &SkipListStats{}
}

func (s *SkipListStats) recordOperation(success bool, waitTime int64, opType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalOperations++
	if success {
		s.successOperations++
	} else {
		s.failedOperations++
	}
	s.totalWaitTime += waitTime
	switch opType {
	case "insert":
		s.insertCount++
	case "delete":
		s.deleteCount++
	case "get":
		s.getCount++
	case "range":
		s.rangeCount++
	}
}

func (s *SkipListStats) recordConflict() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.conflictCount++
}

func (s *SkipListStats) getStats() SkipListStatsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var avgWaitTime float64
	if s.totalOperations > 0 {
		avgWaitTime = float64(s.totalWaitTime) / float64(s.totalOperations) / 1e6
	}
	return SkipListStatsSnapshot{
		TotalOperations:   s.totalOperations,
		SuccessOperations: s.successOperations,
		FailedOperations:  s.failedOperations,
		ConflictCount:     s.conflictCount,
		TotalWaitTime:     s.totalWaitTime,
		AvgWaitTime:       avgWaitTime,
		InsertCount:       s.insertCount,
		DeleteCount:       s.deleteCount,
		GetCount:          s.getCount,
		RangeCount:        s.rangeCount,
	}
}

func (s *SkipListStats) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalOperations = 0
	s.successOperations = 0
	s.failedOperations = 0
	s.conflictCount = 0
	s.totalWaitTime = 0
	s.insertCount = 0
	s.deleteCount = 0
	s.getCount = 0
	s.rangeCount = 0
}

type SkipListStatsSnapshot struct {
	TotalOperations   int64
	SuccessOperations int64
	FailedOperations  int64
	ConflictCount     int64
	TotalWaitTime     int64
	AvgWaitTime       float64
	InsertCount       int64
	DeleteCount       int64
	GetCount          int64
	RangeCount        int64
}

type ConcurrentSkipList interface {
	Insert(key string, value interface{}) bool
	Delete(key string) bool
	Get(key string) (interface{}, bool)
	Range(start, end string) []*KeyValuePair
	GetStats() SkipListStatsSnapshot
	ResetStats()
	Close()
}

type KeyValuePair struct {
	Key   string
	Value interface{}
}
