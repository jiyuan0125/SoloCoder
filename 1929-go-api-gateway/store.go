package main

import (
	"sync"
)

type Store struct {
	mu       sync.RWMutex
	backends map[string]*Backend
	plans    map[string]*Plan
}

func NewStore() *Store {
	return &Store{
		backends: make(map[string]*Backend),
		plans:    make(map[string]*Plan),
	}
}

func (s *Store) AddBackend(backend *Backend) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backends[backend.ID] = backend
}

func (s *Store) GetBackend(id string) (*Backend, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.backends[id]
	return b, ok
}

func (s *Store) ListBackends() []*Backend {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Backend, 0, len(s.backends))
	for _, b := range s.backends {
		result = append(result, b)
	}
	return result
}

func (s *Store) AddPlan(plan *Plan) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.plans[plan.ID] = plan
}

func (s *Store) GetPlan(id string) (*Plan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.plans[id]
	return p, ok
}

func (s *Store) ListPlans() []*Plan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*Plan, 0, len(s.plans))
	for _, p := range s.plans {
		result = append(result, p)
	}
	return result
}
