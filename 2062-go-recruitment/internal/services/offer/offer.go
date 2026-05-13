package offer

import (
	"fmt"
	"time"

	"recruitment/internal/models"
	"recruitment/internal/services/candidate"
	"recruitment/internal/services/interview"
	"recruitment/internal/store"
	"recruitment/pkg/utils"
)

type Service struct {
	store        *store.Store
	candidateSvc *candidate.Service
	interviewSvc *interview.Service
}

func NewService(s *store.Store, cs *candidate.Service, is *interview.Service) *Service {
	return &Service{store: s, candidateSvc: cs, interviewSvc: is}
}

type CreateOfferRequest struct {
	CandidateID string
	Salary      float64
	StartDate   time.Time
	ValidDays   int
}

func (s *Service) Create(req CreateOfferRequest) (*models.Offer, error) {
	if req.CandidateID == "" {
		return nil, fmt.Errorf("候选人ID不能为空")
	}
	if req.Salary <= 0 {
		return nil, fmt.Errorf("薪资必须大于0")
	}
	if req.StartDate.IsZero() {
		return nil, fmt.Errorf("入职日期不能为空")
	}
	if req.ValidDays <= 0 {
		return nil, fmt.Errorf("有效期天数必须大于0")
	}

	cand, err := s.candidateSvc.GetByID(req.CandidateID)
	if err != nil {
		return nil, err
	}

	if cand.Status == models.CandidateStatusRejected ||
		cand.Status == models.CandidateStatusWithdrawn ||
		cand.Status == models.CandidateStatusOfferAbandoned {
		return nil, fmt.Errorf("候选人已淘汰，无法生成 Offer")
	}

	allPassed, err := s.interviewSvc.AllInterviewsPassed(req.CandidateID)
	if err != nil {
		return nil, err
	}
	if !allPassed {
		return nil, fmt.Errorf("候选人尚未通过所有面试")
	}

	existingOffers, err := s.GetByCandidate(req.CandidateID)
	if err != nil {
		return nil, err
	}
	for _, o := range existingOffers {
		if o.Status == models.OfferStatusPending || o.Status == models.OfferStatusAccepted {
			return nil, fmt.Errorf("候选人已有生效中的 Offer")
		}
	}

	offer := &models.Offer{
		ID:          utils.GenerateID(),
		CandidateID: req.CandidateID,
		JobID:       cand.JobID,
		Salary:      req.Salary,
		StartDate:   req.StartDate,
		ValidUntil:  time.Now().AddDate(0, 0, req.ValidDays),
		Status:      models.OfferStatusPending,
		CreatedAt:   time.Now(),
	}

	err = s.store.Transact(func(db *models.Database) error {
		db.Offers = append(db.Offers, *offer)
		for i := range db.Candidates {
			if db.Candidates[i].ID == req.CandidateID {
				db.Candidates[i].Status = models.CandidateStatusOffer
				break
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if err := s.store.TriggerUpdateCheck(cand.JobID); err != nil {
		fmt.Printf("Warning: trigger update check failed: %v\n", err)
	}

	return offer, nil
}

func (s *Service) Accept(id string) error {
	err := s.store.Transact(func(db *models.Database) error {
		for i := range db.Offers {
			if db.Offers[i].ID == id {
				offer := &db.Offers[i]

				if offer.Status != models.OfferStatusPending {
					return fmt.Errorf("Offer 状态不支持接受操作: %s", offer.Status)
				}

				offer.Status = models.OfferStatusAccepted
				repliedAt := time.Now()
				offer.RepliedAt = &repliedAt

				for j := range db.Candidates {
					if db.Candidates[j].ID == offer.CandidateID {
						db.Candidates[j].Status = models.CandidateStatusOnboard
						break
					}
				}

				return nil
			}
		}
		return fmt.Errorf("Offer 不存在: %s", id)
	})

	if err != nil {
		return err
	}

	offer, err := s.GetByID(id)
	if err != nil {
		return nil
	}
	return s.store.TriggerUpdateCheck(offer.JobID)
}

func (s *Service) Reject(id string) error {
	err := s.store.Transact(func(db *models.Database) error {
		for i := range db.Offers {
			if db.Offers[i].ID == id {
				offer := &db.Offers[i]

				if offer.Status != models.OfferStatusPending {
					return fmt.Errorf("Offer 状态不支持拒绝操作: %s", offer.Status)
				}

				offer.Status = models.OfferStatusRejected
				repliedAt := time.Now()
				offer.RepliedAt = &repliedAt

				for j := range db.Candidates {
					if db.Candidates[j].ID == offer.CandidateID {
						db.Candidates[j].Status = models.CandidateStatusRejected
						db.Candidates[j].RejectionReason = "候选人拒绝 Offer"
						rejectedAt := time.Now()
						db.Candidates[j].RejectedAt = &rejectedAt
						break
					}
				}

				return nil
			}
		}
		return fmt.Errorf("Offer 不存在: %s", id)
	})

	if err != nil {
		return err
	}

	offer, err := s.GetByID(id)
	if err != nil {
		return nil
	}
	return s.store.TriggerUpdateCheck(offer.JobID)
}

type AdjustBudgetRequest struct {
	JobID       string
	NewTotal    float64
}

func (s *Service) AdjustBudget(req AdjustBudgetRequest) error {
	if req.JobID == "" {
		return fmt.Errorf("职位ID不能为空")
	}
	if req.NewTotal < 0 {
		return fmt.Errorf("总金额不能为负数")
	}

	return s.store.Transact(func(db *models.Database) error {
		var budget *models.Budget
		for i := range db.Budgets {
			if db.Budgets[i].JobID == req.JobID {
				budget = &db.Budgets[i]
				break
			}
		}

		if budget == nil {
			return fmt.Errorf("职位预算不存在: %s", req.JobID)
		}

		if budget.TotalAmount == 0 {
			return fmt.Errorf("原总金额为0，无法按比例分配")
		}

		ratio := req.NewTotal / budget.TotalAmount

		for i := range budget.Stages {
			budget.Stages[i].PlannedAmount = budget.Stages[i].PlannedAmount * ratio
		}

		budget.TotalAmount = req.NewTotal
		budget.UpdatedAt = time.Now()

		return nil
	})
}

type InitBudgetRequest struct {
	JobID       string
	TotalAmount float64
	Stages      []string
}

func (s *Service) InitBudget(req InitBudgetRequest) error {
	if req.JobID == "" {
		return fmt.Errorf("职位ID不能为空")
	}
	if req.TotalAmount < 0 {
		return fmt.Errorf("总金额不能为负数")
	}
	if len(req.Stages) == 0 {
		return fmt.Errorf("阶段列表不能为空")
	}

	return s.store.Transact(func(db *models.Database) error {
		for i := range db.Budgets {
			if db.Budgets[i].JobID == req.JobID {
				return fmt.Errorf("职位预算已存在: %s", req.JobID)
			}
		}

		stageAmount := req.TotalAmount / float64(len(req.Stages))
		stages := make([]models.BudgetStage, len(req.Stages))
		for i, name := range req.Stages {
			stages[i] = models.BudgetStage{
				Stage:         name,
				PlannedAmount: stageAmount,
			}
		}

		db.Budgets = append(db.Budgets, models.Budget{
			JobID:       req.JobID,
			TotalAmount: req.TotalAmount,
			Stages:      stages,
			UpdatedAt:   time.Now(),
		})

		return nil
	})
}

func (s *Service) GetByID(id string) (*models.Offer, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	for i := range db.Offers {
		if db.Offers[i].ID == id {
			return &db.Offers[i], nil
		}
	}

	return nil, fmt.Errorf("Offer 不存在: %s", id)
}

func (s *Service) GetByCandidate(candidateID string) ([]models.Offer, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	var result []models.Offer
	for _, o := range db.Offers {
		if o.CandidateID == candidateID {
			result = append(result, o)
		}
	}
	return result, nil
}

func (s *Service) List() ([]models.Offer, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	return db.Offers, nil
}

func (s *Service) GetBudget(jobID string) (*models.Budget, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	for i := range db.Budgets {
		if db.Budgets[i].JobID == jobID {
			return &db.Budgets[i], nil
		}
	}

	return nil, fmt.Errorf("职位预算不存在: %s", jobID)
}

func (s *Service) ListBudgets() ([]models.Budget, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	return db.Budgets, nil
}
