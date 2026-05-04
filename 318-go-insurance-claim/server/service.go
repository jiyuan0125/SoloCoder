package main

import (
	"fmt"
	"time"

	"insurance-claim/common"
)

type Service struct {
	storage *Storage
}

func NewService(storage *Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) CreatePolicy(req common.CreatePolicyRequest) (*common.Policy, error) {
	if req.PolicyNumber == "" {
		return nil, fmt.Errorf("policy number cannot be empty")
	}

	if req.TotalAmount <= 0 {
		return nil, fmt.Errorf("total amount must be positive")
	}

	if req.EffectiveDate.IsZero() {
		return nil, fmt.Errorf("effective date is required")
	}

	policy := &common.Policy{
		PolicyNumber:    req.PolicyNumber,
		PolicyType:      req.PolicyType,
		TotalAmount:     req.TotalAmount,
		RemainingAmount: req.TotalAmount,
		EffectiveDate:   req.EffectiveDate,
		Status:          common.PolicyStatusActive,
		CreatedAt:       time.Now(),
	}

	if err := s.storage.CreatePolicy(policy); err != nil {
		return nil, err
	}

	return policy, nil
}

func (s *Service) SubmitClaim(req common.SubmitClaimRequest) (*common.Claim, error) {
	if req.PolicyNumber == "" {
		return nil, fmt.Errorf("policy number cannot be empty")
	}

	if req.IncidentDate.IsZero() {
		return nil, fmt.Errorf("incident date is required")
	}

	if req.IncidentReason == "" {
		return nil, fmt.Errorf("incident reason cannot be empty")
	}

	if len(req.IncidentReason) > common.MaxReasonLength {
		return nil, fmt.Errorf("incident reason cannot exceed %d characters", common.MaxReasonLength)
	}

	if req.RequestedAmount <= 0 {
		return nil, fmt.Errorf("requested amount must be positive")
	}

	policy, err := s.storage.GetPolicy(req.PolicyNumber)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if req.IncidentDate.After(now) {
		return nil, fmt.Errorf("incident date cannot be in the future")
	}

	if req.IncidentDate.Before(policy.EffectiveDate) {
		return nil, fmt.Errorf("incident date cannot be before policy effective date")
	}

	if policy.Status == common.PolicyStatusExhausted {
		return nil, fmt.Errorf("policy is exhausted")
	}

	if s.storage.HasActiveClaim(req.PolicyNumber) {
		return nil, fmt.Errorf("policy has an active claim, please wait for it to be processed")
	}

	if req.RequestedAmount > policy.RemainingAmount {
		return nil, fmt.Errorf("requested amount exceeds remaining policy amount")
	}

	if policy.PolicyType == common.PolicyTypeAuto && req.LiabilityRatio == "" {
		return nil, fmt.Errorf("liability ratio is required for auto insurance")
	}

	payoutPercentage := s.calculatePayoutPercentage(policy.PolicyType, req.LiabilityRatio)
	payoutAmount := req.RequestedAmount * payoutPercentage

	if payoutAmount > policy.RemainingAmount {
		payoutAmount = policy.RemainingAmount
	}

	claim := &common.Claim{
		ClaimID:          generateID(),
		PolicyNumber:     req.PolicyNumber,
		IncidentDate:     req.IncidentDate,
		IncidentReason:   req.IncidentReason,
		RequestedAmount:  req.RequestedAmount,
		LiabilityRatio:   req.LiabilityRatio,
		PayoutPercentage: payoutPercentage,
		PayoutAmount:     payoutAmount,
		Status:           common.ClaimStatusPending,
		CreatedAt:        time.Now(),
	}

	if err := s.storage.CreateClaim(claim); err != nil {
		return nil, err
	}

	return claim, nil
}

func (s *Service) calculatePayoutPercentage(policyType common.PolicyType, liabilityRatio common.LiabilityRatio) float64 {
	switch policyType {
	case common.PolicyTypeMedical:
		return 0.8
	case common.PolicyTypeAccident:
		return 1.0
	case common.PolicyTypeAuto:
		return liabilityRatio.ToPercentage()
	default:
		return 0.0
	}
}

func (s *Service) ReviewClaim(req common.ReviewClaimRequest) (*common.Claim, error) {
	if req.ClaimID == "" {
		return nil, fmt.Errorf("claim ID cannot be empty")
	}

	claim, err := s.storage.GetClaim(req.ClaimID)
	if err != nil {
		return nil, err
	}

	if claim.Status != common.ClaimStatusPending {
		return nil, fmt.Errorf("only pending claims can be reviewed")
	}

	if !req.Approved && req.RejectReason == "" {
		return nil, fmt.Errorf("reject reason is required when rejecting a claim")
	}

	now := time.Now()
	claim.ReviewedAt = &now

	if req.Approved {
		claim.Status = common.ClaimStatusApproved
	} else {
		claim.Status = common.ClaimStatusRejected
		claim.RejectReason = req.RejectReason
	}

	if err := s.storage.UpdateClaim(claim); err != nil {
		return nil, err
	}

	return claim, nil
}

func (s *Service) ConfirmPayment(req common.ConfirmPaymentRequest) (*common.Claim, error) {
	if req.ClaimID == "" {
		return nil, fmt.Errorf("claim ID cannot be empty")
	}

	claim, err := s.storage.GetClaim(req.ClaimID)
	if err != nil {
		return nil, err
	}

	if claim.Status != common.ClaimStatusApproved {
		return nil, fmt.Errorf("only approved claims can be paid")
	}

	policy, err := s.storage.GetPolicy(claim.PolicyNumber)
	if err != nil {
		return nil, err
	}

	if claim.PayoutAmount > policy.RemainingAmount {
		return nil, fmt.Errorf("payout amount exceeds remaining policy amount")
	}

	now := time.Now()
	claim.PaidAt = &now
	claim.Status = common.ClaimStatusPaid

	if err := s.storage.UpdateClaim(claim); err != nil {
		return nil, err
	}

	policy.RemainingAmount -= claim.PayoutAmount
	if policy.RemainingAmount <= 0 {
		policy.Status = common.PolicyStatusExhausted
	}

	if err := s.storage.UpdatePolicy(policy); err != nil {
		return nil, err
	}

	return claim, nil
}

func (s *Service) GetClaim(claimID string) (*common.Claim, error) {
	if claimID == "" {
		return nil, fmt.Errorf("claim ID cannot be empty")
	}
	return s.storage.GetClaim(claimID)
}

func (s *Service) GetPolicyClaims(policyNumber string) ([]*common.Claim, error) {
	if policyNumber == "" {
		return nil, fmt.Errorf("policy number cannot be empty")
	}
	return s.storage.GetPolicyClaims(policyNumber)
}

func (s *Service) GetPendingClaims() []*common.Claim {
	return s.storage.GetPendingClaims()
}

func (s *Service) GetApprovedClaims() []*common.Claim {
	return s.storage.GetApprovedClaims()
}

func (s *Service) ListAllPolicies() []*common.Policy {
	return s.storage.ListAllPolicies()
}

func (s *Service) ListAllClaims() []*common.Claim {
	return s.storage.ListAllClaims()
}
