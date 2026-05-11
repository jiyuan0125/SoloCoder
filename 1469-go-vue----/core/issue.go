package core

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"supervision-log-system/common"
)

type IssueService struct {
	store *Store
}

func NewIssueService(store *Store) *IssueService {
	return &IssueService{store: store}
}

func (s *IssueService) Create(req *common.CreateIssueRequest) (*common.Issue, error) {
	now := time.Now()
	issue := &common.Issue{
		ID:               uuid.NewString(),
		Severity:         req.Severity,
		Description:      req.Description,
		DiscoveredDate:   req.DiscoveredDate,
		ResponsibleUnit:  req.ResponsibleUnit,
		RectificationReq: req.RectificationReq,
		PlanFinishDate:   req.PlanFinishDate,
		Status:           req.Status,
		ReportedToPM:     false,
		Supervisor:       req.Supervisor,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if issue.Severity == common.SeverityMajor {
		issue.ReportedToPM = true
		fmt.Printf("紧急通知：重大问题已报告给项目经理 - 问题ID: %s\n", issue.ID)
	}

	s.store.Lock()
	defer s.store.Unlock()
	s.store.GetIssues()[issue.ID] = issue

	return issue, nil
}

func (s *IssueService) UpdateStatus(req *common.UpdateIssueStatusRequest) error {
	s.store.Lock()
	defer s.store.Unlock()

	issue, exists := s.store.GetIssues()[req.ID]
	if !exists {
		return errors.New("问题不存在")
	}

	issue.Status = req.Status
	issue.UpdatedAt = time.Now()

	return nil
}

func (s *IssueService) CheckOverdue() {
	now := time.Now()
	s.store.Lock()
	defer s.store.Unlock()

	for _, issue := range s.store.GetIssues() {
		if issue.Status != common.IssueStatusResolved &&
			issue.Status != common.IssueStatusClosed &&
			issue.Status != common.IssueStatusOverdue &&
			now.After(issue.PlanFinishDate) {
			issue.Status = common.IssueStatusOverdue
			issue.UpdatedAt = now
		}
	}
}

func (s *IssueService) GetByID(id string) (*common.Issue, error) {
	s.store.RLock()
	defer s.store.RUnlock()

	issue, exists := s.store.GetIssues()[id]
	if !exists {
		return nil, errors.New("问题不存在")
	}

	return issue, nil
}

func (s *IssueService) List() []*common.Issue {
	s.store.RLock()
	defer s.store.RUnlock()

	var results []*common.Issue
	for _, issue := range s.store.GetIssues() {
		results = append(results, issue)
	}

	return results
}

func (s *IssueService) ListBySeverity(severity common.Severity) []*common.Issue {
	s.store.RLock()
	defer s.store.RUnlock()

	var results []*common.Issue
	for _, issue := range s.store.GetIssues() {
		if issue.Severity == severity {
			results = append(results, issue)
		}
	}

	return results
}

func (s *IssueService) ListByStatus(status common.IssueStatus) []*common.Issue {
	s.store.RLock()
	defer s.store.RUnlock()

	var results []*common.Issue
	for _, issue := range s.store.GetIssues() {
		if issue.Status == status {
			results = append(results, issue)
		}
	}

	return results
}
