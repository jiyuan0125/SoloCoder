package queue

import "time"

type Priority int

const (
	PriorityLow    Priority = 1
	PriorityNormal Priority = 2
	PriorityHigh   Priority = 3
	PriorityUrgent Priority = 4
)

type Task struct {
	ID        string
	Payload   []byte
	ExecuteAt time.Time
	Priority  Priority
	index     int
}

func (t *Task) IsExpired(now time.Time) bool {
	return !t.ExecuteAt.After(now)
}

func (t *Task) Index() int {
	return t.index
}
