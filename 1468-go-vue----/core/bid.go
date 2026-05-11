package core

import (
	"sync"
	"time"

	"github.com/google/uuid"

	"tender-management/common"
)

type BidService struct {
	store         *Store
	projectService *ProjectService
	submitMu      sync.Mutex
}

func NewBidService(store *Store, projectService *ProjectService) *BidService {
	return &BidService{
		store:         store,
		projectService: projectService,
	}
}

func (s *BidService) SubmitBid(req common.SubmitBidRequest, now time.Time) (*common.Bid, error) {
	s.submitMu.Lock()
	defer s.submitMu.Unlock()

	project, err := s.projectService.GetProject(req.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.Status == common.ProjectOpened {
		return nil, ErrProjectAlreadyOpened
	}
	if project.Status != common.ProjectPublished {
		return nil, ErrProjectNotPublished
	}
	if now.After(project.BidDeadline) {
		return nil, ErrBidDeadlinePassed
	}
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if project.Method == common.InviteTender {
		invited := false
		for _, s := range project.InvitedSuppliers {
			if s == req.SupplierID {
				invited = true
				break
			}
		}
		if !invited {
			return nil, ErrSupplierNotInvited
		}
	}

	existing, _ := s.store.GetBidBySupplier(req.ProjectID, req.SupplierID)
	if existing != nil {
		return nil, ErrAlreadySubmitted
	}

	technicalScore := calculateTechnicalScore(req.TechnicalPlan)

	bid := &common.Bid{
		ID:            uuid.New().String(),
		ProjectID:     req.ProjectID,
		SupplierID:    req.SupplierID,
		Amount:        req.Amount,
		TechnicalPlan: req.TechnicalPlan,
		TechnicalScore: technicalScore,
		DurationDays:  req.DurationDays,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	s.store.SaveBid(bid)
	return bid, nil
}

func (s *BidService) UpdateBid(bidID string, req common.UpdateBidRequest, now time.Time) (*common.Bid, error) {
	s.submitMu.Lock()
	defer s.submitMu.Unlock()

	bid, ok := s.store.GetBid(bidID)
	if !ok {
		return nil, ErrBidNotFound
	}

	project, err := s.projectService.GetProject(bid.ProjectID)
	if err != nil {
		return nil, err
	}

	if project.Status == common.ProjectOpened {
		return nil, ErrProjectAlreadyOpened
	}
	if now.After(project.BidDeadline) {
		return nil, ErrBidDeadlinePassed
	}
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	bid.Amount = req.Amount
	bid.TechnicalPlan = req.TechnicalPlan
	bid.TechnicalScore = calculateTechnicalScore(req.TechnicalPlan)
	bid.DurationDays = req.DurationDays
	bid.UpdatedAt = now

	s.store.SaveBid(bid)
	return bid, nil
}

func (s *BidService) GetBid(bidID string) (*common.Bid, error) {
	bid, ok := s.store.GetBid(bidID)
	if !ok {
		return nil, ErrBidNotFound
	}
	return bid, nil
}

func (s *BidService) GetBidBySupplier(projectID, supplierID string) (*common.Bid, error) {
	bid, ok := s.store.GetBidBySupplier(projectID, supplierID)
	if !ok {
		return nil, ErrBidNotFound
	}
	return bid, nil
}

func (s *BidService) ListBidsByProject(projectID string) []common.Bid {
	return s.store.ListBidsByProject(projectID)
}

func calculateTechnicalScore(plan string) float64 {
	length := len(plan)
	if length == 0 {
		return 0
	}
	score := float64(length) / 10.0
	if score > 100 {
		score = 100
	}
	return score
}
