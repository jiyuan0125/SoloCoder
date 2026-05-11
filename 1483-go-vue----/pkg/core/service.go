package core

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"insurance-claim/pkg/common"
)

type ClaimService interface {
	SubmitClaim(req *common.SubmitClaimRequest) (*common.ClaimCase, error)
	GetCase(caseID string) (*common.ClaimCase, bool)
	GetCaseByNo(caseNo string) (*common.ClaimCase, bool)
	ListCases() []*common.ClaimCase
	AssignInvestigator(req *common.AssignInvestigatorRequest) error
	SubmitInvestigation(req *common.SubmitInvestigationRequest) error
	AssignAssessor(req *common.AssignAssessorRequest) error
	SubmitAssessment(req *common.SubmitAssessmentRequest) error
	CalculatePayout(caseID string) (*common.Payout, error)
	ApprovePayout(caseID string, operator string) error
	MarkAsPaid(caseID string, operator string) error
	FlagForReview(req *common.FlagForReviewRequest) error
	ResolveReview(req *common.ResolveReviewRequest) error
	CreatePolicy(policy *common.Policy) error
	GetPolicy(policyNo string) (*common.Policy, bool)
}

type claimService struct {
	store      Store
	calculator *PayoutCalculator
	caseLocks  map[string]*sync.Mutex
	locksMu    sync.Mutex
}

func NewClaimService(store Store) ClaimService {
	return &claimService{
		store:      store,
		calculator: NewPayoutCalculator(),
		caseLocks:  make(map[string]*sync.Mutex),
	}
}

func (s *claimService) getCaseLock(caseID string) *sync.Mutex {
	s.locksMu.Lock()
	defer s.locksMu.Unlock()
	if lock, ok := s.caseLocks[caseID]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	s.caseLocks[caseID] = lock
	return lock
}

func (s *claimService) addStatusHistory(caseID string, fromStatus, toStatus common.CaseStatus, operator, reason string) error {
	history := &common.StatusHistory{
		ID:         uuid.New().String(),
		CaseID:     caseID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		Operator:   operator,
		Reason:     reason,
		CreatedAt:  time.Now(),
	}
	return s.store.AddStatusHistory(history)
}

func (s *claimService) SubmitClaim(req *common.SubmitClaimRequest) (*common.ClaimCase, error) {
	if req.PolicyNo == "" {
		return nil, errors.New("policy no is required")
	}

	policy, ok := s.store.GetPolicy(req.PolicyNo)
	if !ok {
		return nil, errors.New("policy not found")
	}

	if req.EstimatedAmount <= 0 {
		return nil, errors.New("estimated amount must be positive")
	}

	now := time.Now()
	caseNo, err := s.store.GenerateCaseNo(now)
	if err != nil {
		return nil, err
	}

	caseItem := &common.ClaimCase{
		ID:                  uuid.New().String(),
		CaseNo:              caseNo,
		PolicyNo:            req.PolicyNo,
		PolicyLimit:         policy.Limit,
		AccidentTime:        req.AccidentTime,
		ReportTime:          now,
		AccidentLocation:    req.AccidentLocation,
		AccidentType:        req.AccidentType,
		AccidentDescription: req.AccidentDescription,
		EstimatedAmount:     req.EstimatedAmount,
		Status:              common.CaseStatusSubmitted,
		IsUnderReview:       false,
		StatusHistory:       []common.StatusHistory{},
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	err = s.store.CreateCase(caseItem)
	if err != nil {
		return nil, err
	}

	s.addStatusHistory(caseItem.ID, "", common.CaseStatusSubmitted, "system", "案件提交")

	return caseItem, nil
}

func (s *claimService) GetCase(caseID string) (*common.ClaimCase, bool) {
	return s.store.GetCase(caseID)
}

func (s *claimService) GetCaseByNo(caseNo string) (*common.ClaimCase, bool) {
	return s.store.GetCaseByNo(caseNo)
}

func (s *claimService) ListCases() []*common.ClaimCase {
	return s.store.ListCases()
}

func (s *claimService) AssignInvestigator(req *common.AssignInvestigatorRequest) error {
	lock := s.getCaseLock(req.CaseID)
	lock.Lock()
	defer lock.Unlock()

	c, ok := s.store.GetCase(req.CaseID)
	if !ok {
		return errors.New("case not found")
	}

	if c.Status != common.CaseStatusSubmitted {
		return errors.New("case is not in submitted status")
	}

	if c.IsUnderReview {
		return errors.New("case is under review")
	}

	oldStatus := c.Status
	c.InvestigatorID = req.InvestigatorID
	c.InvestigatorName = req.InvestigatorName
	c.Status = common.CaseStatusAssigned
	c.UpdatedAt = time.Now()

	err := s.store.UpdateCase(c)
	if err != nil {
		return err
	}

	s.addStatusHistory(c.ID, oldStatus, common.CaseStatusAssigned, req.Operator, "分配查勘员")
	return nil
}

func (s *claimService) SubmitInvestigation(req *common.SubmitInvestigationRequest) error {
	lock := s.getCaseLock(req.CaseID)
	lock.Lock()
	defer lock.Unlock()

	c, ok := s.store.GetCase(req.CaseID)
	if !ok {
		return errors.New("case not found")
	}

	if c.Status != common.CaseStatusAssigned && c.Status != common.CaseStatusInvestigating {
		return errors.New("case is not in assignable status for investigation")
	}

	if c.IsUnderReview {
		return errors.New("case is under review")
	}

	if c.InvestigatorID != req.InvestigatorID {
		return errors.New("investigator mismatch")
	}

	report := &common.InvestigationReport{
		CaseID:            req.CaseID,
		InvestigatorID:    req.InvestigatorID,
		InvestigatorName:  req.InvestigatorName,
		CauseAnalysis:     req.CauseAnalysis,
		Liability:         req.Liability,
		DamagedParts:      req.DamagedParts,
		HasValidReason:    req.HasValidReason,
		ReasonDescription: req.ReasonDescription,
		InvestigatedAt:    time.Now(),
	}

	oldStatus := c.Status
	c.InvestigationReport = report
	c.Status = common.CaseStatusInvestigated
	c.UpdatedAt = time.Now()

	err := s.store.UpdateCase(c)
	if err != nil {
		return err
	}

	s.addStatusHistory(c.ID, oldStatus, common.CaseStatusInvestigated, req.InvestigatorID, "提交查勘报告")
	return nil
}

func (s *claimService) AssignAssessor(req *common.AssignAssessorRequest) error {
	lock := s.getCaseLock(req.CaseID)
	lock.Lock()
	defer lock.Unlock()

	c, ok := s.store.GetCase(req.CaseID)
	if !ok {
		return errors.New("case not found")
	}

	if c.Status != common.CaseStatusInvestigated {
		return errors.New("case is not in investigated status")
	}

	if c.IsUnderReview {
		return errors.New("case is under review")
	}

	oldStatus := c.Status
	c.AssessorID = req.AssessorID
	c.AssessorName = req.AssessorName
	c.Status = common.CaseStatusAssessing
	c.UpdatedAt = time.Now()

	err := s.store.UpdateCase(c)
	if err != nil {
		return err
	}

	s.addStatusHistory(c.ID, oldStatus, common.CaseStatusAssessing, req.Operator, "分配定损员")
	return nil
}

func (s *claimService) SubmitAssessment(req *common.SubmitAssessmentRequest) error {
	lock := s.getCaseLock(req.CaseID)
	lock.Lock()
	defer lock.Unlock()

	c, ok := s.store.GetCase(req.CaseID)
	if !ok {
		return errors.New("case not found")
	}

	if c.Status != common.CaseStatusAssessing {
		return errors.New("case is not in assessing status")
	}

	if c.IsUnderReview {
		return errors.New("case is under review")
	}

	if c.AssessorID != req.AssessorID {
		return errors.New("assessor mismatch")
	}

	if len(req.Items) == 0 {
		return errors.New("at least one damage item is required")
	}

	items := make([]common.DamageItem, 0, len(req.Items))
	totalAmount := int64(0)
	for _, item := range req.Items {
		if item.Amount <= 0 {
			return errors.New("item amount must be positive")
		}
		di := common.DamageItem{
			ID:         uuid.New().String(),
			CaseID:     req.CaseID,
			ItemName:   item.ItemName,
			RepairType: item.RepairType,
			Amount:     item.Amount,
			CreatedAt:  time.Now(),
		}
		items = append(items, di)
		totalAmount += item.Amount
	}

	if totalAmount > c.PolicyLimit {
		totalAmount = c.PolicyLimit
	}

	assessment := &common.Assessment{
		CaseID:       req.CaseID,
		AssessorID:   req.AssessorID,
		AssessorName: req.AssessorName,
		TotalAmount:  totalAmount,
		Items:        items,
		AssessedAt:   time.Now(),
	}

	oldStatus := c.Status
	c.Assessment = assessment
	c.Status = common.CaseStatusAssessed
	c.UpdatedAt = time.Now()

	err := s.store.UpdateCase(c)
	if err != nil {
		return err
	}

	s.addStatusHistory(c.ID, oldStatus, common.CaseStatusAssessed, req.AssessorID, "提交定损报告")
	return nil
}

func (s *claimService) CalculatePayout(caseID string) (*common.Payout, error) {
	lock := s.getCaseLock(caseID)
	lock.Lock()
	defer lock.Unlock()

	c, ok := s.store.GetCase(caseID)
	if !ok {
		return nil, errors.New("case not found")
	}

	if c.Status != common.CaseStatusAssessed {
		return nil, errors.New("case is not in assessed status")
	}

	if c.IsUnderReview {
		return nil, errors.New("case is under review")
	}

	if c.InvestigationReport == nil || c.Assessment == nil {
		return nil, errors.New("missing investigation or assessment")
	}

	claimCount := s.store.CountClaimsThisYear(c.PolicyNo, c.CreatedAt)

	payout := s.calculator.CalculatePayout(
		c.Assessment.TotalAmount,
		c.PolicyLimit,
		c.InvestigationReport.Liability,
		claimCount,
		c.AccidentTime,
		c.ReportTime,
		c.InvestigationReport.HasValidReason,
	)
	payout.CaseID = caseID

	c.Payout = payout
	c.UpdatedAt = time.Now()

	err := s.store.UpdateCase(c)
	if err != nil {
		return nil, err
	}

	return payout, nil
}

func (s *claimService) ApprovePayout(caseID string, operator string) error {
	lock := s.getCaseLock(caseID)
	lock.Lock()
	defer lock.Unlock()

	c, ok := s.store.GetCase(caseID)
	if !ok {
		return errors.New("case not found")
	}

	if c.Status != common.CaseStatusAssessed {
		return errors.New("case is not in assessed status")
	}

	if c.IsUnderReview {
		return errors.New("case is under review")
	}

	if c.Payout == nil {
		return errors.New("payout not calculated")
	}

	oldStatus := c.Status
	c.Status = common.CaseStatusApproved
	c.UpdatedAt = time.Now()

	err := s.store.UpdateCase(c)
	if err != nil {
		return err
	}

	s.addStatusHistory(c.ID, oldStatus, common.CaseStatusApproved, operator, "审批赔付")
	return nil
}

func (s *claimService) MarkAsPaid(caseID string, operator string) error {
	lock := s.getCaseLock(caseID)
	lock.Lock()
	defer lock.Unlock()

	c, ok := s.store.GetCase(caseID)
	if !ok {
		return errors.New("case not found")
	}

	if c.Status != common.CaseStatusApproved {
		return errors.New("case is not in approved status")
	}

	if c.IsUnderReview {
		return errors.New("case is under review")
	}

	oldStatus := c.Status
	c.Status = common.CaseStatusPaid
	c.UpdatedAt = time.Now()

	err := s.store.UpdateCase(c)
	if err != nil {
		return err
	}

	s.addStatusHistory(c.ID, oldStatus, common.CaseStatusPaid, operator, "标记已赔付")
	return nil
}

func (s *claimService) FlagForReview(req *common.FlagForReviewRequest) error {
	lock := s.getCaseLock(req.CaseID)
	lock.Lock()
	defer lock.Unlock()

	c, ok := s.store.GetCase(req.CaseID)
	if !ok {
		return errors.New("case not found")
	}

	if c.IsUnderReview {
		return errors.New("case is already under review")
	}

	oldStatus := c.Status
	c.IsUnderReview = true
	c.Status = common.CaseStatusReviewing
	c.UpdatedAt = time.Now()

	err := s.store.UpdateCase(c)
	if err != nil {
		return err
	}

	s.addStatusHistory(c.ID, oldStatus, common.CaseStatusReviewing, req.Operator, req.Reason)
	return nil
}

func (s *claimService) ResolveReview(req *common.ResolveReviewRequest) error {
	lock := s.getCaseLock(req.CaseID)
	lock.Lock()
	defer lock.Unlock()

	c, ok := s.store.GetCase(req.CaseID)
	if !ok {
		return errors.New("case not found")
	}

	if !c.IsUnderReview {
		return errors.New("case is not under review")
	}

	oldStatus := c.Status
	c.IsUnderReview = false

	if req.ShouldReject {
		c.Status = common.CaseStatusRejected
	} else {
		switch oldStatus {
		case common.CaseStatusAssigned, common.CaseStatusInvestigating:
			c.Status = common.CaseStatusAssigned
		case common.CaseStatusInvestigated, common.CaseStatusAssessing:
			c.Status = common.CaseStatusInvestigated
		case common.CaseStatusAssessed:
			c.Status = common.CaseStatusAssessed
		case common.CaseStatusApproved:
			c.Status = common.CaseStatusApproved
		default:
			c.Status = common.CaseStatusSubmitted
		}
	}

	c.UpdatedAt = time.Now()

	err := s.store.UpdateCase(c)
	if err != nil {
		return err
	}

	s.addStatusHistory(c.ID, oldStatus, c.Status, req.Operator, req.Reason)
	return nil
}

func (s *claimService) CreatePolicy(policy *common.Policy) error {
	return s.store.CreatePolicy(policy)
}

func (s *claimService) GetPolicy(policyNo string) (*common.Policy, bool) {
	return s.store.GetPolicy(policyNo)
}
