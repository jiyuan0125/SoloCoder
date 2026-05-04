package main

import (
	"encoding/json"
	"os"
	"school-enrollment/common"
	"sync"
	"time"
)

type Store struct {
	plans         map[string]*common.EnrollmentPlan
	registrations map[string]*common.Registration
	idCardIndex   map[string]string
	mu            sync.RWMutex
	persistFile   string
}

func NewStore(persistFile string) *Store {
	s := &Store{
		plans:         make(map[string]*common.EnrollmentPlan),
		registrations: make(map[string]*common.Registration),
		idCardIndex:   make(map[string]string),
		persistFile:   persistFile,
	}
	s.load()
	return s
}

func (s *Store) CreatePlan(plan *common.EnrollmentPlan) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan.TotalCapacity = plan.Classes * plan.MaxPerClass
	plan.EnrolledCount = 0
	plan.IsOpen = true
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()
	s.plans[plan.ID] = plan
	return s.save()
}

func (s *Store) GetPlan(id string) *common.EnrollmentPlan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.plans[id]
}

func (s *Store) ListPlans() []*common.EnrollmentPlan {
	s.mu.RLock()
	defer s.mu.RUnlock()
	plans := make([]*common.EnrollmentPlan, 0, len(s.plans))
	for _, p := range s.plans {
		plans = append(plans, p)
	}
	return plans
}

func (s *Store) UpdatePlanCapacity(planID string, classes, maxPerClass int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan, exists := s.plans[planID]
	if !exists {
		return nil
	}
	newCapacity := classes * maxPerClass
	if newCapacity < plan.EnrolledCount {
		return nil
	}
	plan.Classes = classes
	plan.MaxPerClass = maxPerClass
	plan.TotalCapacity = newCapacity
	plan.UpdatedAt = time.Now()
	return s.save()
}

func (s *Store) ClosePlan(planID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan, exists := s.plans[planID]
	if !exists {
		return nil
	}
	plan.IsOpen = false
	plan.UpdatedAt = time.Now()
	return s.save()
}

func (s *Store) CreateRegistration(reg *common.Registration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan, exists := s.plans[reg.PlanID]
	if !exists {
		return nil
	}
	if !plan.IsOpen {
		return nil
	}
	if plan.EnrolledCount >= plan.TotalCapacity {
		return nil
	}
	indexKey := reg.PlanID + ":" + reg.IDCard
	if _, exists := s.idCardIndex[indexKey]; exists {
		return nil
	}
	reg.Status = common.StatusPending
	reg.SubmittedAt = time.Now()
	s.registrations[reg.ID] = reg
	s.idCardIndex[indexKey] = reg.ID
	plan.EnrolledCount++
	plan.UpdatedAt = time.Now()
	return s.save()
}

func (s *Store) GetRegistration(id string) *common.Registration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.registrations[id]
}

func (s *Store) GetRegistrationByIDCard(planID, idCard string) *common.Registration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	indexKey := planID + ":" + idCard
	regID, exists := s.idCardIndex[indexKey]
	if !exists {
		return nil
	}
	return s.registrations[regID]
}

func (s *Store) ListRegistrations(planID string, status string) []*common.Registration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var results []*common.Registration
	for _, reg := range s.registrations {
		if reg.PlanID == planID && (status == "" || reg.Status == status) {
			results = append(results, reg)
		}
	}
	return results
}

func (s *Store) AcceptRegistration(regID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	reg, exists := s.registrations[regID]
	if !exists {
		return nil
	}
	if reg.Status != common.StatusPending {
		return nil
	}
	reg.Status = common.StatusAccepted
	reg.ReviewedAt = time.Now()
	return s.save()
}

func (s *Store) RejectRegistration(regID, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	reg, exists := s.registrations[regID]
	if !exists {
		return nil
	}
	if reg.Status != common.StatusPending {
		return nil
	}
	reg.Status = common.StatusRejected
	reg.RejectReason = reason
	reg.ReviewedAt = time.Now()
	plan := s.plans[reg.PlanID]
	if plan != nil && plan.EnrolledCount > 0 {
		plan.EnrolledCount--
		plan.UpdatedAt = time.Now()
	}
	return s.save()
}

type storeData struct {
	Plans         map[string]*common.EnrollmentPlan `json:"plans"`
	Registrations map[string]*common.Registration   `json:"registrations"`
	IDCardIndex   map[string]string                 `json:"id_card_index"`
}

func (s *Store) save() error {
	data := storeData{
		Plans:         s.plans,
		Registrations: s.registrations,
		IDCardIndex:   s.idCardIndex,
	}
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.persistFile, bytes, 0644)
}

func (s *Store) load() error {
	bytes, err := os.ReadFile(s.persistFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var data storeData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return err
	}
	if data.Plans != nil {
		s.plans = data.Plans
	}
	if data.Registrations != nil {
		s.registrations = data.Registrations
	}
	if data.IDCardIndex != nil {
		s.idCardIndex = data.IDCardIndex
	}
	return nil
}
