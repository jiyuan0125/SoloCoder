package candidate

import (
	"fmt"
	"time"

	"recruitment/internal/models"
	"recruitment/internal/services/job"
	"recruitment/internal/store"
	"recruitment/pkg/utils"
)

type Service struct {
	store     *store.Store
	jobService *job.Service
}

func NewService(s *store.Store, js *job.Service) *Service {
	return &Service{store: s, jobService: js}
}

type ApplyRequest struct {
	ResumeID string
	Name     string
	Email    string
	Phone    string
	JobID    string
}

func (s *Service) Apply(req ApplyRequest) (*models.Candidate, error) {
	if req.ResumeID == "" {
		return nil, fmt.Errorf("简历ID不能为空")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("姓名不能为空")
	}
	if req.Email == "" {
		return nil, fmt.Errorf("邮箱不能为空")
	}
	if req.Phone == "" {
		return nil, fmt.Errorf("电话不能为空")
	}
	if req.JobID == "" {
		return nil, fmt.Errorf("职位ID不能为空")
	}

	if !s.jobService.Exists(req.JobID) {
		return nil, fmt.Errorf("职位不存在: %s", req.JobID)
	}

	isOpen, err := s.jobService.IsOpen(req.JobID)
	if err != nil {
		return nil, err
	}
	if !isOpen {
		return nil, fmt.Errorf("职位已关闭，无法投递简历")
	}

	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	for _, cand := range db.Candidates {
		if cand.JobID == req.JobID && cand.ResumeID == req.ResumeID {
			return nil, fmt.Errorf("该简历已投递过此职位，不能重复投递")
		}
	}

	candidate := &models.Candidate{
		ID:        utils.GenerateID(),
		ResumeID:  req.ResumeID,
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		JobID:     req.JobID,
		Status:    models.CandidateStatusApplied,
		AppliedAt: time.Now(),
	}

	err = s.store.Transact(func(db *models.Database) error {
		db.Candidates = append(db.Candidates, *candidate)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return candidate, nil
}

func (s *Service) GetByID(id string) (*models.Candidate, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	for i := range db.Candidates {
		if db.Candidates[i].ID == id {
			return &db.Candidates[i], nil
		}
	}

	return nil, fmt.Errorf("候选人不存在: %s", id)
}

func (s *Service) ListByJob(jobID string) ([]models.Candidate, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	var result []models.Candidate
	for _, cand := range db.Candidates {
		if cand.JobID == jobID {
			result = append(result, cand)
		}
	}
	return result, nil
}

func (s *Service) List() ([]models.Candidate, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	return db.Candidates, nil
}

func (s *Service) Screen(id string) error {
	return s.store.Transact(func(db *models.Database) error {
		for i := range db.Candidates {
			if db.Candidates[i].ID == id {
				cand := &db.Candidates[i]

				if s.isRejected(cand.Status) {
					return fmt.Errorf("候选人已淘汰，无法继续流程")
				}

				switch cand.Status {
				case models.CandidateStatusApplied:
					cand.Status = models.CandidateStatusScreening
					return nil
				default:
					return fmt.Errorf("当前状态不支持筛选操作: %s", cand.Status)
				}
			}
		}
		return fmt.Errorf("候选人不存在: %s", id)
	})
}

func (s *Service) Reject(id string, reason string) error {
	if reason == "" {
		return fmt.Errorf("淘汰原因不能为空")
	}

	err := s.store.Transact(func(db *models.Database) error {
		for i := range db.Candidates {
			if db.Candidates[i].ID == id {
				cand := &db.Candidates[i]

				if s.isRejected(cand.Status) {
					return fmt.Errorf("候选人已淘汰")
				}

				cand.Status = models.CandidateStatusRejected
				cand.RejectionReason = reason
				rejectedAt := time.Now()
				cand.RejectedAt = &rejectedAt
				return nil
			}
		}
		return fmt.Errorf("候选人不存在: %s", id)
	})

	if err != nil {
		return err
	}

	cand, err := s.GetByID(id)
	if err != nil {
		return nil
	}
	return s.store.TriggerUpdateCheck(cand.JobID)
}

func (s *Service) UpdateStatus(id string, newStatus models.CandidateStatus) error {
	return s.store.Transact(func(db *models.Database) error {
		for i := range db.Candidates {
			if db.Candidates[i].ID == id {
				cand := &db.Candidates[i]

				if s.isRejected(cand.Status) {
					return fmt.Errorf("候选人已淘汰，无法继续流程")
				}

				cand.Status = newStatus
				return nil
			}
		}
		return fmt.Errorf("候选人不存在: %s", id)
	})
}

func (s *Service) isRejected(status models.CandidateStatus) bool {
	return status == models.CandidateStatusRejected ||
		status == models.CandidateStatusWithdrawn ||
		status == models.CandidateStatusOfferAbandoned
}

func (s *Service) Exists(id string) bool {
	_, err := s.GetByID(id)
	return err == nil
}

func (s *Service) CanInterview(id string) (bool, error) {
	cand, err := s.GetByID(id)
	if err != nil {
		return false, err
	}

	if s.isRejected(cand.Status) {
		return false, fmt.Errorf("候选人已淘汰")
	}

	return cand.Status == models.CandidateStatusScreening ||
		cand.Status == models.CandidateStatusInterview, nil
}
