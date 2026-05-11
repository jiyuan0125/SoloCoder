package core

import (
	"constructionms/common"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Store struct {
	mu       sync.RWMutex
	projects map[string]*common.Project
	counter  int
}

func NewStore() *Store {
	return &Store{
		projects: make(map[string]*common.Project),
		counter:  0,
	}
}

func (s *Store) generateProjectID() string {
	s.counter++
	return fmt.Sprintf("PRJ-%06d", s.counter)
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *Store) CreateProject(req *common.CreateProjectRequest) (*common.Project, error) {
	if req.Name == "" {
		return nil, errors.New("项目名称不能为空")
	}
	if req.ContractFee <= 0 {
		return nil, errors.New("合同金额必须大于0")
	}

	planStart, err := time.Parse("2006-01-02", req.PlanStart)
	if err != nil {
		return nil, errors.New("计划开工日期格式错误")
	}

	planEnd, err := time.Parse("2006-01-02", req.PlanEnd)
	if err != nil {
		return nil, errors.New("计划竣工日期格式错误")
	}

	if planEnd.Before(planStart) {
		return nil, errors.New("计划竣工日期不能早于计划开工日期")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	project := &common.Project{
		ID:             s.generateProjectID(),
		Name:           req.Name,
		Client:         req.Client,
		Contractor:     req.Contractor,
		ContractFee:    req.ContractFee,
		PlanStart:      planStart,
		PlanEnd:        planEnd,
		ProjectManager: req.ProjectManager,
		Milestones:     generateMilestones(planStart, planEnd),
		Expenses:       []common.Expense{},
		QualityChecks:  []common.QualityCheck{},
		CreatedAt:      time.Now(),
	}

	s.projects[project.ID] = project
	return project, nil
}

func generateMilestones(start, end time.Time) []common.Milestone {
	names := []string{"基础工程", "主体结构", "装饰装修", "竣工验收"}
	duration := end.Sub(start)
	intervals := []float64{0.25, 0.50, 0.80, 1.0}

	milestones := make([]common.Milestone, 4)
	for i, name := range names {
		planFinish := start.Add(time.Duration(float64(duration) * intervals[i]))
		milestones[i] = common.Milestone{
			Name:           name,
			PlanFinishDate: planFinish,
			Progress:       0,
			ActualFinishDate: nil,
			IsDelayed:      false,
		}
	}
	return milestones
}

func (s *Store) GetProject(id string) (*common.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	project, exists := s.projects[id]
	if !exists {
		return nil, errors.New("项目不存在")
	}
	return project, nil
}

func (s *Store) ListProjects() []*common.Project {
	s.mu.RLock()
	defer s.mu.RUnlock()

	projects := make([]*common.Project, 0, len(s.projects))
	for _, p := range s.projects {
		projects = append(projects, p)
	}
	return projects
}

func (s *Store) UpdateProgress(projectID string, req *common.UpdateProgressRequest) error {
	if req.Progress < 0 || req.Progress > 100 {
		return errors.New("进度百分比必须在0到100之间")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	project, exists := s.projects[projectID]
	if !exists {
		return errors.New("项目不存在")
	}

	var actualFinish *time.Time
	if req.ActualFinishDate != "" {
		date, err := time.Parse("2006-01-02", req.ActualFinishDate)
		if err != nil {
			return errors.New("实际完成日期格式错误")
		}
		actualFinish = &date
	}

	for i := range project.Milestones {
		if project.Milestones[i].Name == req.MilestoneName {
			project.Milestones[i].Progress = req.Progress
			if actualFinish != nil {
				project.Milestones[i].ActualFinishDate = actualFinish
				project.Milestones[i].IsDelayed = actualFinish.After(project.Milestones[i].PlanFinishDate)
			}
			return nil
		}
	}

	return errors.New("里程碑不存在")
}

func (s *Store) GetOverallProgress(projectID string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	project, exists := s.projects[projectID]
	if !exists {
		return 0, errors.New("项目不存在")
	}

	if len(project.Milestones) == 0 {
		return 0, nil
	}

	weights := []float64{0.20, 0.40, 0.25, 0.15}
	var totalWeight float64
	var progressSum float64

	for i, milestone := range project.Milestones {
		if milestone.Progress > 0 {
			totalWeight += weights[i]
			progressSum += weights[i] * float64(milestone.Progress) / 100.0
		}
	}

	if totalWeight == 0 {
		return 0, nil
	}

	return progressSum / totalWeight, nil
}

func (s *Store) AddExpense(projectID string, req *common.AddExpenseRequest) (*common.Expense, error) {
	if req.Amount <= 0 {
		return nil, errors.New("支出金额必须大于0")
	}

	validCategory := false
	for _, c := range common.ExpenseCategories {
		if c == req.Category {
			validCategory = true
			break
		}
	}
	if !validCategory {
		return nil, errors.New("无效的支出类别")
	}

	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("支出日期格式错误")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	project, exists := s.projects[projectID]
	if !exists {
		return nil, errors.New("项目不存在")
	}

	expense := &common.Expense{
		ID:        generateID(),
		Category:  req.Category,
		Amount:    req.Amount,
		Date:      date,
		VoucherNo: req.VoucherNo,
	}

	project.Expenses = append(project.Expenses, *expense)
	return expense, nil
}

func (s *Store) GetCategoryBudget(contractFee int64, category string) int64 {
	ratio := common.BudgetRatios[category]
	return int64(float64(contractFee) * ratio)
}

func (s *Store) GetCategoryActual(projectID string, category string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	project, exists := s.projects[projectID]
	if !exists {
		return 0, errors.New("项目不存在")
	}

	var total int64
	for _, exp := range project.Expenses {
		if exp.Category == category {
			total += exp.Amount
		}
	}
	return total, nil
}

func (s *Store) IsCategoryOverBudget(projectID string, category string) (bool, error) {
	actual, err := s.GetCategoryActual(projectID, category)
	if err != nil {
		return false, err
	}

	project, err := s.GetProject(projectID)
	if err != nil {
		return false, err
	}

	budget := s.GetCategoryBudget(project.ContractFee, category)
	return actual > budget, nil
}

func (s *Store) AddQualityCheck(projectID string, req *common.AddQualityCheckRequest) (*common.QualityCheck, error) {
	if req.CheckItem == "" {
		return nil, errors.New("检查项目不能为空")
	}

	checkDate, err := time.Parse("2006-01-02", req.CheckDate)
	if err != nil {
		return nil, errors.New("检查日期格式错误")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	project, exists := s.projects[projectID]
	if !exists {
		return nil, errors.New("项目不存在")
	}

	check := &common.QualityCheck{
		ID:            generateID(),
		CheckItem:     req.CheckItem,
		Result:        req.Result,
		IsPass:        req.Result == "合格",
		Inspector:     req.Inspector,
		CheckDate:     checkDate,
		Rectification: nil,
		Closed:        req.Result == "合格",
	}

	project.QualityChecks = append(project.QualityChecks, *check)
	return check, nil
}

func (s *Store) AddRectification(projectID string, checkID string, req *common.RectificationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	project, exists := s.projects[projectID]
	if !exists {
		return errors.New("项目不存在")
	}

	rectDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return errors.New("整改日期格式错误")
	}

	for i := range project.QualityChecks {
		if project.QualityChecks[i].ID == checkID {
			if project.QualityChecks[i].Closed {
				return errors.New("该检查已关闭，无需整改")
			}

			project.QualityChecks[i].Rectification = &common.Rectification{
				ID:        generateID(),
				Result:    req.Result,
				Passed:    req.Result == "合格",
				Inspector: req.Inspector,
				Date:      rectDate,
			}

			if req.Result == "合格" {
				project.QualityChecks[i].Closed = true
			}

			return nil
		}
	}

	return errors.New("质量检查记录不存在")
}

func (s *Store) IsRiskProject(projectID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	project, exists := s.projects[projectID]
	if !exists {
		return false, errors.New("项目不存在")
	}

	failCount := 0
	for _, check := range project.QualityChecks {
		if !check.IsPass {
			failCount++
		}
	}

	return failCount > 5, nil
}
