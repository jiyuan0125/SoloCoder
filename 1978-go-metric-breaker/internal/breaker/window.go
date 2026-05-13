package breaker

import (
	"sync"
	"time"
)

type bucket struct {
	Timestamp       time.Time
	TotalRequests   int64
	ErrorRequests   int64
	TimeoutRequests int64
	TotalDuration   time.Duration
	CustomValues    map[string]float64
}

type slidingWindow struct {
	mu           sync.RWMutex
	duration     time.Duration
	bucketSize   time.Duration
	buckets      []*bucket
}

func newSlidingWindow(duration time.Duration) *slidingWindow {
	bucketSize := duration / 10
	if bucketSize < time.Millisecond*100 {
		bucketSize = time.Millisecond * 100
	}
	return &slidingWindow{
		duration:   duration,
		bucketSize: bucketSize,
		buckets:    make([]*bucket, 0),
	}
}

func (w *slidingWindow) getOrCreateBucket(now time.Time) *bucket {
	for i := len(w.buckets) - 1; i >= 0; i-- {
		b := w.buckets[i]
		if now.Sub(b.Timestamp) < w.bucketSize {
			return b
		}
	}
	b := &bucket{
		Timestamp:    now,
		CustomValues: make(map[string]float64),
	}
	w.buckets = append(w.buckets, b)
	return b
}

func (w *slidingWindow) evict(now time.Time) {
	cutoff := now.Add(-w.duration)
	startIdx := 0
	for i, b := range w.buckets {
		if b.Timestamp.After(cutoff) {
			startIdx = i
			break
		}
	}
	if startIdx > 0 {
		w.buckets = w.buckets[startIdx:]
	}
}

func (w *slidingWindow) Record(start, end time.Time, statusCode int, timeoutLimit time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	w.evict(now)
	b := w.getOrCreateBucket(now)

	duration := end.Sub(start)
	b.TotalRequests++
	b.TotalDuration += duration

	if statusCode >= 500 {
		b.ErrorRequests++
	}

	if timeoutLimit > 0 && duration >= timeoutLimit {
		b.TimeoutRequests++
	}
}

func (w *slidingWindow) RecordCustom(name string, value float64) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	w.evict(now)
	b := w.getOrCreateBucket(now)
	b.CustomValues[name] = value
}

func (w *slidingWindow) Snapshot() *Snapshot {
	w.mu.RLock()
	defer w.mu.RUnlock()

	now := time.Now()
	w.evict(now)

	snap := &Snapshot{
		CustomValues: make(map[string]float64),
	}

	for _, b := range w.buckets {
		snap.TotalRequests += b.TotalRequests
		snap.ErrorRequests += b.ErrorRequests
		snap.TimeoutRequests += b.TimeoutRequests
		snap.TotalDuration += b.TotalDuration

		for name, val := range b.CustomValues {
			snap.CustomValues[name] = val
		}
	}

	return snap
}

func (w *slidingWindow) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buckets = make([]*bucket, 0)
}
