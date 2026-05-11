package core

import (
	"settlement/pkg/common"
)

func (s *Store) CalculateSettlement(projectID string) (*common.Settlement, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	project, ok := s.projects[projectID]
	if !ok {
		return nil, ErrProjectNotFound
	}

	var visaIncreaseAmount, visaDecreaseAmount float64
	for _, v := range s.visas[projectID] {
		if v.Status == common.VisaStatusConfirmed {
			if v.Amount > 0 {
				visaIncreaseAmount += v.Amount
			} else {
				visaDecreaseAmount += -v.Amount
			}
		}
	}

	var quantityAdjustment float64
	for _, boq := range s.boqs[projectID] {
		quantityAdjustment += boq.Adjustment
	}

	total := project.ContractAmount + visaIncreaseAmount - visaDecreaseAmount + quantityAdjustment

	settlement := &common.Settlement{
		ProjectID:          projectID,
		ContractAmount:     roundToCents(project.ContractAmount),
		VisaIncreaseAmount: roundToCents(visaIncreaseAmount),
		VisaDecreaseAmount: roundToCents(visaDecreaseAmount),
		QuantityAdjustment: roundToCents(quantityAdjustment),
		TotalAmount:        roundToCents(total),
	}

	if existing, ok := s.settlements[projectID]; ok {
		settlement.AuditOpinion = existing.AuditOpinion
		settlement.AuditedAt = existing.AuditedAt
		settlement.Auditor = existing.Auditor
	}

	return settlement, nil
}

func (s *Store) AddAuditOpinion(projectID, auditor, opinion string) error {
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

	settlement, err := s.CalculateSettlement(projectID)
	if err != nil {
		s.mu.Unlock()
		return err
	}

	settlement.AuditOpinion = opinion
	s.settlements[projectID] = settlement
	s.addLog(projectID, auditor, common.AuditActionAuditComment, "添加审计意见")
	s.mu.Unlock()

	return nil
}

func (s *Store) AuditPass(projectID, auditor, opinion string) (*common.Settlement, error) {
	plock := s.getProjectLock(projectID)
	plock.Lock()
	defer plock.Unlock()

	s.mu.Lock()
	if s.checkProjectLocked(projectID) {
		s.mu.Unlock()
		return nil, ErrProjectLocked
	}

	project, ok := s.projects[projectID]
	if !ok {
		s.mu.Unlock()
		return nil, ErrProjectNotFound
	}

	settlement, err := s.CalculateSettlement(projectID)
	if err != nil {
		s.mu.Unlock()
		return nil, err
	}

	nowTime := now()
	settlement.AuditOpinion = opinion
	settlement.AuditedAt = &nowTime
	settlement.Auditor = auditor
	s.settlements[projectID] = settlement

	project.Status = common.ProjectStatusLocked
	project.UpdatedAt = nowTime
	s.projects[projectID] = project

	s.addLog(projectID, auditor, common.AuditActionAuditPass, "审计通过，项目已锁定")
	s.mu.Unlock()

	return settlement, nil
}

func (s *Store) GetSettlement(projectID string) (*common.Settlement, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.projects[projectID]; !ok {
		return nil, ErrProjectNotFound
	}

	return s.CalculateSettlement(projectID)
}

func (s *Store) GetAuditLogs(projectID string) ([]*common.AuditLog, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.projects[projectID]; !ok {
		return nil, ErrProjectNotFound
	}

	logs := make([]*common.AuditLog, len(s.logs[projectID]))
	copy(logs, s.logs[projectID])
	return logs, nil
}
