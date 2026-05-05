package progress

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ProgressCallback func(current, total int64, percentage float64)

type Tracker struct {
	mu        sync.RWMutex
	total     int64
	current   int64
	startTime time.Time
	callbacks []ProgressCallback
	subTasks  []*SubTask
	cancel    context.CancelFunc
	ctx       context.Context
	closed    bool
	cancelled bool
}

type SubTask struct {
	tracker *Tracker
	parent  *Tracker
	weight  float64
}

func New(total int64) *Tracker {
	return NewWithContext(context.Background(), total)
}

func NewWithContext(ctx context.Context, total int64) *Tracker {
	t := &Tracker{
		total:     total,
		current:   0,
		startTime: time.Now(),
		callbacks: make([]ProgressCallback, 0),
		subTasks:  make([]*SubTask, 0),
	}

	t.ctx, t.cancel = context.WithCancel(ctx)
	return t
}

func (t *Tracker) Update(delta int64) {
	t.mu.Lock()

	if t.closed || t.cancelled {
		t.mu.Unlock()
		return
	}

	select {
	case <-t.ctx.Done():
		t.mu.Unlock()
		return
	default:
	}

	newCurrent := t.current + delta
	if newCurrent <= t.current {
		t.mu.Unlock()
		return
	}

	if t.total > 0 && newCurrent > t.total {
		newCurrent = t.total
	}

	t.current = newCurrent

	current := t.current
	total := t.total
	percentage := t.calculatePercentageLocked()
	callbacks := make([]ProgressCallback, len(t.callbacks))
	copy(callbacks, t.callbacks)

	t.mu.Unlock()

	t.notifyCallbacksWithValues(callbacks, current, total, percentage)
}

func (t *Tracker) Set(current int64) {
	t.mu.Lock()

	if t.closed || t.cancelled {
		t.mu.Unlock()
		return
	}

	select {
	case <-t.ctx.Done():
		t.mu.Unlock()
		return
	default:
	}

	if current <= t.current {
		t.mu.Unlock()
		return
	}

	if t.total > 0 && current > t.total {
		current = t.total
	}

	t.current = current

	savedCurrent := t.current
	savedTotal := t.total
	percentage := t.calculatePercentageLocked()
	callbacks := make([]ProgressCallback, len(t.callbacks))
	copy(callbacks, t.callbacks)

	t.mu.Unlock()

	t.notifyCallbacksWithValues(callbacks, savedCurrent, savedTotal, percentage)
}

func (t *Tracker) Current() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.current
}

func (t *Tracker) Total() int64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.total
}

func (t *Tracker) Percentage() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.calculatePercentageLocked()
}

func (t *Tracker) calculatePercentageLocked() float64 {
	if t.total <= 0 {
		return 0.0
	}

	if len(t.subTasks) > 0 {
		return t.calculateSubTaskPercentageLocked()
	}

	return (float64(t.current) / float64(t.total)) * 100.0
}

func (t *Tracker) calculateSubTaskPercentageLocked() float64 {
	var totalPercentage float64
	var totalWeight float64

	for _, st := range t.subTasks {
		select {
		case <-st.tracker.ctx.Done():
			continue
		default:
		}
		st.tracker.mu.RLock()
		subPercentage := st.tracker.calculatePercentageLocked()
		st.tracker.mu.RUnlock()
		totalPercentage += subPercentage * st.weight
		totalWeight += st.weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalPercentage / totalWeight
}

func (t *Tracker) RemainingTime() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()

	percentage := t.calculatePercentageLocked()

	if percentage <= 0 {
		return -1
	}

	if percentage >= 100 {
		return 0
	}

	elapsed := time.Since(t.startTime)
	remainingRatio := (100 - percentage) / percentage
	return time.Duration(float64(elapsed) * remainingRatio)
}

func (t *Tracker) RegisterCallback(callback ProgressCallback) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.callbacks = append(t.callbacks, callback)
}

func (t *Tracker) notifyCallbacksWithValues(callbacks []ProgressCallback, current, total int64, percentage float64) {
	for _, cb := range callbacks {
		cb(current, total, percentage)
	}
}

func (t *Tracker) AddSubTask(total int64) *Tracker {
	return t.AddSubTaskWithWeight(total, 0)
}

func (t *Tracker) AddSubTaskWithWeight(total int64, weight float64) *Tracker {
	t.mu.Lock()

	subTracker := NewWithContext(t.ctx, total)

	subTask := &SubTask{
		tracker: subTracker,
		parent:  t,
		weight:  weight,
	}

	t.subTasks = append(t.subTasks, subTask)

	if weight == 0 {
		t.rebalanceSubTaskWeightsLocked()
	}

	parent := t
	subTracker.RegisterCallback(func(current, total int64, percentage float64) {
		parent.mu.RLock()
		parentCurrent := parent.current
		parentTotal := parent.total
		parentPercentage := parent.calculatePercentageLocked()
		parentCallbacks := make([]ProgressCallback, len(parent.callbacks))
		copy(parentCallbacks, parent.callbacks)
		parent.mu.RUnlock()

		parent.notifyCallbacksWithValues(parentCallbacks, parentCurrent, parentTotal, parentPercentage)
	})

	t.mu.Unlock()

	return subTracker
}

func (t *Tracker) rebalanceSubTaskWeightsLocked() {
	if len(t.subTasks) == 0 {
		return
	}

	equalWeight := 1.0 / float64(len(t.subTasks))
	for _, st := range t.subTasks {
		if st.weight == 0 {
			st.weight = equalWeight
		}
	}

	var totalWeight float64
	for _, st := range t.subTasks {
		totalWeight += st.weight
	}

	if totalWeight != 1.0 {
		for _, st := range t.subTasks {
			st.weight = st.weight / totalWeight
		}
	}
}

func (t *Tracker) Cancel() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cancelled || t.closed {
		return
	}

	t.cancelled = true

	if t.cancel != nil {
		t.cancel()
	}
}

func (t *Tracker) Done() <-chan struct{} {
	return t.ctx.Done()
}

func (t *Tracker) String() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	percentage := t.calculatePercentageLocked()
	return fmt.Sprintf("已完成%d/%d (%.1f%%)", t.current, t.total, percentage)
}

func (t *Tracker) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return
	}

	t.closed = true

	if t.cancel != nil {
		t.cancel()
	}

	for _, st := range t.subTasks {
		st.tracker.Close()
	}
}

func (t *Tracker) IsDone() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.current >= t.total
}

func (t *Tracker) IsClosed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.closed
}

func (t *Tracker) IsCancelled() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cancelled
}
