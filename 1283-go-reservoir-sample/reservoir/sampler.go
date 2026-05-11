package reservoir

import (
	"math/rand"
	"sync"
	"time"
)

type Sampler struct {
	mu    sync.RWMutex
	data  []string
	k     int
	rand  *rand.Rand
}

func NewSampler(k int) *Sampler {
	return &Sampler{
		data: make([]string, 0),
		k:    k,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *Sampler) Add(item string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append(s.data, item)
}

func (s *Sampler) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func (s *Sampler) SetK(k int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.k = k
}

func (s *Sampler) GetK() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.k
}

func (s *Sampler) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make([]string, 0)
}

func (s *Sampler) Sample(k int) []string {
	if k <= 0 {
		return []string{}
	}

	s.mu.RLock()
	data := make([]string, len(s.data))
	copy(data, s.data)
	s.mu.RUnlock()

	if len(data) == 0 {
		return []string{}
	}

	if len(data) <= k {
		result := make([]string, len(data))
		copy(result, data)
		return result
	}

	result := make([]string, k)
	copy(result, data[:k])

	for i := k; i < len(data); i++ {
		prob := float64(k) / float64(i+1)
		if s.rand.Float64() < prob {
			idx := s.rand.Intn(k)
			result[idx] = data[i]
		}
	}

	return result
}

func (s *Sampler) SampleWithDefault() []string {
	return s.Sample(s.GetK())
}
