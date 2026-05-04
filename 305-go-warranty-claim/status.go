package main

import (
	"time"
)

type StatusService struct {
	storage *Storage
}

func NewStatusService(storage *Storage) *StatusService {
	return &StatusService{storage: storage}
}

func (s *StatusService) ReviewClaim(claimID string, approved bool, reason string) (*WarrantyClaim, error) {
	claim, exists := s.storage.GetClaimByID(claimID)
	if !exists {
		return nil, &ValidationError{
			Field:   "claim_id",
			Message: "申请不存在",
		}
	}

	if claim.Status != StatusPending {
		return nil, &ValidationError{
			Field:   "status",
			Message: "只有待处理的申请才能审核",
		}
	}

	if approved {
		claim.Status = StatusApproved
		claim.ApprovedAt = time.Now()
		claim.WarrantyStartAt = claim.ApprovedAt
		claim.RejectReason = ""
	} else {
		claim.Status = StatusRejected
		claim.RejectReason = reason
	}

	if err := s.storage.UpdateClaim(claim); err != nil {
		return nil, err
	}

	return claim, nil
}

func (s *StatusService) SubmitAppeal(claimID string, reason string) (*Appeal, error) {
	claim, exists := s.storage.GetClaimByID(claimID)
	if !exists {
		return nil, &ValidationError{
			Field:   "claim_id",
			Message: "申请不存在",
		}
	}

	if claim.Status != StatusApproved && claim.Status != StatusRejected {
		return nil, &ValidationError{
			Field:   "status",
			Message: "只能对已审核的申请发起申诉",
		}
	}

	if len(reason) == 0 {
		return nil, &ValidationError{
			Field:   "reason",
			Message: "申诉原因不能为空",
		}
	}

	appeal := &Appeal{
		ID:        GenerateID(),
		ClaimID:   claimID,
		Reason:    reason,
		CreatedAt: time.Now(),
		Resolved:  false,
	}

	if claim.Appeals == nil {
		claim.Appeals = []Appeal{}
	}
	claim.Appeals = append(claim.Appeals, *appeal)

	if err := s.storage.UpdateClaim(claim); err != nil {
		return nil, err
	}

	return appeal, nil
}

func (s *StatusService) ResolveAppeal(claimID string, appealID string, resolution string) error {
	claim, exists := s.storage.GetClaimByID(claimID)
	if !exists {
		return &ValidationError{
			Field:   "claim_id",
			Message: "申请不存在",
		}
	}

	found := false
	for i := range claim.Appeals {
		if claim.Appeals[i].ID == appealID {
			if claim.Appeals[i].Resolved {
				return &ValidationError{
					Field:   "appeal_id",
					Message: "申诉已处理",
				}
			}
			claim.Appeals[i].Resolved = true
			claim.Appeals[i].ResolvedAt = time.Now()
			claim.Appeals[i].Resolution = resolution
			found = true
			break
		}
	}

	if !found {
		return &ValidationError{
			Field:   "appeal_id",
			Message: "申诉不存在",
		}
	}

	return s.storage.UpdateClaim(claim)
}
