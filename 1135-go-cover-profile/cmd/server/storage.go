package main

import (
	"sync"

	"github.com/solocoder/coverage-analyzer/pkg/coverage"
)

type AnalysisStore struct {
	mu       sync.RWMutex
	analyses map[string]*coverage.AnalysisResult
}

func NewAnalysisStore() *AnalysisStore {
	return &AnalysisStore{
		analyses: make(map[string]*coverage.AnalysisResult),
	}
}

func (s *AnalysisStore) Save(id string, result *coverage.AnalysisResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.analyses[id] = result
}

func (s *AnalysisStore) Get(id string) (*coverage.AnalysisResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result, exists := s.analyses[id]
	return result, exists
}

func (s *AnalysisStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.analyses, id)
}
