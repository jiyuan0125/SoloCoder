package store

import (
	"abtest/api"
	"sync"
	"time"
)

type UserMetrics struct {
	UserID       string
	Converted    bool
	StayDuration float64
	RecordedAt   time.Time
}

type ExperimentMetrics struct {
	ControlUsers   map[string]*UserMetrics
	TreatmentUsers map[string]*UserMetrics
}

type Store struct {
	mu          sync.RWMutex
	experiments map[string]*api.Experiment
	assignments map[string]map[string]api.GroupType
	metrics     map[string]*ExperimentMetrics
}

func NewStore() *Store {
	return &Store{
		experiments: make(map[string]*api.Experiment),
		assignments: make(map[string]map[string]api.GroupType),
		metrics:     make(map[string]*ExperimentMetrics),
	}
}

func (s *Store) CreateExperiment(exp *api.Experiment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.experiments[exp.ID]; exists {
		return nil
	}

	s.experiments[exp.ID] = exp
	s.assignments[exp.ID] = make(map[string]api.GroupType)
	s.metrics[exp.ID] = &ExperimentMetrics{
		ControlUsers:   make(map[string]*UserMetrics),
		TreatmentUsers: make(map[string]*UserMetrics),
	}
	return nil
}

func (s *Store) GetExperiment(id string) (*api.Experiment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	exp, exists := s.experiments[id]
	if !exists {
		return nil, false
	}
	copyExp := *exp
	return &copyExp, true
}

func (s *Store) UpdateExperiment(exp *api.Experiment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.experiments[exp.ID]; !exists {
		return nil
	}

	s.experiments[exp.ID] = exp
	return nil
}

func (s *Store) ListExperiments(status *api.ExperimentStatus) []api.Experiment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []api.Experiment
	for _, exp := range s.experiments {
		if status == nil || exp.Status == *status {
			result = append(result, *exp)
		}
	}
	return result
}

func (s *Store) GetOrCreateAssignment(experimentID, userID string, assign func() api.GroupType) api.GroupType {
	s.mu.Lock()
	defer s.mu.Unlock()

	if assignments, exists := s.assignments[experimentID]; exists {
		if group, assigned := assignments[userID]; assigned {
			return group
		}
		group := assign()
		assignments[userID] = group
		return group
	}
	return ""
}

func (s *Store) GetAssignment(experimentID, userID string) (api.GroupType, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if assignments, exists := s.assignments[experimentID]; exists {
		if group, assigned := assignments[userID]; assigned {
			return group, true
		}
	}
	return "", false
}

func (s *Store) GetUserAssignments(userID string) map[string]api.GroupType {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]api.GroupType)
	for expID, assignments := range s.assignments {
		if group, exists := assignments[userID]; exists {
			result[expID] = group
		}
	}
	return result
}

func (s *Store) RecordMetrics(experimentID string, group api.GroupType, userID string, converted bool, stayDuration float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metrics, exists := s.metrics[experimentID]
	if !exists {
		return
	}

	userMetrics := &UserMetrics{
		UserID:       userID,
		Converted:    converted,
		StayDuration: stayDuration,
		RecordedAt:   time.Now(),
	}

	if group == api.GroupControl {
		metrics.ControlUsers[userID] = userMetrics
	} else {
		metrics.TreatmentUsers[userID] = userMetrics
	}
}

func (s *Store) GetMetrics(experimentID string) (control []*UserMetrics, treatment []*UserMetrics) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics, exists := s.metrics[experimentID]
	if !exists {
		return nil, nil
	}

	for _, m := range metrics.ControlUsers {
		control = append(control, m)
	}
	for _, m := range metrics.TreatmentUsers {
		treatment = append(treatment, m)
	}

	return control, treatment
}

func (s *Store) GetUserMetrics(experimentID, userID string) *UserMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics, exists := s.metrics[experimentID]
	if !exists {
		return nil
	}

	if m, exists := metrics.ControlUsers[userID]; exists {
		return m
	}
	if m, exists := metrics.TreatmentUsers[userID]; exists {
		return m
	}
	return nil
}
