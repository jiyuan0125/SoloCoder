package service

import (
	"errors"
	"fmt"
	"time"

	"drg-system/internal/model"
	"drg-system/internal/repository"
)

type SettlementService struct {
	repo            repository.Repository
	groupingService *GroupingService
	paymentService  *PaymentService
}

func NewSettlementService(
	repo repository.Repository,
	groupingService *GroupingService,
	paymentService *PaymentService,
) *SettlementService {
	return &SettlementService{
		repo:            repo,
		groupingService: groupingService,
		paymentService:  paymentService,
	}
}

func (s *SettlementService) CreateSettlement(hospitalID string, period string, records []model.MedicalRecord) (*model.Settlement, error) {
	now := time.Now()
	settlement := model.Settlement{
		ID:         fmt.Sprintf("st_%d", now.UnixNano()),
		HospitalID: hospitalID,
		Period:     period,
		Status:     model.SettlementStatusDraft,
	}

	processedRecords := make([]model.MedicalRecord, 0, len(records))
	for _, r := range records {
		rec := r
		rec.ID = fmt.Sprintf("mr_%d_%s", now.UnixNano(), rec.CaseNo)
		rec.HospitalID = hospitalID
		rec.SettlementID = settlement.ID

		if rec.ActualCost <= 0 {
			rec.IsExcluded = true
			rec.ExclusionReason = "实际费用必须大于0"
		} else {
			rec.DRGGroupCode = s.groupingService.ClassifyRecord(rec)
		}

		_, err := s.repo.CreateMedicalRecord(rec)
		if err != nil {
			return nil, err
		}
		processedRecords = append(processedRecords, rec)
	}

	settlement.GroupSummaries = s.calculateSummaries(processedRecords)
	settlement.TotalActualCost, settlement.TotalPayment, settlement.Difference = s.calculateTotals(settlement.GroupSummaries)

	return s.repo.CreateSettlement(settlement)
}

func (s *SettlementService) calculateSummaries(records []model.MedicalRecord) []model.GroupSummary {
	groupedRecords := make(map[string][]model.MedicalRecord)
	for _, r := range records {
		if !r.IsExcluded {
			groupedRecords[r.DRGGroupCode] = append(groupedRecords[r.DRGGroupCode], r)
		}
	}

	param, err := s.paymentService.GetActiveParam()
	if err != nil {
		return nil
	}

	summaries := make([]model.GroupSummary, 0)
	for groupCode, recs := range groupedRecords {
		group, err := s.repo.GetDRGGroupByCode(groupCode)
		if err != nil {
			continue
		}
		summary := s.paymentService.CalculateGroupSummary(
			groupCode,
			group.Name,
			recs,
			param.Rate,
			group.Weight,
		)
		summaries = append(summaries, summary)
	}

	return summaries
}

func (s *SettlementService) calculateTotals(summaries []model.GroupSummary) (totalActual int64, totalPayment int64, difference int64) {
	for _, s := range summaries {
		totalActual += s.TotalActualCost
		groupPayment := s.AdjustedPayment * int64(s.CaseCount)
		totalPayment += groupPayment
	}
	difference = totalPayment - totalActual
	return
}

func (s *SettlementService) ListSettlements() []model.Settlement {
	return s.repo.ListSettlements()
}

func (s *SettlementService) GetSettlement(id string) (*model.Settlement, error) {
	return s.repo.GetSettlement(id)
}

func (s *SettlementService) SubmitSettlement(id string) (*model.Settlement, error) {
	settlement, err := s.repo.GetSettlement(id)
	if err != nil {
		return nil, err
	}
	if settlement.Status != model.SettlementStatusDraft {
		return nil, errors.New("只有草稿状态可以提交")
	}
	now := time.Now()
	settlement.Status = model.SettlementStatusSubmitted
	settlement.SubmittedAt = &now
	return s.repo.UpdateSettlement(id, *settlement)
}

func (s *SettlementService) ApproveSettlement(id string) (*model.Settlement, error) {
	settlement, err := s.repo.GetSettlement(id)
	if err != nil {
		return nil, err
	}
	if settlement.Status != model.SettlementStatusSubmitted {
		return nil, errors.New("只有已提交状态可以审核")
	}
	now := time.Now()
	settlement.Status = model.SettlementStatusApproved
	settlement.ApprovedAt = &now
	return s.repo.UpdateSettlement(id, *settlement)
}

func (s *SettlementService) RejectSettlement(id string, comment string) (*model.Settlement, error) {
	settlement, err := s.repo.GetSettlement(id)
	if err != nil {
		return nil, err
	}
	if settlement.Status != model.SettlementStatusSubmitted {
		return nil, errors.New("只有已提交状态可以审核")
	}
	settlement.Status = model.SettlementStatusDraft
	settlement.AuditComment = comment
	return s.repo.UpdateSettlement(id, *settlement)
}

func (s *SettlementService) PublishSettlement(id string) (*model.Settlement, error) {
	settlement, err := s.repo.GetSettlement(id)
	if err != nil {
		return nil, err
	}
	if settlement.Status != model.SettlementStatusApproved {
		return nil, errors.New("只有已审核状态可以发布")
	}
	now := time.Now()
	settlement.Status = model.SettlementStatusPublished
	settlement.PublishedAt = &now
	return s.repo.UpdateSettlement(id, *settlement)
}

func (s *SettlementService) WithdrawSettlement(id string) (*model.Settlement, error) {
	settlement, err := s.repo.GetSettlement(id)
	if err != nil {
		return nil, err
	}
	if settlement.Status != model.SettlementStatusPublished {
		return nil, errors.New("只有已发布状态可以撤回")
	}
	settlement.Status = model.SettlementStatusDraft
	return s.repo.UpdateSettlement(id, *settlement)
}

func (s *SettlementService) RecalculateSettlement(id string) (*model.Settlement, error) {
	settlement, err := s.repo.GetSettlement(id)
	if err != nil {
		return nil, err
	}
	if settlement.Status == model.SettlementStatusPublished {
		return nil, errors.New("已发布的结算不能重新计算")
	}

	records := s.repo.ListMedicalRecordsBySettlement(id)
	settlement.GroupSummaries = s.calculateSummaries(records)
	settlement.TotalActualCost, settlement.TotalPayment, settlement.Difference = s.calculateTotals(settlement.GroupSummaries)

	return s.repo.UpdateSettlement(id, *settlement)
}

func (s *SettlementService) GetStatistics() map[string]interface{} {
	settlements := s.repo.ListSettlements()
	groups := s.repo.ListDRGGroups()

	totalSettlements := len(settlements)
	publishedCount := 0
	draftCount := 0
	submittedCount := 0
	approvedCount := 0
	var totalActualCost int64 = 0
	var totalPayment int64 = 0

	for _, s := range settlements {
		switch s.Status {
		case model.SettlementStatusPublished:
			publishedCount++
		case model.SettlementStatusDraft:
			draftCount++
		case model.SettlementStatusSubmitted:
			submittedCount++
		case model.SettlementStatusApproved:
			approvedCount++
		}
		totalActualCost += s.TotalActualCost
		totalPayment += s.TotalPayment
	}

	return map[string]interface{}{
		"total_settlements": totalSettlements,
		"status_counts": map[string]int{
			"draft":     draftCount,
			"submitted": submittedCount,
			"approved":  approvedCount,
			"published": publishedCount,
		},
		"total_actual_cost": totalActualCost,
		"total_payment":     totalPayment,
		"total_groups":      len(groups),
	}
}
