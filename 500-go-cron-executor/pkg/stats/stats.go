package stats

import (
	"sync"
	"time"

	"cron-executor/pkg/model"
)

type Stats struct {
	TotalTasks       int
	RunningTasks     int
	TotalExecutions  int64
	SuccessCount     int64
	FailCount        int64
	TimeoutCount     int64
	TotalDuration    time.Duration
	AvgDuration      time.Duration
	SuccessRate      float64
	RecentFailures   []*model.Execution
}

type StatsManager struct {
	mu             sync.RWMutex
	totalExecs     int64
	successCount   int64
	failCount      int64
	timeoutCount   int64
	totalDuration  time.Duration
	recentFailures []*model.Execution
}

func NewStatsManager() *StatsManager {
	return &StatsManager{
		recentFailures: make([]*model.Execution, 0, 100),
	}
}

func (sm *StatsManager) RecordExecution(exec *model.Execution) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.totalExecs++

	if exec.Status == model.ExecStatusSuccess {
		sm.successCount++
	} else if exec.Status == model.ExecStatusFailed {
		sm.failCount++
	} else if exec.Status == model.ExecStatusTimeout {
		sm.timeoutCount++
	}

	if exec.Status == model.ExecStatusSuccess {
		sm.totalDuration += exec.Duration
	}

	if exec.Status == model.ExecStatusFailed || exec.Status == model.ExecStatusTimeout {
		sm.recentFailures = append(sm.recentFailures, exec)
		if len(sm.recentFailures) > 100 {
			sm.recentFailures = sm.recentFailures[len(sm.recentFailures)-100:]
		}
	}
}

func (sm *StatsManager) GetStats() *Stats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	s := &Stats{
		TotalExecutions: sm.totalExecs,
		SuccessCount:    sm.successCount,
		FailCount:       sm.failCount,
		TimeoutCount:    sm.timeoutCount,
	}

	if sm.successCount+sm.failCount+sm.timeoutCount > 0 {
		s.SuccessRate = float64(sm.successCount) / float64(sm.successCount+sm.failCount+sm.timeoutCount) * 100
	}

	if sm.successCount > 0 {
		s.AvgDuration = sm.totalDuration / time.Duration(sm.successCount)
	}

	s.RecentFailures = make([]*model.Execution, len(sm.recentFailures))
	copy(s.RecentFailures, sm.recentFailures)

	return s
}
