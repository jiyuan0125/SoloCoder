package topk

import (
	"container/heap"
	"sort"
	"sync"

	"topk-service/pkg/api"
)

type TopKService struct {
	mu        sync.RWMutex
	k         int
	freqMap   map[float64]int
}

func NewTopKService(k int) *TopKService {
	if k < 0 {
		k = 0
	}
	return &TopKService{
		k:       k,
		freqMap: make(map[float64]int),
	}
}

func (s *TopKService) Add(values []float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	for _, v := range values {
		s.freqMap[v]++
	}
}

func (s *TopKService) Query() []api.ElementFreq {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.k == 0 || len(s.freqMap) == 0 {
		return []api.ElementFreq{}
	}
	
	elements := make([]elementWithFreq, 0, len(s.freqMap))
	for elem, freq := range s.freqMap {
		elements = append(elements, elementWithFreq{element: elem, freq: freq})
	}
	
	sort.Slice(elements, func(i, j int) bool {
		if elements[i].freq == elements[j].freq {
			return elements[i].element < elements[j].element
		}
		return elements[i].freq < elements[j].freq
	})
	
	limit := s.k
	if len(elements) < s.k {
		limit = len(elements)
	}
	
	result := make([]api.ElementFreq, 0, limit)
	for i := 0; i < limit; i++ {
		result = append(result, api.ElementFreq{
			Element: elements[i].element,
			Freq:    elements[i].freq,
		})
	}
	
	return result
}

func (s *TopKService) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.freqMap = make(map[float64]int)
}

func (s *TopKService) SetK(k int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if k < 0 {
		k = 0
	}
	s.k = k
}

func (s *TopKService) GetK() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.k
}

func (s *TopKService) QueryWithHeap() []api.ElementFreq {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.k == 0 || len(s.freqMap) == 0 {
		return []api.ElementFreq{}
	}
	
	h := &maxHeap{}
	heap.Init(h)
	
	for elem, freq := range s.freqMap {
		ef := elementWithFreq{element: elem, freq: freq}
		if h.Len() < s.k {
			heap.Push(h, ef)
		} else {
			top := (*h)[0]
			if ef.freq < top.freq || (ef.freq == top.freq && ef.element < top.element) {
				heap.Pop(h)
				heap.Push(h, ef)
			}
		}
	}
	
	result := make([]api.ElementFreq, 0, h.Len())
	for h.Len() > 0 {
		ef := heap.Pop(h).(elementWithFreq)
		result = append(result, api.ElementFreq{
			Element: ef.element,
			Freq:    ef.freq,
		})
	}
	
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	
	return result
}
