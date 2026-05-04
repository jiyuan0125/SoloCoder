package circuitbreaker

import "sync"

type SlidingWindow struct {
	mu       sync.RWMutex
	records  []WindowRecord
	size     int
	failures int
	successes int
}

func NewSlidingWindow(size int) *SlidingWindow {
	return &SlidingWindow{
		records: make([]WindowRecord, 0, size),
		size:    size,
	}
}

func (w *SlidingWindow) Add(result RequestResult) {
	w.mu.Lock()
	defer w.mu.Unlock()

	record := WindowRecord{
		Result: result,
	}

	if len(w.records) >= w.size {
		removed := w.records[0]
		w.records = w.records[1:]
		if removed.Result == ResultFailure {
			w.failures--
		} else {
			w.successes--
		}
	}

	w.records = append(w.records, record)
	if result == ResultFailure {
		w.failures++
	} else {
		w.successes++
	}
}

func (w *SlidingWindow) Failures() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.failures
}

func (w *SlidingWindow) Successes() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.successes
}

func (w *SlidingWindow) Total() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.records)
}

func (w *SlidingWindow) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.records = w.records[:0]
	w.failures = 0
	w.successes = 0
}

func (w *SlidingWindow) GetRecords() []WindowRecord {
	w.mu.RLock()
	defer w.mu.RUnlock()
	result := make([]WindowRecord, len(w.records))
	copy(result, w.records)
	return result
}

func (w *SlidingWindow) Size() int {
	return w.size
}
