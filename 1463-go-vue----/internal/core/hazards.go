package core

import (
	"errors"
	"time"

	"safetymanager/internal/api"
)

func (s *Store) GetHazard(id string) (*Hazard, error) {
	s.hazardMu.RLock()
	defer s.hazardMu.RUnlock()

	hazard, exists := s.hazards[id]
	if !exists {
		return nil, errors.New("hazard not found")
	}
	return hazard, nil
}

func (s *Store) GetAllHazards() []*Hazard {
	s.hazardMu.RLock()
	defer s.hazardMu.RUnlock()

	hazards := make([]*Hazard, 0, len(s.hazards))
	for _, hazard := range s.hazards {
		hazards = append(hazards, hazard)
	}
	return hazards
}

func (s *Store) SubmitRemediation(hazardID, note string) error {
	if hazardID == "" {
		return errors.New("hazard ID cannot be empty")
	}
	if note == "" {
		return errors.New("remediation note cannot be empty")
	}

	s.hazardMu.Lock()
	defer s.hazardMu.Unlock()

	hazard, exists := s.hazards[hazardID]
	if !exists {
		return errors.New("hazard not found")
	}

	if hazard.Status != api.HazardStatusInProgress {
		return errors.New("hazard not in remediation state")
	}

	now := time.Now()
	hazard.RemediationNote = &note
	hazard.RemediatedAt = &now
	hazard.Status = api.HazardStatusPendingReview

	s.appendAuditLog("remediation_submitted", "hazard", hazardID, "note_length="+string(rune(len(note))))

	return nil
}

func (s *Store) ReviewRemediation(hazardID string, approved bool) error {
	if hazardID == "" {
		return errors.New("hazard ID cannot be empty")
	}

	s.hazardMu.Lock()
	defer s.hazardMu.Unlock()

	hazard, exists := s.hazards[hazardID]
	if !exists {
		return errors.New("hazard not found")
	}

	if hazard.Status != api.HazardStatusPendingReview {
		return errors.New("hazard not pending review")
	}

	if approved {
		now := time.Now()
		hazard.ClosedAt = &now
		hazard.Status = api.HazardStatusClosed
		s.appendAuditLog("remediation_approved", "hazard", hazardID, "status=closed")
	} else {
		hazard.Status = api.HazardStatusInProgress
		hazard.RemediationNote = nil
		hazard.RemediatedAt = nil
		s.appendAuditLog("remediation_rejected", "hazard", hazardID, "status=reopened")
	}

	return nil
}

func (s *Store) RequestLevelChange(hazardID string, proposedLevel api.HazardLevel, reason string) (*LevelChangeRequest, error) {
	if hazardID == "" {
		return nil, errors.New("hazard ID cannot be empty")
	}
	if reason == "" {
		return nil, errors.New("reason cannot be empty")
	}
	if proposedLevel != api.HazardLevelGeneral && proposedLevel != api.HazardLevelMajor && proposedLevel != api.HazardLevelCritical {
		return nil, errors.New("invalid hazard level")
	}

	s.hazardMu.Lock()
	defer s.hazardMu.Unlock()

	hazard, exists := s.hazards[hazardID]
	if !exists {
		return nil, errors.New("hazard not found")
	}

	if hazard.LevelChangeReq != nil && hazard.LevelChangeReq.Status == api.LevelChangeStatusPending {
		return nil, errors.New("pending level change request exists")
	}

	req := &LevelChangeRequest{
		ID:            generateID(),
		HazardID:      hazardID,
		ProposedLevel: proposedLevel,
		Reason:        reason,
		Status:        api.LevelChangeStatusPending,
		CreatedAt:     time.Now(),
	}

	hazard.LevelChangeReq = req

	s.appendAuditLog("level_change_requested", "hazard", hazardID,
		"from="+string(hazard.Level)+",to="+string(proposedLevel))

	return req, nil
}

func (s *Store) ReviewLevelChange(requestID string, approved bool) error {
	if requestID == "" {
		return errors.New("request ID cannot be empty")
	}

	s.hazardMu.Lock()
	defer s.hazardMu.Unlock()

	var targetHazard *Hazard
	for _, hazard := range s.hazards {
		if hazard.LevelChangeReq != nil && hazard.LevelChangeReq.ID == requestID {
			targetHazard = hazard
			break
		}
	}

	if targetHazard == nil {
		return errors.New("level change request not found")
	}

	req := targetHazard.LevelChangeReq
	if req.Status != api.LevelChangeStatusPending {
		return errors.New("request not pending")
	}

	if approved {
		req.Status = api.LevelChangeStatusApproved
		oldLevel := targetHazard.Level
		targetHazard.Level = req.ProposedLevel
		targetHazard.DueAt = calculateDueDate(req.ProposedLevel)
		s.appendAuditLog("level_change_approved", "hazard", targetHazard.ID,
			"from="+string(oldLevel)+",to="+string(req.ProposedLevel))
	} else {
		req.Status = api.LevelChangeStatusRejected
		s.appendAuditLog("level_change_rejected", "hazard", targetHazard.ID, "status=rejected")
	}

	return nil
}

func (s *Store) CheckEscalations() []string {
	now := time.Now()
	escalated := make([]string, 0)

	s.hazardMu.Lock()
	defer s.hazardMu.Unlock()

	for _, hazard := range s.hazards {
		if hazard.Status == api.HazardStatusClosed || hazard.Escalated {
			continue
		}

		if now.After(hazard.DueAt) {
			hazard.Escalated = true
			escalated = append(escalated, hazard.ID)
			s.appendAuditLog("hazard_escalated", "hazard", hazard.ID, "status=overdue")
		}
	}

	return escalated
}

func buildHazardResponse(hazard *Hazard, zoneName string) api.HazardResponse {
	return api.HazardResponse{
		ID:              hazard.ID,
		TaskID:          hazard.TaskID,
		ZoneID:          hazard.ZoneID,
		ZoneName:        zoneName,
		ItemName:        hazard.ItemName,
		Description:     hazard.Description,
		Level:           hazard.Level,
		Status:          hazard.Status,
		ResponsibleDept: hazard.ResponsibleDept,
		DueAt:           hazard.DueAt.Format(time.RFC3339),
		RemediationNote: toStringPtrPtr(hazard.RemediationNote),
		RemediatedAt:    toTimePtrPtr(hazard.RemediatedAt),
		ClosedAt:        toTimePtrPtr(hazard.ClosedAt),
		CreatedAt:       hazard.CreatedAt.Format(time.RFC3339),
	}
}

func buildLevelChangeResponse(req *LevelChangeRequest) api.LevelChangeRequestResponse {
	return api.LevelChangeRequestResponse{
		ID:            req.ID,
		HazardID:      req.HazardID,
		ProposedLevel: req.ProposedLevel,
		Reason:        req.Reason,
		Status:        req.Status,
		CreatedAt:     req.CreatedAt.Format(time.RFC3339),
	}
}
