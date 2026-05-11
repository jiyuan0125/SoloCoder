package core

import (
	"sync"
	"time"
)

type Store interface {
	CreateCustomer(c *Customer) error
	GetCustomer(id string) (*Customer, error)
	ListCustomers() ([]*Customer, error)
	UpdateCustomer(c *Customer) error

	CreateContract(c *Contract) error
	GetContract(id string) (*Contract, error)
	ListContracts() ([]*Contract, error)
	ListActiveContracts() ([]*Contract, error)

	CreateControlPoint(cp *ControlPoint) error
	GetControlPointsByContract(contractID string) ([]*ControlPoint, error)

	CreateInspection(i *Inspection) error
	GetInspection(id string) (*Inspection, error)
	ListInspectionsByContract(contractID string) ([]*Inspection, error)
	GetLatestInspections(contractID string, limit int) ([]*Inspection, error)

	CreateChemical(c *Chemical) error
	GetChemical(id string) (*Chemical, error)
	ListChemicals() ([]*Chemical, error)
	UpdateChemical(c *Chemical) error

	CreateStaff(s *Staff) error
	GetStaff(id string) (*Staff, error)
	ListActiveStaff() ([]*Staff, error)

	CreateOperation(o *Operation) error
	GetOperation(id string) (*Operation, error)
	ListOperationsByContract(contractID string) ([]*Operation, error)
	ListOperationsByStaff(staffID string, date time.Time) ([]*Operation, error)
	GetLastOperation(contractID string) (*Operation, error)

	CreateEffectEvaluation(e *EffectEvaluation) error
	GetEffectEvaluation(inspectionID string) (*EffectEvaluation, error)
}

type InMemoryStore struct {
	mu         sync.RWMutex
	customers  map[string]*Customer
	contracts  map[string]*Contract
	controlPoints map[string]*ControlPoint
	inspections map[string]*Inspection
	chemicals  map[string]*Chemical
	chemicalMu sync.Mutex
	staffs     map[string]*Staff
	operations map[string]*Operation
	evaluations map[string]*EffectEvaluation
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		customers:     make(map[string]*Customer),
		contracts:     make(map[string]*Contract),
		controlPoints: make(map[string]*ControlPoint),
		inspections:   make(map[string]*Inspection),
		chemicals:     make(map[string]*Chemical),
		staffs:        make(map[string]*Staff),
		operations:    make(map[string]*Operation),
		evaluations:   make(map[string]*EffectEvaluation),
	}
}

func (s *InMemoryStore) CreateCustomer(c *Customer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	s.customers[c.ID] = c
	return nil
}

func (s *InMemoryStore) GetCustomer(id string) (*Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.customers[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *InMemoryStore) ListCustomers() ([]*Customer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Customer, 0, len(s.customers))
	for _, c := range s.customers {
		list = append(list, c)
	}
	return list, nil
}

func (s *InMemoryStore) UpdateCustomer(c *Customer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.customers[c.ID]; !ok {
		return ErrNotFound
	}
	c.UpdatedAt = time.Now()
	s.customers[c.ID] = c
	return nil
}

func (s *InMemoryStore) CreateContract(c *Contract) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	s.contracts[c.ID] = c
	return nil
}

func (s *InMemoryStore) GetContract(id string) (*Contract, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.contracts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *InMemoryStore) ListContracts() ([]*Contract, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Contract, 0, len(s.contracts))
	for _, c := range s.contracts {
		list = append(list, c)
	}
	return list, nil
}

func (s *InMemoryStore) ListActiveContracts() ([]*Contract, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	list := make([]*Contract, 0)
	for _, c := range s.contracts {
		if !c.StartDate.After(now) && !c.EndDate.Before(now) {
			list = append(list, c)
		}
	}
	return list, nil
}

func (s *InMemoryStore) CreateControlPoint(cp *ControlPoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.controlPoints[cp.ID] = cp
	return nil
}

func (s *InMemoryStore) GetControlPointsByContract(contractID string) ([]*ControlPoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*ControlPoint, 0)
	for _, cp := range s.controlPoints {
		if cp.ContractID == contractID {
			list = append(list, cp)
		}
	}
	return list, nil
}

func (s *InMemoryStore) CreateInspection(i *Inspection) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	i.CreatedAt = time.Now()
	s.inspections[i.ID] = i
	return nil
}

func (s *InMemoryStore) GetInspection(id string) (*Inspection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	i, ok := s.inspections[id]
	if !ok {
		return nil, ErrNotFound
	}
	return i, nil
}

func (s *InMemoryStore) ListInspectionsByContract(contractID string) ([]*Inspection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Inspection, 0)
	for _, i := range s.inspections {
		if i.ContractID == contractID {
			list = append(list, i)
		}
	}
	return list, nil
}

func (s *InMemoryStore) GetLatestInspections(contractID string, limit int) ([]*Inspection, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Inspection, 0)
	for _, i := range s.inspections {
		if i.ContractID == contractID {
			list = append(list, i)
		}
	}
	for i := 0; i < len(list)-1; i++ {
		for j := i + 1; j < len(list); j++ {
			if list[i].InspectionDate.Before(list[j].InspectionDate) {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (s *InMemoryStore) CreateChemical(c *Chemical) error {
	s.chemicalMu.Lock()
	defer s.chemicalMu.Unlock()
	s.chemicals[c.ID] = c
	return nil
}

func (s *InMemoryStore) GetChemical(id string) (*Chemical, error) {
	s.chemicalMu.Lock()
	defer s.chemicalMu.Unlock()
	c, ok := s.chemicals[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *InMemoryStore) ListChemicals() ([]*Chemical, error) {
	s.chemicalMu.Lock()
	defer s.chemicalMu.Unlock()
	list := make([]*Chemical, 0, len(s.chemicals))
	for _, c := range s.chemicals {
		list = append(list, c)
	}
	return list, nil
}

func (s *InMemoryStore) UpdateChemical(c *Chemical) error {
	s.chemicalMu.Lock()
	defer s.chemicalMu.Unlock()
	if _, ok := s.chemicals[c.ID]; !ok {
		return ErrNotFound
	}
	s.chemicals[c.ID] = c
	return nil
}

func (s *InMemoryStore) CreateStaff(st *Staff) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.staffs[st.ID] = st
	return nil
}

func (s *InMemoryStore) GetStaff(id string) (*Staff, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.staffs[id]
	if !ok {
		return nil, ErrNotFound
	}
	return st, nil
}

func (s *InMemoryStore) ListActiveStaff() ([]*Staff, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Staff, 0)
	for _, st := range s.staffs {
		if st.Active {
			list = append(list, st)
		}
	}
	return list, nil
}

func (s *InMemoryStore) CreateOperation(o *Operation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	o.CreatedAt = time.Now()
	s.operations[o.ID] = o
	return nil
}

func (s *InMemoryStore) GetOperation(id string) (*Operation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.operations[id]
	if !ok {
		return nil, ErrNotFound
	}
	return o, nil
}

func (s *InMemoryStore) ListOperationsByContract(contractID string) ([]*Operation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Operation, 0)
	for _, o := range s.operations {
		if o.ContractID == contractID {
			list = append(list, o)
		}
	}
	return list, nil
}

func (s *InMemoryStore) ListOperationsByStaff(staffID string, date time.Time) ([]*Operation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*Operation, 0)
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := start.Add(24 * time.Hour)
	for _, o := range s.operations {
		if o.OperatorID == staffID && !o.OperationDate.Before(start) && o.OperationDate.Before(end) {
			list = append(list, o)
		}
	}
	return list, nil
}

func (s *InMemoryStore) GetLastOperation(contractID string) (*Operation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var latest *Operation
	for _, o := range s.operations {
		if o.ContractID == contractID {
			if latest == nil || o.OperationDate.After(latest.OperationDate) {
				latest = o
			}
		}
	}
	return latest, nil
}

func (s *InMemoryStore) CreateEffectEvaluation(e *EffectEvaluation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.evaluations[e.InspectionID] = e
	return nil
}

func (s *InMemoryStore) GetEffectEvaluation(inspectionID string) (*EffectEvaluation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.evaluations[inspectionID]
	if !ok {
		return nil, ErrNotFound
	}
	return e, nil
}
