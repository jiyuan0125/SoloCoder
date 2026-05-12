package store

import (
	"sync"

	"reverse-proxy/types"
)

type Store struct {
	backends     map[string]*types.Backend
	hostToBackend map[string]*types.Backend
	mu           sync.RWMutex
}

func New() *Store {
	return &Store{
		backends:      make(map[string]*types.Backend),
		hostToBackend: make(map[string]*types.Backend),
	}
}

func (s *Store) Add(backend *types.Backend) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backends[backend.ID] = backend
	for _, host := range backend.Hosts {
		s.hostToBackend[host] = backend
	}
}

func (s *Store) Remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	backend, ok := s.backends[id]
	if !ok {
		return
	}
	for _, host := range backend.Hosts {
		delete(s.hostToBackend, host)
	}
	delete(s.backends, id)
}

func (s *Store) GetByHost(host string) *types.Backend {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hostToBackend[host]
}

func (s *Store) GetAll() []*types.Backend {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*types.Backend, 0, len(s.backends))
	for _, b := range s.backends {
		result = append(result, b)
	}
	return result
}

func (s *Store) GetByID(id string) *types.Backend {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.backends[id]
}
