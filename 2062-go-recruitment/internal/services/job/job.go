package job

import (
	"fmt"
	"time"

	"recruitment/internal/models"
	"recruitment/internal/store"
	"recruitment/pkg/utils"
)

type Service struct {
	store *store.Store
}

func NewService(s *store.Store) *Service {
	return &Service{store: s}
}

type CreateJobRequest struct {
	Title        string
	Department   string
	Requirements string
	SalaryMin    float64
	SalaryMax    float64
}

func (s *Service) Create(req CreateJobRequest) (*models.Job, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("职位标题不能为空")
	}
	if req.Department == "" {
		return nil, fmt.Errorf("部门不能为空")
	}
	if req.SalaryMin < 0 || req.SalaryMax < 0 {
		return nil, fmt.Errorf("薪资不能为负数")
	}
	if req.SalaryMin > req.SalaryMax {
		return nil, fmt.Errorf("最低薪资不能高于最高薪资")
	}

	job := &models.Job{
		ID:           utils.GenerateID(),
		Title:        req.Title,
		Department:   req.Department,
		Requirements: req.Requirements,
		SalaryMin:    req.SalaryMin,
		SalaryMax:    req.SalaryMax,
		Status:       models.JobStatusOpen,
		CreatedAt:    time.Now(),
	}

	err := s.store.Transact(func(db *models.Database) error {
		db.Jobs = append(db.Jobs, *job)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return job, nil
}

func (s *Service) GetByID(id string) (*models.Job, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}

	for i := range db.Jobs {
		if db.Jobs[i].ID == id {
			return &db.Jobs[i], nil
		}
	}

	return nil, fmt.Errorf("职位不存在: %s", id)
}

func (s *Service) List() ([]models.Job, error) {
	db, err := s.store.Load()
	if err != nil {
		return nil, err
	}
	return db.Jobs, nil
}

func (s *Service) Close(id string) error {
	err := s.store.Transact(func(db *models.Database) error {
		for i := range db.Jobs {
			if db.Jobs[i].ID == id {
				if db.Jobs[i].Status == models.JobStatusClosed {
					return fmt.Errorf("职位已关闭")
				}
				db.Jobs[i].Status = models.JobStatusClosed
				closedAt := time.Now()
				db.Jobs[i].ClosedAt = &closedAt
				return nil
			}
		}
		return fmt.Errorf("职位不存在: %s", id)
	})

	if err != nil {
		return err
	}

	return s.store.TriggerUpdateCheck(id)
}

func (s *Service) IsOpen(id string) (bool, error) {
	job, err := s.GetByID(id)
	if err != nil {
		return false, err
	}
	return job.Status == models.JobStatusOpen, nil
}

func (s *Service) Exists(id string) bool {
	_, err := s.GetByID(id)
	return err == nil
}
