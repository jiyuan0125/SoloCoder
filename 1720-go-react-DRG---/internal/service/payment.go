package service

import (
	"math"

	"drg-system/internal/model"
	"drg-system/internal/repository"
)

type PaymentService struct {
	repo repository.Repository
}

func NewPaymentService(repo repository.Repository) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) GetActiveParam() (*model.PaymentParam, error) {
	return s.repo.GetActivePaymentParam()
}

func (s *PaymentService) CreateParam(param model.PaymentParam) (*model.PaymentParam, error) {
	return s.repo.CreatePaymentParam(param)
}

func (s *PaymentService) ListParams() []model.PaymentParam {
	return s.repo.ListPaymentParams()
}

func (s *PaymentService) CalculateStandardPayment(weight float64, rate float64) int64 {
	standard := weight * rate
	return int64(math.Round(standard * 100))
}

func (s *PaymentService) CalculateEfficiencyIndex(avgActualCost int64, standardPayment int64) float64 {
	if standardPayment <= 0 {
		return 0
	}
	return float64(avgActualCost) / float64(standardPayment)
}

func (s *PaymentService) CalculateAdjustedPayment(standardPayment int64, efficiencyIndex float64) int64 {
	if efficiencyIndex < 0.85 {
		return int64(math.Round(float64(standardPayment) * 1.05))
	} else if efficiencyIndex > 1.15 {
		return int64(math.Round(float64(standardPayment) * 0.90))
	}
	return standardPayment
}

func (s *PaymentService) CalculateGroupSummary(
	groupCode string,
	groupName string,
	records []model.MedicalRecord,
	rate float64,
	weight float64,
) model.GroupSummary {
	caseCount := 0
	var totalActualCost int64 = 0

	for _, r := range records {
		if !r.IsExcluded {
			caseCount++
			totalActualCost += r.ActualCost
		}
	}

	standardPayment := s.CalculateStandardPayment(weight, rate)

	var efficiencyIndex float64
	if caseCount > 0 {
		avgActualCost := totalActualCost / int64(caseCount)
		efficiencyIndex = s.CalculateEfficiencyIndex(avgActualCost, standardPayment)
	}

	adjustedPayment := s.CalculateAdjustedPayment(standardPayment, efficiencyIndex)

	return model.GroupSummary{
		DRGGroupCode:    groupCode,
		DRGGroupName:    groupName,
		CaseCount:       caseCount,
		TotalActualCost: totalActualCost,
		EfficiencyIndex: efficiencyIndex,
		StandardPayment: standardPayment,
		AdjustedPayment: adjustedPayment,
	}
}
