package core

import (
	"sync"
)

type Store struct {
	mu     sync.RWMutex
	fences map[string]*Fence
	bikes  map[string]*Bike
	rides  map[string]*Ride
	tasks  map[string]*DispatchTask
}

func NewStore() *Store {
	return &Store{
		fences: make(map[string]*Fence),
		bikes:  make(map[string]*Bike),
		rides:  make(map[string]*Ride),
		tasks:  make(map[string]*DispatchTask),
	}
}

func (s *Store) GetFences() []*Fence {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Fence, 0, len(s.fences))
	for _, f := range s.fences {
		result = append(result, f)
	}
	return result
}

func (s *Store) GetFence(id string) *Fence {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fences[id]
}

func (s *Store) saveFence(f *Fence) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fences[f.ID] = f
}

func (s *Store) removeFence(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.fences, id)
}

func (s *Store) GetBikes() []*Bike {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Bike, 0, len(s.bikes))
	for _, b := range s.bikes {
		result = append(result, b)
	}
	return result
}

func (s *Store) GetBike(id string) *Bike {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bikes[id]
}

func (s *Store) saveBike(b *Bike) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bikes[b.ID] = b
}

func (s *Store) GetRides() []*Ride {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Ride, 0, len(s.rides))
	for _, r := range s.rides {
		result = append(result, r)
	}
	return result
}

func (s *Store) GetRide(id string) *Ride {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rides[id]
}

func (s *Store) saveRide(r *Ride) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rides[r.ID] = r
}

func (s *Store) GetTasks() []*DispatchTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*DispatchTask, 0, len(s.tasks))
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result
}

func (s *Store) GetTask(id string) *DispatchTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.tasks[id]
}

func (s *Store) saveTask(t *DispatchTask) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.ID] = t
}
