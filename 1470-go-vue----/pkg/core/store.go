package core

import (
	"sync"

	"settlement/pkg/common"
)

type Store struct {
	projects     map[string]*common.Project
	boqs         map[string][]*common.BillOfQuantity
	visas        map[string][]*common.Visa
	visaCounters map[string]int
	logs         map[string][]*common.AuditLog
	settlements  map[string]*common.Settlement
	mu           sync.RWMutex
	projectLocks map[string]*sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		projects:     make(map[string]*common.Project),
		boqs:         make(map[string][]*common.BillOfQuantity),
		visas:        make(map[string][]*common.Visa),
		visaCounters: make(map[string]int),
		logs:         make(map[string][]*common.AuditLog),
		settlements:  make(map[string]*common.Settlement),
		projectLocks: make(map[string]*sync.RWMutex),
	}
}

func (s *Store) getProjectLock(projectID string) *sync.RWMutex {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.projectLocks[projectID]; !ok {
		s.projectLocks[projectID] = &sync.RWMutex{}
	}
	return s.projectLocks[projectID]
}

func (s *Store) addLog(projectID, operator string, action common.AuditAction, desc string) {
	log := &common.AuditLog{
		ID:          generateID(),
		ProjectID:   projectID,
		Action:      action,
		Operator:    operator,
		Description: desc,
		CreatedAt:   now(),
	}
	s.logs[projectID] = append(s.logs[projectID], log)
}

func (s *Store) checkProjectLocked(projectID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.projects[projectID]
	if !ok {
		return false
	}
	return p.Status == common.ProjectStatusLocked
}
