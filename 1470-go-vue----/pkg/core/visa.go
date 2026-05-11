package core

import (
	"errors"

	"settlement/pkg/common"
)

var (
	ErrVisaNotFound        = errors.New("visa not found")
	ErrVisaAlreadyConfirmed = errors.New("visa is already confirmed, cannot modify")
	ErrInvalidRole         = errors.New("invalid role, must be 'supervisor' or 'owner'")
	ErrVisaNotPending      = errors.New("visa is not in pending status")
)

func (s *Store) CreateVisa(projectID, reason, changeContent string, increaseQty, decreaseQty, unitPrice float64, creator string) (*common.Visa, error) {
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

	s.visaCounters[projectID]++
	visaNumber := formatVisaNumber(s.visaCounters[projectID])

	amount := roundToCents((increaseQty - decreaseQty) * unitPrice)

	visa := &common.Visa{
		ID:                generateID(),
		ProjectID:         projectID,
		VisaNumber:        visaNumber,
		Reason:            reason,
		ChangeContent:     changeContent,
		IncreaseQty:       increaseQty,
		DecreaseQty:       decreaseQty,
		UnitPrice:         roundToCents(unitPrice),
		Amount:            amount,
		Status:            common.VisaStatusPending,
		SupervisorConfirm: false,
		OwnerConfirm:      false,
		Creator:           creator,
		CreatedAt:         now(),
		UpdatedAt:         now(),
	}

	s.visas[projectID] = append(s.visas[projectID], visa)
	s.projects[projectID].UpdatedAt = now()
	s.addLog(projectID, creator, common.AuditActionVisaSubmit, "提交签证: "+visaNumber)
	s.mu.Unlock()

	return visa, nil
}

func (s *Store) UpdateVisa(projectID, visaID string, req *common.UpdateVisaRequest, operator string) (*common.Visa, error) {
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

	var target *common.Visa
	idx := -1
	for i, v := range s.visas[projectID] {
		if v.ID == visaID {
			target = v
			idx = i
			break
		}
	}

	if target == nil {
		s.mu.Unlock()
		return nil, ErrVisaNotFound
	}

	if target.Status == common.VisaStatusConfirmed {
		s.mu.Unlock()
		return nil, ErrVisaAlreadyConfirmed
	}

	if req.Reason != "" {
		target.Reason = req.Reason
	}
	if req.ChangeContent != "" {
		target.ChangeContent = req.ChangeContent
	}
	if req.IncreaseQty >= 0 {
		target.IncreaseQty = req.IncreaseQty
	}
	if req.DecreaseQty >= 0 {
		target.DecreaseQty = req.DecreaseQty
	}
	if req.UnitPrice > 0 {
		target.UnitPrice = roundToCents(req.UnitPrice)
	}

	target.Amount = roundToCents((target.IncreaseQty - target.DecreaseQty) * target.UnitPrice)
	target.UpdatedAt = now()

	if target.Status == common.VisaStatusRejected {
		target.Status = common.VisaStatusPending
		target.SupervisorConfirm = false
		target.OwnerConfirm = false
	}

	s.visas[projectID][idx] = target
	s.projects[projectID].UpdatedAt = now()
	s.addLog(projectID, operator, common.AuditActionUpdate, "更新签证: "+target.VisaNumber)
	s.mu.Unlock()

	return target, nil
}

func (s *Store) ConfirmVisa(projectID, visaID, operator, role string) (*common.Visa, error) {
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

	var target *common.Visa
	idx := -1
	for i, v := range s.visas[projectID] {
		if v.ID == visaID {
			target = v
			idx = i
			break
		}
	}

	if target == nil {
		s.mu.Unlock()
		return nil, ErrVisaNotFound
	}

	if target.Status != common.VisaStatusPending {
		s.mu.Unlock()
		return nil, ErrVisaNotPending
	}

	switch role {
	case "supervisor":
		target.SupervisorConfirm = true
		s.addLog(projectID, operator, common.AuditActionVisaConfirm, "监理单位确认签证: "+target.VisaNumber)
	case "owner":
		target.OwnerConfirm = true
		s.addLog(projectID, operator, common.AuditActionVisaConfirm, "建设单位确认签证: "+target.VisaNumber)
	default:
		s.mu.Unlock()
		return nil, ErrInvalidRole
	}

	if target.SupervisorConfirm && target.OwnerConfirm {
		target.Status = common.VisaStatusConfirmed
	}

	target.UpdatedAt = now()
	s.visas[projectID][idx] = target
	s.projects[projectID].UpdatedAt = now()
	s.mu.Unlock()

	return target, nil
}

func (s *Store) RejectVisa(projectID, visaID, operator, role, reason string) (*common.Visa, error) {
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

	var target *common.Visa
	idx := -1
	for i, v := range s.visas[projectID] {
		if v.ID == visaID {
			target = v
			idx = i
			break
		}
	}

	if target == nil {
		s.mu.Unlock()
		return nil, ErrVisaNotFound
	}

	if target.Status == common.VisaStatusConfirmed {
		s.mu.Unlock()
		return nil, ErrVisaAlreadyConfirmed
	}

	target.Status = common.VisaStatusRejected
	target.SupervisorConfirm = false
	target.OwnerConfirm = false
	target.UpdatedAt = now()

	s.visas[projectID][idx] = target
	s.projects[projectID].UpdatedAt = now()

	desc := "驳回签证: " + target.VisaNumber
	if reason != "" {
		desc += " - " + reason
	}
	s.addLog(projectID, operator, common.AuditActionVisaReject, desc)
	s.mu.Unlock()

	return target, nil
}

func (s *Store) ListVisa(projectID string) ([]*common.Visa, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.projects[projectID]; !ok {
		return nil, ErrProjectNotFound
	}

	return s.visas[projectID], nil
}

func (s *Store) GetVisa(projectID, visaID string) (*common.Visa, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.projects[projectID]; !ok {
		return nil, ErrProjectNotFound
	}

	for _, v := range s.visas[projectID] {
		if v.ID == visaID {
			return v, nil
		}
	}
	return nil, ErrVisaNotFound
}
