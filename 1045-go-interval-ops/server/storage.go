package main

import (
	"intervalops/common"
	"sync"
	"time"
)

type BatchStore struct {
	mu    sync.RWMutex
	store map[string]*common.BatchResult
}

func NewBatchStore() *BatchStore {
	return &BatchStore{
		store: make(map[string]*common.BatchResult),
	}
}

func (s *BatchStore) Save(batchID string, req common.OperationRequest, result common.OperationResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[batchID] = &common.BatchResult{
		BatchID:   batchID,
		Request:   req,
		Result:    result,
		CreatedAt: time.Now().UTC(),
	}
}

func (s *BatchStore) Get(batchID string) (*common.BatchResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res, ok := s.store[batchID]
	return res, ok
}
