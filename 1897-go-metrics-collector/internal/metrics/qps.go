package metrics

import (
	"sync"
	"time"
)

const QPSWindow = 60

type QPSCounter struct {
	buckets   []int
	startTime time.Time
	mu        sync.Mutex
}

func NewQPSCounter() *QPSCounter {
	return &QPSCounter{
		buckets:   make([]int, QPSWindow),
		startTime: time.Now(),
	}
}

func (q *QPSCounter) currentBucketIndex() int {
	elapsed := time.Since(q.startTime)
	return int(elapsed.Seconds()) % QPSWindow
}

func (q *QPSCounter) rotate() {
	now := time.Now()
	elapsed := int(now.Sub(q.startTime).Seconds())
	if elapsed >= QPSWindow {
		rotateCount := elapsed - QPSWindow + 1
		for i := 0; i < rotateCount; i++ {
			idx := (q.currentBucketIndex() + i + 1) % QPSWindow
			q.buckets[idx] = 0
		}
		q.startTime = now
	}
}

func (q *QPSCounter) Increment() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.rotate()
	idx := q.currentBucketIndex()
	q.buckets[idx]++
}

func (q *QPSCounter) QPS() float64 {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.rotate()

	elapsed := int(time.Since(q.startTime).Seconds())
	if elapsed == 0 {
		elapsed = 1
	}
	if elapsed > QPSWindow {
		elapsed = QPSWindow
	}

	total := 0
	for i := 0; i < QPSWindow; i++ {
		total += q.buckets[i]
	}

	return float64(total) / float64(elapsed)
}

func (q *QPSCounter) Reset() {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i := range q.buckets {
		q.buckets[i] = 0
	}
	q.startTime = time.Now()
}
