package core

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	mu           sync.RWMutex
	contracts    map[string]*Contract
	auditLogs    []*AuditLog
	adjustments  []*ContractAdjustment
}

func NewStore() *Store {
	return &Store{
		contracts:   make(map[string]*Contract),
		auditLogs:   []*AuditLog{},
		adjustments: []*ContractAdjustment{},
	}
}

func (s *Store) CreateContract(c *Contract) (*Contract, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c.ID = uuid.New().String()
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt
	s.contracts[c.ID] = c
	return c, nil
}

func (s *Store) GetContract(id string) (*Contract, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.contracts[id]
	if !ok {
		return nil, ErrContractNotFound
	}
	return c, nil
}

func (s *Store) ListContracts() []*Contract {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Contract, 0, len(s.contracts))
	for _, c := range s.contracts {
		list = append(list, c)
	}
	return list
}

func (s *Store) UpdateContract(c *Contract) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.contracts[c.ID]; !ok {
		return ErrContractNotFound
	}
	c.UpdatedAt = time.Now()
	s.contracts[c.ID] = c
	return nil
}

func (s *Store) GetMilestone(contractID, milestoneID string) (*Milestone, error) {
	c, err := s.GetContract(contractID)
	if err != nil {
		return nil, err
	}
	for _, m := range c.Milestones {
		if m.ID == milestoneID {
			return m, nil
		}
	}
	return nil, ErrMilestoneNotFound
}

func (s *Store) AddAuditLog(log *AuditLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.ID = uuid.New().String()
	log.Timestamp = time.Now()
	s.auditLogs = append(s.auditLogs, log)
}

func (s *Store) GetAuditLogs(contractID string) []*AuditLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	logs := []*AuditLog{}
	for _, l := range s.auditLogs {
		if l.ContractID == contractID {
			logs = append(logs, l)
		}
	}
	return logs
}

func (s *Store) AddAdjustment(a *ContractAdjustment) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a.ID = uuid.New().String()
	a.CreatedAt = time.Now()
	s.adjustments = append(s.adjustments, a)
}

func (s *Store) Lock() {
	s.mu.Lock()
}

func (s *Store) Unlock() {
	s.mu.Unlock()
}

func (s *Store) RLock() {
	s.mu.RLock()
}

func (s *Store) RUnlock() {
	s.mu.RUnlock()
}
