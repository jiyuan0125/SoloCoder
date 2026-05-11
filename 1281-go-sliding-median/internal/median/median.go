package median

import (
	"container/list"
	"sort"
)

type SlidingMedian struct {
	windowSize int
	data       *list.List
	sorted     []float64
}

func NewSlidingMedian(windowSize int) *SlidingMedian {
	if windowSize <= 0 {
		windowSize = 1
	}
	return &SlidingMedian{
		windowSize: windowSize,
		data:       list.New(),
		sorted:     make([]float64, 0),
	}
}

func (sm *SlidingMedian) Push(value float64) {
	if sm.data.Len() >= sm.windowSize {
		oldest := sm.data.Front()
		sm.data.Remove(oldest)
		sm.removeFromSorted(oldest.Value.(float64))
	}

	sm.data.PushBack(value)
	sm.insertIntoSorted(value)
}

func (sm *SlidingMedian) Median() (float64, bool) {
	if len(sm.sorted) == 0 {
		return 0, false
	}

	n := len(sm.sorted)
	if n%2 == 1 {
		return sm.sorted[n/2], true
	} else {
		return (sm.sorted[n/2-1] + sm.sorted[n/2]) / 2.0, true
	}
}

func (sm *SlidingMedian) SetWindowSize(size int) {
	if size <= 0 {
		size = 1
	}
	sm.windowSize = size

	for sm.data.Len() > sm.windowSize {
		oldest := sm.data.Front()
		sm.data.Remove(oldest)
		sm.removeFromSorted(oldest.Value.(float64))
	}
}

func (sm *SlidingMedian) Reset() {
	sm.data = list.New()
	sm.sorted = make([]float64, 0)
}

func (sm *SlidingMedian) WindowSize() int {
	return sm.windowSize
}

func (sm *SlidingMedian) DataCount() int {
	return sm.data.Len()
}

func (sm *SlidingMedian) insertIntoSorted(value float64) {
	idx := sort.SearchFloat64s(sm.sorted, value)
	sm.sorted = append(sm.sorted, 0)
	copy(sm.sorted[idx+1:], sm.sorted[idx:])
	sm.sorted[idx] = value
}

func (sm *SlidingMedian) removeFromSorted(value float64) {
	idx := sort.SearchFloat64s(sm.sorted, value)
	for idx < len(sm.sorted) && sm.sorted[idx] != value {
		idx++
	}
	if idx < len(sm.sorted) {
		sm.sorted = append(sm.sorted[:idx], sm.sorted[idx+1:]...)
	}
}
