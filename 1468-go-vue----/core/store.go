package core

import (
	"sync"

	"tender-management/common"
)

type Store struct {
	projectsMu  sync.RWMutex
	projects    map[string]*common.Project
	bidsMu      sync.RWMutex
	bids        map[string][]*common.Bid
	bidIndex    map[string]*common.Bid
	openingMu   sync.RWMutex
	opening     map[string]*common.OpeningResult
}

func NewStore() *Store {
	return &Store{
		projects: make(map[string]*common.Project),
		bids:     make(map[string][]*common.Bid),
		bidIndex: make(map[string]*common.Bid),
		opening:  make(map[string]*common.OpeningResult),
	}
}

func (s *Store) SaveProject(project *common.Project) {
	s.projectsMu.Lock()
	defer s.projectsMu.Unlock()
	s.projects[project.ID] = project
}

func (s *Store) GetProject(id string) (*common.Project, bool) {
	s.projectsMu.RLock()
	defer s.projectsMu.RUnlock()
	p, ok := s.projects[id]
	if !ok {
		return nil, false
	}
	copy := *p
	return &copy, true
}

func (s *Store) ListProjects() []common.Project {
	s.projectsMu.RLock()
	defer s.projectsMu.RUnlock()
	list := make([]common.Project, 0, len(s.projects))
	for _, p := range s.projects {
		list = append(list, *p)
	}
	return list
}

func (s *Store) SaveBid(bid *common.Bid) {
	s.bidsMu.Lock()
	defer s.bidsMu.Unlock()
	s.bidIndex[bid.ID] = bid
	existing := s.bids[bid.ProjectID]
	found := false
	for i, b := range existing {
		if b.ID == bid.ID {
			existing[i] = bid
			found = true
			break
		}
	}
	if !found {
		s.bids[bid.ProjectID] = append(existing, bid)
	}
}

func (s *Store) GetBid(id string) (*common.Bid, bool) {
	s.bidsMu.RLock()
	defer s.bidsMu.RUnlock()
	b, ok := s.bidIndex[id]
	if !ok {
		return nil, false
	}
	copy := *b
	return &copy, true
}

func (s *Store) ListBidsByProject(projectID string) []common.Bid {
	s.bidsMu.RLock()
	defer s.bidsMu.RUnlock()
	bids := s.bids[projectID]
	list := make([]common.Bid, len(bids))
	for i, b := range bids {
		list[i] = *b
	}
	return list
}

func (s *Store) GetBidBySupplier(projectID, supplierID string) (*common.Bid, bool) {
	s.bidsMu.RLock()
	defer s.bidsMu.RUnlock()
	bids := s.bids[projectID]
	for _, b := range bids {
		if b.SupplierID == supplierID {
			copy := *b
			return &copy, true
		}
	}
	return nil, false
}

func (s *Store) SaveOpeningResult(result *common.OpeningResult) {
	s.openingMu.Lock()
	defer s.openingMu.Unlock()
	s.opening[result.ProjectID] = result
}

func (s *Store) GetOpeningResult(projectID string) (*common.OpeningResult, bool) {
	s.openingMu.RLock()
	defer s.openingMu.RUnlock()
	r, ok := s.opening[projectID]
	if !ok {
		return nil, false
	}
	copy := *r
	return &copy, true
}
