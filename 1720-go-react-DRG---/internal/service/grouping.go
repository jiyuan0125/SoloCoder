package service

import (
	"errors"
	"strings"

	"drg-system/internal/model"
	"drg-system/internal/repository"
)

type GroupingService struct {
	repo repository.Repository
}

func NewGroupingService(repo repository.Repository) *GroupingService {
	return &GroupingService{repo: repo}
}

func (s *GroupingService) ListDRGGroups() []model.DRGGroup {
	return s.repo.ListDRGGroups()
}

func (s *GroupingService) GetDRGGroup(code string) (*model.DRGGroup, error) {
	return s.repo.GetDRGGroupByCode(code)
}

func (s *GroupingService) CreateDRGGroup(group model.DRGGroup) (*model.DRGGroup, error) {
	if group.Code == "" {
		return nil, errors.New("DRG组编号不能为空")
	}
	if group.Weight <= 0 {
		return nil, errors.New("权重必须大于0")
	}
	return s.repo.CreateDRGGroup(group)
}

func (s *GroupingService) UpdateDRGGroup(code string, group model.DRGGroup) (*model.DRGGroup, error) {
	if group.Weight <= 0 {
		return nil, errors.New("权重必须大于0")
	}
	return s.repo.UpdateDRGGroup(code, group)
}

func (s *GroupingService) DeleteDRGGroup(code string) error {
	return s.repo.DeleteDRGGroup(code)
}

func (s *GroupingService) ListGroupingRules() []model.GroupingRule {
	return s.repo.ListGroupingRules()
}

func (s *GroupingService) CreateGroupingRule(rule model.GroupingRule) (*model.GroupingRule, error) {
	return s.repo.CreateGroupingRule(rule)
}

func (s *GroupingService) ClassifyRecord(record model.MedicalRecord) string {
	rules := s.repo.ListGroupingRules()

	mccRules := make([]model.GroupingRule, 0)
	ccRules := make([]model.GroupingRule, 0)
	noneRules := make([]model.GroupingRule, 0)

	for _, rule := range rules {
		switch rule.CCFlag {
		case model.CCFlagMCC:
			mccRules = append(mccRules, rule)
		case model.CCFlagCC:
			ccRules = append(ccRules, rule)
		default:
			noneRules = append(noneRules, rule)
		}
	}

	var priorityRules []model.GroupingRule
	switch record.CCFlag {
	case model.CCFlagMCC:
		priorityRules = append(mccRules, append(ccRules, noneRules...)...)
	case model.CCFlagCC:
		priorityRules = append(ccRules, noneRules...)
	default:
		priorityRules = noneRules
	}

	for _, rule := range priorityRules {
		if rule.DiagnosisPrefix != "" && !strings.HasPrefix(record.MainDiagnosis, rule.DiagnosisPrefix) {
			continue
		}
		if rule.ProcedurePrefix != "" && !strings.HasPrefix(record.MainProcedure, rule.ProcedurePrefix) {
			continue
		}
		if rule.CCFlag != model.CCFlagNone && rule.CCFlag != record.CCFlag {
			continue
		}
		if rule.AgeMin > 0 && record.Age < rule.AgeMin {
			continue
		}
		if rule.AgeMax > 0 && record.Age > rule.AgeMax {
			continue
		}
		if rule.Gender != "" && rule.Gender != record.Gender {
			continue
		}
		return rule.DRGGroupCode
	}

	return "未分组"
}
