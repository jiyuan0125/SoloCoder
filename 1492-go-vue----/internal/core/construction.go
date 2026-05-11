package core

import (
	"errors"
	"sort"
	"time"

	"renovation-management/internal/api"
)

func (s *Store) CreatePhase(req *api.CreatePhaseRequest) (*api.ConstructionPhase, error) {
	if req.QuotationID == "" {
		return nil, errors.New("quotation_id is required")
	}
	if req.Category == "" {
		return nil, errors.New("category is required")
	}
	if req.ExpectedEnd.Before(req.ExpectedStart) {
		return nil, errors.New("expected end must be after expected start")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.quotations[req.QuotationID]; !ok {
		return nil, errors.New("quotation not found")
	}

	phase := &api.ConstructionPhase{
		ID:            GenerateID("PHASE"),
		QuotationID:   req.QuotationID,
		Category:      req.Category,
		OrderIndex:    req.OrderIndex,
		ExpectedStart: req.ExpectedStart,
		ExpectedEnd:   req.ExpectedEnd,
		Progress:      0,
		DelayAlerted:  false,
	}

	s.phases[phase.ID] = phase
	s.phasesByQuotation[req.QuotationID] = append(
		s.phasesByQuotation[req.QuotationID],
		phase,
	)

	return phase, nil
}

func (s *Store) getPhasesSorted(quotationID string) []*api.ConstructionPhase {
	phases := s.phasesByQuotation[quotationID]
	sorted := make([]*api.ConstructionPhase, len(phases))
	copy(sorted, phases)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].OrderIndex < sorted[j].OrderIndex
	})
	return sorted
}

func (s *Store) canUpdatePhase(phase *api.ConstructionPhase, quotationID string) bool {
	phases := s.getPhasesSorted(quotationID)

	for _, p := range phases {
		if p.ID == phase.ID {
			break
		}
		if p.Progress < 100 {
			return false
		}
	}

	return true
}

func (s *Store) checkDelay(phase *api.ConstructionPhase) bool {
	if phase.ActualEnd != nil {
		return false
	}

	now := time.Now()
	if now.After(phase.ExpectedEnd.AddDate(0, 0, 3)) {
		return true
	}
	return false
}

func (s *Store) UpdateProgress(req *api.UpdateProgressRequest) (*api.ConstructionPhase, error) {
	if req.PhaseID == "" {
		return nil, errors.New("phase_id is required")
	}
	if req.UpdatedBy == "" {
		return nil, errors.New("updated_by is required")
	}
	if req.Progress < 0 || req.Progress > 100 {
		return nil, errors.New("progress must be between 0 and 100")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	phase, ok := s.phases[req.PhaseID]
	if !ok {
		return nil, errors.New("phase not found")
	}

	if !s.canUpdatePhase(phase, phase.QuotationID) {
		return nil, errors.New("previous phase not completed")
	}

	phase.Progress = req.Progress
	phase.LastUpdatedBy = req.UpdatedBy

	if req.ActualStart != nil {
		phase.ActualStart = req.ActualStart
	}
	if req.ActualEnd != nil {
		phase.ActualEnd = req.ActualEnd
	}

	if s.checkDelay(phase) && !phase.DelayAlerted {
		phase.DelayAlerted = true
	}

	return phase, nil
}

func (s *Store) GetPhase(id string) (*api.ConstructionPhase, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	phase, ok := s.phases[id]
	if !ok {
		return nil, errors.New("phase not found")
	}
	return phase, nil
}

func (s *Store) ListPhasesByQuotation(quotationID string) []*api.ConstructionPhase {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.getPhasesSorted(quotationID)
}

func (s *Store) ListDelayedPhases() []*api.ConstructionPhase {
	s.mu.RLock()
	defer s.mu.RUnlock()

	delayed := make([]*api.ConstructionPhase, 0)
	for _, phase := range s.phases {
		if phase.DelayAlerted {
			delayed = append(delayed, phase)
		}
	}
	return delayed
}
