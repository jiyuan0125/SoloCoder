package taskqueue

import "sync/atomic"

type DeadLetterTask struct {
	Task       *Task
	RetryCount int
	LastError  error
}

func (q *TaskQueue) DeadLettersWithInfo() []*DeadLetterTask {
	q.mu.Lock()
	defer q.mu.Unlock()

	result := make([]*DeadLetterTask, len(q.deadLetters))
	for i, task := range q.deadLetters {
		result[i] = &DeadLetterTask{
			Task:       task,
			RetryCount: task.retryCount,
			LastError:  task.lastErr,
		}
	}
	return result
}

func (q *TaskQueue) ClearDeadLetters() {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := len(q.deadLetters)
	if count > 0 {
		current := atomic.LoadUint64(&q.stats.deadLetters)
		if current >= uint64(count) {
			atomic.StoreUint64(&q.stats.deadLetters, current-uint64(count))
		} else {
			atomic.StoreUint64(&q.stats.deadLetters, 0)
		}
	}
	q.deadLetters = make([]*Task, 0)
}
