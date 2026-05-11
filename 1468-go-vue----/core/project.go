package core

import (
	"time"

	"github.com/google/uuid"

	"tender-management/common"
)

const (
	minOpenTimeGap = 2 * time.Hour
)

type ProjectService struct {
	store *Store
}

func NewProjectService(store *Store) *ProjectService {
	return &ProjectService{store: store}
}

func (s *ProjectService) CreateProject(req common.CreateProjectRequest) (*common.Project, error) {
	if req.Method == common.InviteTender && len(req.InvitedSuppliers) == 0 {
		return nil, ErrInviteRequiresSuppliers
	}
	if !req.OpenTime.After(req.BidDeadline.Add(minOpenTimeGap)) {
		return nil, ErrOpenTimeTooEarly
	}

	now := time.Now()
	project := &common.Project{
		ID:               uuid.New().String(),
		Name:             req.Name,
		Description:      req.Description,
		Method:           req.Method,
		BidDeadline:      req.BidDeadline,
		OpenTime:         req.OpenTime,
		EvaluationMethod: req.EvaluationMethod,
		InvitedSuppliers: req.InvitedSuppliers,
		Status:           common.ProjectDraft,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.store.SaveProject(project)
	return project, nil
}

func (s *ProjectService) UpdateProject(projectID string, req common.UpdateProjectRequest) error {
	project, ok := s.store.GetProject(projectID)
	if !ok {
		return ErrProjectNotFound
	}
	if project.Status != common.ProjectDraft {
		return ErrProjectNotDraft
	}

	updated := false
	if req.Name != "" {
		project.Name = req.Name
		updated = true
	}
	if req.Description != "" {
		project.Description = req.Description
		updated = true
	}
	if req.Method != "" {
		project.Method = req.Method
		updated = true
	}
	if req.EvaluationMethod != "" {
		project.EvaluationMethod = req.EvaluationMethod
		updated = true
	}
	if req.InvitedSuppliers != nil {
		project.InvitedSuppliers = req.InvitedSuppliers
		updated = true
	}

	if req.BidDeadline != nil {
		project.BidDeadline = *req.BidDeadline
		updated = true
	}
	if req.OpenTime != nil {
		project.OpenTime = *req.OpenTime
		updated = true
	}

	if req.BidDeadline != nil || req.OpenTime != nil {
		if !project.OpenTime.After(project.BidDeadline.Add(minOpenTimeGap)) {
			return ErrOpenTimeTooEarly
		}
	}

	if project.Method == common.InviteTender && len(project.InvitedSuppliers) == 0 {
		return ErrInviteRequiresSuppliers
	}

	if updated {
		project.UpdatedAt = time.Now()
		s.store.SaveProject(project)
	}
	return nil
}

func (s *ProjectService) PublishProject(projectID string) error {
	project, ok := s.store.GetProject(projectID)
	if !ok {
		return ErrProjectNotFound
	}
	if project.Status != common.ProjectDraft {
		return ErrProjectNotDraft
	}
	if project.Method == common.InviteTender && len(project.InvitedSuppliers) == 0 {
		return ErrInviteRequiresSuppliers
	}

	project.Status = common.ProjectPublished
	project.UpdatedAt = time.Now()
	s.store.SaveProject(project)
	return nil
}

func (s *ProjectService) GetProject(projectID string) (*common.Project, error) {
	project, ok := s.store.GetProject(projectID)
	if !ok {
		return nil, ErrProjectNotFound
	}
	return project, nil
}

func (s *ProjectService) ListProjectsForSupplier(supplierID string) []common.Project {
	all := s.store.ListProjects()
	filtered := make([]common.Project, 0)
	for _, p := range all {
		if p.Method == common.PublicTender {
			filtered = append(filtered, p)
		} else if p.Method == common.InviteTender {
			for _, invited := range p.InvitedSuppliers {
				if invited == supplierID {
					filtered = append(filtered, p)
					break
				}
			}
		}
	}
	return filtered
}

func (s *ProjectService) ListAllProjects() []common.Project {
	return s.store.ListProjects()
}
