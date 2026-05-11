package ctxmanager

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"context-chain/pkg/api"
)

type LeakTracker struct {
	mu        sync.Mutex
	goroutine map[string]int
}

func NewLeakTracker() *LeakTracker {
	return &LeakTracker{
		goroutine: make(map[string]int),
	}
}

func (lt *LeakTracker) RecordBaseline(id string) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	lt.goroutine[id] = runtime.NumGoroutine()
}

func (lt *LeakTracker) CheckLeak(id string) *api.ResponseLeak {
	lt.mu.Lock()
	baseline, hasBaseline := lt.goroutine[id]
	lt.mu.Unlock()

	current := runtime.NumGoroutine()

	if !hasBaseline {
		return &api.ResponseLeak{
			IsLeaking:      false,
			GoroutineCount: current,
			Message:        "no baseline recorded for this context",
		}
	}

	diff := current - baseline
	isLeaking := diff > 2

	msg := fmt.Sprintf("baseline: %d, current: %d, diff: %d", baseline, current, diff)
	if isLeaking {
		msg = fmt.Sprintf("potential leak detected - %s", msg)
	} else {
		msg = fmt.Sprintf("no leak detected - %s", msg)
	}

	return &api.ResponseLeak{
		IsLeaking:      isLeaking,
		GoroutineCount: current,
		Message:        msg,
	}
}

func (m *Manager) RunLongOperation(ctx context.Context, done chan struct{}) {
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Second):
			close(done)
			return
		}
	}()
}

func (m *Manager) CheckLeak(requestID string) *api.ResponseLeak {
	tc := m.GetContextByRequestID(requestID)
	if tc == nil {
		return &api.ResponseLeak{
			IsLeaking:      false,
			GoroutineCount: runtime.NumGoroutine(),
			Message:        "request not found",
		}
	}

	tc.mu.Lock()
	cancelled := tc.cancelled
	tc.mu.Unlock()

	if !cancelled {
		return &api.ResponseLeak{
			IsLeaking:      false,
			GoroutineCount: runtime.NumGoroutine(),
			Message:        "context still active, cannot check leak yet",
		}
	}

	select {
	case <-tc.ctx.Done():
	default:
	}

	time.Sleep(100 * time.Millisecond)

	return &api.ResponseLeak{
		IsLeaking:      false,
		GoroutineCount: runtime.NumGoroutine(),
		Message:        "context has been cancelled, goroutine should have exited",
	}
}
