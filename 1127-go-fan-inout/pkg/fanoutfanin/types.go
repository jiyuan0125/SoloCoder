package fanoutfanin

import (
	"sync"
	"sync/atomic"
	"time"
)

type DataItem struct {
	Source    string
	Payload   string
	Timestamp int64
}

type SourceInfo struct {
	Name        string
	Channel     chan<- DataItem
	Stop        chan struct{}
	Interval    time.Duration
	Payload     string
	Active      atomic.Bool
	TotalRecv   atomic.Int64
	TotalSent   atomic.Int64
	Dropped     atomic.Int64
	BufferCap   int
}

type Worker struct {
	ID         int
	Channel    chan DataItem
	Processed  atomic.Int64
}

type BackpressurePolicy string

const (
	PolicyBlock BackpressurePolicy = "block"
	PolicyDrop  BackpressurePolicy = "drop"
)

type Scheduler struct {
	mu             sync.RWMutex
	sources        map[string]*SourceInfo
	aggregated     chan DataItem
	workers        []*Worker
	policy         BackpressurePolicy
	workerPoolSize int
	workerBufCap   int
	sourceBufCap   int
	aggBufCap      int
	closed         atomic.Bool
	wg             sync.WaitGroup
}

type SchedulerConfig struct {
	WorkerPoolSize  int
	WorkerBufferCap int
	SourceBufferCap int
	InitialPolicy   BackpressurePolicy
	AggBufferCap    int
}

func DefaultConfig() SchedulerConfig {
	return SchedulerConfig{
		WorkerPoolSize:  4,
		WorkerBufferCap: 8,
		SourceBufferCap: 16,
		InitialPolicy:   PolicyBlock,
		AggBufferCap:    32,
	}
}
