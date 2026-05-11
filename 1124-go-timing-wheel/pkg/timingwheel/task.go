package timingwheel

import (
	"sync/atomic"
	"time"
)

type TaskStatus int32

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusRunning
	TaskStatusCancelled
	TaskStatusExecuted
)

type Task struct {
	ID          string
	Callback    string
	Delay       time.Duration
	CreatedAt   time.Time
	ExpireAt    time.Time
	status      atomic.Int32
	wheelLevel  int
	slotIndex   int
	onExecute   func(*Task)
}

func NewTask(id string, callback string, delay time.Duration) *Task {
	now := time.Now()
	t := &Task{
		ID:        id,
		Callback:  callback,
		Delay:     delay,
		CreatedAt: now,
		ExpireAt:  now.Add(delay),
	}
	t.status.Store(int32(TaskStatusPending))
	return t
}

func (t *Task) Status() TaskStatus {
	return TaskStatus(t.status.Load())
}

func (t *Task) Cancel() bool {
	return t.status.CompareAndSwap(int32(TaskStatusPending), int32(TaskStatusCancelled))
}

func (t *Task) SetStatus(s TaskStatus) {
	t.status.Store(int32(s))
}

func (t *Task) IsCancelled() bool {
	return t.Status() == TaskStatusCancelled
}

func (t *Task) Remaining() time.Duration {
	if time.Now().After(t.ExpireAt) {
		return 0
	}
	return time.Until(t.ExpireAt)
}
