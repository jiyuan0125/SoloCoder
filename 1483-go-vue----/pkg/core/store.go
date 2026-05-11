package core

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"insurance-claim/pkg/common"
)

type Store interface {
	CreateCase(c *common.ClaimCase) error
	GetCase(id string) (*common.ClaimCase, bool)
	GetCaseByNo(caseNo string) (*common.ClaimCase, bool)
	ListCases() []*common.ClaimCase
	ListCasesByPolicy(policyNo string) []*common.ClaimCase
	UpdateCase(c *common.ClaimCase) error
	DeleteCase(id string) error
	GenerateCaseNo(date time.Time) (string, error)
	GetPolicy(policyNo string) (*common.Policy, bool)
	CreatePolicy(p *common.Policy) error
	CountClaimsThisYear(policyNo string, from time.Time) int
	AddStatusHistory(h *common.StatusHistory) error
}

type memoryStore struct {
	mu         sync.RWMutex
	cases      map[string]*common.ClaimCase
	caseNoMap  map[string]*common.ClaimCase
	caseNoSeq  map[string]int
	policies   map[string]*common.Policy
	histories  map[string][]*common.StatusHistory
}

func NewMemoryStore() Store {
	return &memoryStore{
		cases:     make(map[string]*common.ClaimCase),
		caseNoMap: make(map[string]*common.ClaimCase),
		caseNoSeq: make(map[string]int),
		policies:  make(map[string]*common.Policy),
		histories: make(map[string][]*common.StatusHistory),
	}
}

func (s *memoryStore) CreateCase(c *common.ClaimCase) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.cases[c.ID]; exists {
		return errors.New("case already exists")
	}
	s.cases[c.ID] = c
	s.caseNoMap[c.CaseNo] = c
	return nil
}

func (s *memoryStore) GetCase(id string) (*common.ClaimCase, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.cases[id]
	return c, ok
}

func (s *memoryStore) GetCaseByNo(caseNo string) (*common.ClaimCase, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.caseNoMap[caseNo]
	return c, ok
}

func (s *memoryStore) ListCases() []*common.ClaimCase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*common.ClaimCase, 0, len(s.cases))
	for _, c := range s.cases {
		result = append(result, c)
	}
	return result
}

func (s *memoryStore) ListCasesByPolicy(policyNo string) []*common.ClaimCase {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*common.ClaimCase, 0)
	for _, c := range s.cases {
		if c.PolicyNo == policyNo {
			result = append(result, c)
		}
	}
	return result
}

func (s *memoryStore) UpdateCase(c *common.ClaimCase) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.cases[c.ID]; !exists {
		return errors.New("case not found")
	}
	s.cases[c.ID] = c
	return nil
}

func (s *memoryStore) DeleteCase(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.cases[id]
	if !ok {
		return errors.New("case not found")
	}
	delete(s.cases, id)
	delete(s.caseNoMap, c.CaseNo)
	return nil
}

func (s *memoryStore) GenerateCaseNo(date time.Time) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := date.Format("20060102")
	seq := s.caseNoSeq[key] + 1
	s.caseNoSeq[key] = seq
	caseNo := fmt.Sprintf("%s%s%06d", common.CaseNoPrefix, key, seq)
	return caseNo, nil
}

func (s *memoryStore) GetPolicy(policyNo string) (*common.Policy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.policies[policyNo]
	return p, ok
}

func (s *memoryStore) CreatePolicy(p *common.Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.policies[p.PolicyNo]; exists {
		return errors.New("policy already exists")
	}
	s.policies[p.PolicyNo] = p
	return nil
}

func (s *memoryStore) CountClaimsThisYear(policyNo string, from time.Time) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	startOfYear := time.Date(from.Year(), 1, 1, 0, 0, 0, 0, from.Location())
	count := 0
	for _, c := range s.cases {
		if c.PolicyNo == policyNo && c.CreatedAt.After(startOfYear) && !c.CreatedAt.After(from) {
			count++
		}
	}
	return count
}

func (s *memoryStore) AddStatusHistory(h *common.StatusHistory) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.histories[h.CaseID] = append(s.histories[h.CaseID], h)
	if c, ok := s.cases[h.CaseID]; ok {
		c.StatusHistory = append(c.StatusHistory, *h)
	}
	return nil
}
