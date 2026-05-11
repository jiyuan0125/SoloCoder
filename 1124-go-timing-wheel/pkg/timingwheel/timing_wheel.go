package timingwheel

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type WheelLevel int

const (
	WheelLevelSecond WheelLevel = iota
	WheelLevelMinute
	WheelLevelHour
)

const (
	secondSize = 60
	minuteSize = 60
	hourSize   = 24
)

type Stats struct {
	TotalPending   int
	ExecutedCount  int64
	CancelledCount int64
	HourTasks      int
	MinuteTasks    int
	SecondTasks    int
	HourTick       int
	MinuteTick     int
	SecondTick     int
}

type TimingWheel struct {
	hourWheel   *Wheel
	minuteWheel *Wheel
	secondWheel *Wheel

	tasksMu sync.RWMutex
	tasks   map[string]*Task

	executedCount  atomicInt64
	cancelledCount atomicInt64

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	started bool
}

type atomicInt64 struct {
	val int64
	mu  sync.Mutex
}

func (a *atomicInt64) Add(n int64) int64 {
	a.mu.Lock()
	a.val += n
	result := a.val
	a.mu.Unlock()
	return result
}

func (a *atomicInt64) Load() int64 {
	a.mu.Lock()
	n := a.val
	a.mu.Unlock()
	return n
}

func New() *TimingWheel {
	return &TimingWheel{
		hourWheel:   NewWheel("hour", hourSize),
		minuteWheel: NewWheel("minute", minuteSize),
		secondWheel: NewWheel("second", secondSize),
		tasks:       make(map[string]*Task),
	}
}

func (tw *TimingWheel) Start() {
	tw.tasksMu.Lock()
	if tw.started {
		tw.tasksMu.Unlock()
		return
	}
	tw.started = true
	tw.ctx, tw.cancel = context.WithCancel(context.Background())
	tw.tasksMu.Unlock()

	tw.wg.Add(1)
	go tw.tickLoop()
}

func (tw *TimingWheel) Stop() {
	tw.tasksMu.Lock()
	if !tw.started {
		tw.tasksMu.Unlock()
		return
	}
	tw.started = false
	tw.tasksMu.Unlock()

	tw.cancel()
	tw.wg.Wait()
}

func (tw *TimingWheel) tickLoop() {
	defer tw.wg.Done()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-tw.ctx.Done():
			return
		case <-ticker.C:
			tw.advanceSecond()
		}
	}
}

func (tw *TimingWheel) advanceSecond() {
	tasks := tw.secondWheel.drainCurrentTick()
	for _, t := range tasks {
		if t.IsCancelled() {
			tw.removeTaskFromMap(t.ID)
			tw.cancelledCount.Add(1)
			continue
		}
		if t.Status() == TaskStatusPending {
			if t.onExecute != nil {
				t.onExecute(t)
			}
			tw.removeTaskFromMap(t.ID)
			tw.executedCount.Add(1)
		}
	}

	if tw.secondWheel.currentTick() == 0 {
		tw.advanceMinute()
	}
}

func (tw *TimingWheel) advanceMinute() {
	tasks := tw.minuteWheel.drainCurrentTick()
	for _, t := range tasks {
		if t.IsCancelled() {
			tw.removeTaskFromMap(t.ID)
			tw.cancelledCount.Add(1)
			continue
		}
		remaining := t.Remaining()
		tw.reinsertTask(t, remaining)
	}

	if tw.minuteWheel.currentTick() == 0 {
		tw.advanceHour()
	}
}

func (tw *TimingWheel) advanceHour() {
	tasks := tw.hourWheel.drainCurrentTick()
	for _, t := range tasks {
		if t.IsCancelled() {
			tw.removeTaskFromMap(t.ID)
			tw.cancelledCount.Add(1)
			continue
		}
		remaining := t.Remaining()
		tw.reinsertTask(t, remaining)
	}
}

func (tw *TimingWheel) reinsertTask(t *Task, remaining time.Duration) {
	if remaining <= 0 {
		if t.onExecute != nil {
			t.onExecute(t)
		}
		tw.removeTaskFromMap(t.ID)
		tw.executedCount.Add(1)
		return
	}

	hours := int(remaining / time.Hour)
	remaining -= time.Duration(hours) * time.Hour
	minutes := int(remaining / time.Minute)
	seconds := int((remaining - time.Duration(minutes)*time.Minute) / time.Second)

	if hours > 0 {
		offset := hours
		if minutes > 0 || seconds > 0 {
			offset++
		}
		tw.hourWheel.addTask(t, offset)
	} else if minutes > 0 {
		offset := minutes
		if seconds > 0 {
			offset++
		}
		tw.minuteWheel.addTask(t, offset)
	} else {
		tw.secondWheel.addTask(t, seconds)
	}
}

func (tw *TimingWheel) Add(id string, callback string, delay time.Duration) error {
	tw.tasksMu.Lock()
	if _, exists := tw.tasks[id]; exists {
		tw.tasksMu.Unlock()
		return fmt.Errorf("task %s already exists", id)
	}
	t := NewTask(id, callback, delay)
	t.onExecute = tw.onExecute
	tw.tasks[id] = t
	tw.tasksMu.Unlock()

	tw.reinsertTask(t, delay)
	return nil
}

func (tw *TimingWheel) onExecute(t *Task) {
	t.SetStatus(TaskStatusExecuted)
}

func (tw *TimingWheel) Cancel(id string) (bool, error) {
	tw.tasksMu.RLock()
	t, exists := tw.tasks[id]
	tw.tasksMu.RUnlock()

	if !exists {
		return false, fmt.Errorf("task %s not found", id)
	}

	if !t.Cancel() {
		return false, fmt.Errorf("task %s is not pending", id)
	}

	return true, nil
}

func (tw *TimingWheel) Reset(id string, newDelay time.Duration) error {
	tw.tasksMu.RLock()
	oldTask, exists := tw.tasks[id]
	tw.tasksMu.RUnlock()

	if !exists {
		return fmt.Errorf("task %s not found", id)
	}

	callback := oldTask.Callback

	cancelled, _ := tw.Cancel(id)
	if !cancelled {
		return fmt.Errorf("task %s is not cancellable for reset", id)
	}

	tw.removeTaskFromMap(id)

	t := NewTask(id, callback, newDelay)
	t.onExecute = tw.onExecute

	tw.tasksMu.Lock()
	tw.tasks[id] = t
	tw.tasksMu.Unlock()

	tw.reinsertTask(t, newDelay)
	return nil
}

func (tw *TimingWheel) removeTaskFromMap(id string) {
	tw.tasksMu.Lock()
	delete(tw.tasks, id)
	tw.tasksMu.Unlock()
}

func (tw *TimingWheel) List() []*Task {
	tw.tasksMu.RLock()
	result := make([]*Task, 0, len(tw.tasks))
	for _, t := range tw.tasks {
		result = append(result, t)
	}
	tw.tasksMu.RUnlock()
	return result
}

func (tw *TimingWheel) Get(id string) (*Task, error) {
	tw.tasksMu.RLock()
	t, exists := tw.tasks[id]
	tw.tasksMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("task %s not found", id)
	}
	return t, nil
}

func (tw *TimingWheel) Stats() *Stats {
	return &Stats{
		TotalPending:   tw.hourWheel.totalTasks() + tw.minuteWheel.totalTasks() + tw.secondWheel.totalTasks(),
		ExecutedCount:  tw.executedCount.Load(),
		CancelledCount: tw.cancelledCount.Load(),
		HourTasks:      tw.hourWheel.totalTasks(),
		MinuteTasks:    tw.minuteWheel.totalTasks(),
		SecondTasks:    tw.secondWheel.totalTasks(),
		HourTick:       tw.hourWheel.currentTick(),
		MinuteTick:     tw.minuteWheel.currentTick(),
		SecondTick:     tw.secondWheel.currentTick(),
	}
}
