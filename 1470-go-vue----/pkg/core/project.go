package core

import (
	"errors"

	"settlement/pkg/common"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrProjectLocked       = errors.New("project is locked, cannot modify")
	ErrDuplicateItemCode   = errors.New("item code already exists in this project")
	ErrInvalidContractQty  = errors.New("contract quantity must be greater than zero")
	ErrNegativeQty         = errors.New("quantity cannot be negative")
)

func (s *Store) CreateProject(name string, contractAmount float64, operator string) (*common.Project, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := &common.Project{
		ID:             generateID(),
		Name:           name,
		ContractAmount: roundToCents(contractAmount),
		Status:         common.ProjectStatusDraft,
		CreatedAt:      now(),
		UpdatedAt:      now(),
	}

	s.projects[p.ID] = p
	s.boqs[p.ID] = []*common.BillOfQuantity{}
	s.visas[p.ID] = []*common.Visa{}
	s.visaCounters[p.ID] = 0
	s.logs[p.ID] = []*common.AuditLog{}

	s.addLog(p.ID, operator, common.AuditActionCreate, "创建项目: "+name)
	return p, nil
}

func (s *Store) GetProject(projectID string) (*common.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.projects[projectID]
	if !ok {
		return nil, ErrProjectNotFound
	}
	return p, nil
}

func (s *Store) ListProjects() []*common.Project {
	s.mu.RLock()
	defer s.mu.RUnlock()

	projects := make([]*common.Project, 0, len(s.projects))
	for _, p := range s.projects {
		projects = append(projects, p)
	}
	return projects
}

func (s *Store) AddBOQ(projectID, itemCode, itemName, unit string, contractQty, actualQty, unitPrice float64, operator string) (*common.BillOfQuantity, error) {
	plock := s.getProjectLock(projectID)
	plock.Lock()
	defer plock.Unlock()

	s.mu.Lock()
	if s.checkProjectLocked(projectID) {
		s.mu.Unlock()
		return nil, ErrProjectLocked
	}

	if _, ok := s.projects[projectID]; !ok {
		s.mu.Unlock()
		return nil, ErrProjectNotFound
	}

	if contractQty <= 0 {
		s.mu.Unlock()
		return nil, ErrInvalidContractQty
	}
	if actualQty < 0 {
		s.mu.Unlock()
		return nil, ErrNegativeQty
	}

	for _, boq := range s.boqs[projectID] {
		if boq.ItemCode == itemCode {
			s.mu.Unlock()
			return nil, ErrDuplicateItemCode
		}
	}

	devRate := ((actualQty - contractQty) / contractQty) * 100
	isAbnormal := devRate < -5 || devRate > 5

	contractAmount := roundToCents(contractQty * unitPrice)

	var adjustment float64
	if devRate > 5 {
		adjustment = roundToCents(contractAmount * (devRate - 5) / 100)
	} else if devRate < -5 {
		adjustment = roundToCents(contractAmount * (devRate + 5) / 100)
	} else {
		adjustment = 0
	}

	boq := &common.BillOfQuantity{
		ID:             generateID(),
		ProjectID:      projectID,
		ItemCode:       itemCode,
		ItemName:       itemName,
		Unit:           unit,
		ContractQty:    contractQty,
		ActualQty:      actualQty,
		DeviationRate:  roundToCents(devRate),
		IsAbnormal:     isAbnormal,
		UnitPrice:      roundToCents(unitPrice),
		ContractAmount: contractAmount,
		Adjustment:     adjustment,
	}

	s.boqs[projectID] = append(s.boqs[projectID], boq)
	s.projects[projectID].UpdatedAt = now()
	s.addLog(projectID, operator, common.AuditActionCreate, "添加工程量清单: "+itemCode)
	s.mu.Unlock()

	return boq, nil
}

func (s *Store) UpdateBOQ(projectID, boqID string, req *common.UpdateBOQRequest, operator string) (*common.BillOfQuantity, error) {
	plock := s.getProjectLock(projectID)
	plock.Lock()
	defer plock.Unlock()

	s.mu.Lock()
	if s.checkProjectLocked(projectID) {
		s.mu.Unlock()
		return nil, ErrProjectLocked
	}

	if _, ok := s.projects[projectID]; !ok {
		s.mu.Unlock()
		return nil, ErrProjectNotFound
	}

	var target *common.BillOfQuantity
	idx := -1
	for i, b := range s.boqs[projectID] {
		if b.ID == boqID {
			target = b
			idx = i
			break
		}
	}

	if target == nil {
		s.mu.Unlock()
		return nil, errors.New("BOQ not found")
	}

	if req.ItemName != "" {
		target.ItemName = req.ItemName
	}
	if req.Unit != "" {
		target.Unit = req.Unit
	}
	if req.ContractQty > 0 {
		target.ContractQty = req.ContractQty
	}
	if req.ActualQty >= 0 {
		target.ActualQty = req.ActualQty
	}
	if req.UnitPrice > 0 {
		target.UnitPrice = roundToCents(req.UnitPrice)
	}

	if target.ContractQty <= 0 {
		s.mu.Unlock()
		return nil, ErrInvalidContractQty
	}
	if target.ActualQty < 0 {
		s.mu.Unlock()
		return nil, ErrNegativeQty
	}

	devRate := ((target.ActualQty - target.ContractQty) / target.ContractQty) * 100
	target.DeviationRate = roundToCents(devRate)
	target.IsAbnormal = devRate < -5 || devRate > 5
	target.ContractAmount = roundToCents(target.ContractQty * target.UnitPrice)

	if devRate > 5 {
		target.Adjustment = roundToCents(target.ContractAmount * (devRate - 5) / 100)
	} else if devRate < -5 {
		target.Adjustment = roundToCents(target.ContractAmount * (devRate + 5) / 100)
	} else {
		target.Adjustment = 0
	}

	s.boqs[projectID][idx] = target
	s.projects[projectID].UpdatedAt = now()
	s.addLog(projectID, operator, common.AuditActionUpdate, "更新工程量清单: "+target.ItemCode)
	s.mu.Unlock()

	return target, nil
}

func (s *Store) DeleteBOQ(projectID, boqID, operator string) error {
	plock := s.getProjectLock(projectID)
	plock.Lock()
	defer plock.Unlock()

	s.mu.Lock()
	if s.checkProjectLocked(projectID) {
		s.mu.Unlock()
		return ErrProjectLocked
	}

	if _, ok := s.projects[projectID]; !ok {
		s.mu.Unlock()
		return ErrProjectNotFound
	}

	idx := -1
	var itemCode string
	for i, b := range s.boqs[projectID] {
		if b.ID == boqID {
			idx = i
			itemCode = b.ItemCode
			break
		}
	}

	if idx == -1 {
		s.mu.Unlock()
		return errors.New("BOQ not found")
	}

	s.boqs[projectID] = append(s.boqs[projectID][:idx], s.boqs[projectID][idx+1:]...)
	s.projects[projectID].UpdatedAt = now()
	s.addLog(projectID, operator, common.AuditActionDelete, "删除工程量清单: "+itemCode)
	s.mu.Unlock()

	return nil
}

func (s *Store) ListBOQ(projectID string) ([]*common.BillOfQuantity, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.projects[projectID]; !ok {
		return nil, ErrProjectNotFound
	}

	return s.boqs[projectID], nil
}
