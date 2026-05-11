package core

import (
	"sort"
	"time"

	"tender-management/common"
)

var ErrOpeningResultNotFound = ErrBidNotFound

type OpeningService struct {
	store          *Store
	projectService *ProjectService
	bidService     *BidService
}

func NewOpeningService(store *Store, projectService *ProjectService, bidService *BidService) *OpeningService {
	return &OpeningService{
		store:          store,
		projectService: projectService,
		bidService:     bidService,
	}
}

func (s *OpeningService) OpenProject(projectID string, now time.Time) (*common.OpeningResult, error) {
	project, err := s.projectService.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	if project.Status == common.ProjectOpened {
		return nil, ErrProjectAlreadyOpened
	}
	if project.Status != common.ProjectPublished {
		return nil, ErrProjectNotPublished
	}
	if now.Before(project.OpenTime) {
		return nil, ErrOpenTimeNotReached
	}

	bids := s.bidService.ListBidsByProject(projectID)
	if len(bids) == 0 {
		return nil, ErrNoBids
	}

	bidPtrs := make([]*common.Bid, len(bids))
	for i := range bids {
		bidPtrs[i] = &bids[i]
	}

	calculateScores(bidPtrs, project.EvaluationMethod)

	sortBids(bidPtrs, project.EvaluationMethod)

	winner := bidPtrs[0]

	result := &common.OpeningResult{
		ProjectID:        projectID,
		WinnerBidID:      winner.ID,
		WinnerSupplierID: winner.SupplierID,
		WinnerAmount:     winner.Amount,
		OpenedAt:         now,
	}

	project.Status = common.ProjectOpened
	project.UpdatedAt = now
	s.store.SaveProject(project)
	s.store.SaveOpeningResult(result)

	for _, b := range bidPtrs {
		s.store.SaveBid(b)
	}

	return result, nil
}

func (s *OpeningService) GetOpeningResult(projectID string) (*common.OpeningResult, error) {
	result, ok := s.store.GetOpeningResult(projectID)
	if !ok {
		return nil, ErrOpeningResultNotFound
	}
	return result, nil
}

func calculateScores(bids []*common.Bid, method common.EvaluationMethod) {
	if method != common.ComprehensiveScoreMethod {
		return
	}

	var lowestAmount int64 = 0
	for _, b := range bids {
		if lowestAmount == 0 || b.Amount < lowestAmount {
			lowestAmount = b.Amount
		}
	}

	const (
		priceWeight     = 0.60
		technicalWeight = 0.40
	)

	for _, b := range bids {
		priceScore := (float64(lowestAmount) / float64(b.Amount)) * 100 * priceWeight
		techScore := b.TechnicalScore * technicalWeight
		b.ComprehensiveScore = priceScore + techScore
	}
}

func sortBids(bids []*common.Bid, method common.EvaluationMethod) {
	sort.Slice(bids, func(i, j int) bool {
		if method == common.LowestPriceMethod {
			if bids[i].Amount != bids[j].Amount {
				return bids[i].Amount < bids[j].Amount
			}
		} else {
			if bids[i].ComprehensiveScore != bids[j].ComprehensiveScore {
				return bids[i].ComprehensiveScore > bids[j].ComprehensiveScore
			}
		}
		return bids[i].CreatedAt.Before(bids[j].CreatedAt)
	})
}
