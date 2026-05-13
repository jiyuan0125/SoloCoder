package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"qc-process/internal/aql"
	"qc-process/internal/models"
	"qc-process/internal/statemachine"
)

func CreateQCReport(batchNo string, batchSize int, inspectionLevel string, items []models.InspectionItem) (*models.QCReport, error) {
	if batchSize <= 0 {
		return nil, fmt.Errorf("错误：批量大小必须大于 0，当前值: %d", batchSize)
	}

	sampleSize, err := aql.CalculateSampleSize(batchSize, inspectionLevel)
	if err != nil {
		return nil, err
	}

	now := time.Now().Format(time.RFC3339)
	return &models.QCReport{
		ID:              uuid.New().String(),
		BatchNo:         batchNo,
		BatchSize:       batchSize,
		InspectionLevel: inspectionLevel,
		SampleSize:      sampleSize,
		Status:          models.StatusPendingSampling,
		Items:           items,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func TransitionToSampling(report *models.QCReport) error {
	newStatus, err := statemachine.ValidateAndTransition(report.Status, models.StatusSampling)
	if err != nil {
		return err
	}
	report.Status = newStatus
	report.UpdatedAt = time.Now().Format(time.RFC3339)
	return nil
}

func TransitionToTesting(report *models.QCReport) error {
	newStatus, err := statemachine.ValidateAndTransition(report.Status, models.StatusTesting)
	if err != nil {
		return err
	}
	report.Status = newStatus
	report.UpdatedAt = time.Now().Format(time.RFC3339)
	return nil
}

func RecordTestResult(report *models.QCReport, itemName string, actualValue float64) error {
	if report.Status != models.StatusTesting {
		return fmt.Errorf("错误：当前状态 %q 不允许记录检测结果，请先进入检测中状态", report.Status)
	}

	found := false
	for i := range report.Items {
		if report.Items[i].Name == itemName {
			report.Items[i].ActualValue = &actualValue
			report.Items[i].Evaluate()
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("错误：未找到检测项目 %q", itemName)
	}

	report.UpdatedAt = time.Now().Format(time.RFC3339)
	return nil
}

func CompleteTesting(report *models.QCReport, unqualifiedRecords []models.UnqualifiedRecord) error {
	if report.Status != models.StatusTesting {
		return fmt.Errorf("错误：当前状态 %q 不允许完成检测", report.Status)
	}

	for _, record := range unqualifiedRecords {
		if record.Reason == "" {
			return fmt.Errorf("错误：不合格项 %q 必须记录不合格原因", record.ItemName)
		}
	}

	report.UnqualifiedItems = unqualifiedRecords
	report.Status = models.StatusTestComplete
	report.UpdatedAt = time.Now().Format(time.RFC3339)
	return nil
}

func EvaluateResult(report *models.QCReport) (models.QCStatus, error) {
	if report.Status != models.StatusTestComplete {
		return report.Status, fmt.Errorf("错误：当前状态 %q 不允许评估结果", report.Status)
	}

	if len(report.UnqualifiedItems) == 0 {
		report.Status = models.StatusQualified
		report.FinalResult = "合格"
		report.UpdatedAt = time.Now().Format(time.RFC3339)
		return report.Status, nil
	}

	hasMajor := false
	for _, item := range report.UnqualifiedItems {
		if item.Severity == models.SeverityMajor {
			hasMajor = true
			break
		}
	}

	if hasMajor {
		report.Status = models.StatusUnqualified
		report.FinalResult = "不合格（严重不合格）"
		report.UpdatedAt = time.Now().Format(time.RFC3339)
		return report.Status, nil
	}

	report.Status = models.StatusQualified
	report.FinalResult = "合格（轻微/一般不合格）"
	report.UpdatedAt = time.Now().Format(time.RFC3339)
	return report.Status, nil
}

func RequestRecheck(report *models.QCReport) error {
	if report.Recheck != nil && report.Recheck.Attempt {
		return fmt.Errorf("错误：复检最多只能申请一次，已申请过复检")
	}

	if report.Status != models.StatusQualified && report.Status != models.StatusUnqualified {
		return fmt.Errorf("错误：当前状态 %q 不允许申请复检", report.Status)
	}

	report.Recheck = &models.RecheckRecord{Attempt: true}
	report.Status = models.StatusRecheck
	report.UpdatedAt = time.Now().Format(time.RFC3339)
	return nil
}

func PerformRecheck(report *models.QCReport, items []models.InspectionItem) error {
	if report.Status != models.StatusRecheck {
		return fmt.Errorf("错误：当前状态 %q 不允许执行复检", report.Status)
	}

	if report.Recheck == nil || !report.Recheck.Attempt {
		return fmt.Errorf("错误：尚未申请复检")
	}

	for _, newItem := range items {
		found := false
		for i := range report.Items {
			if report.Items[i].Name == newItem.Name {
				report.Items[i].ActualValue = newItem.ActualValue
				report.Items[i].Evaluate()
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("错误：复检项目 %q 不在原始检测项目中", newItem.Name)
		}
	}

	allPassed := true
	for _, item := range report.Items {
		if item.IsQualified != nil && !*item.IsQualified {
			allPassed = false
			break
		}
	}

	report.Recheck.Items = allPassed
	report.Status = models.StatusFinalCheck
	report.UpdatedAt = time.Now().Format(time.RFC3339)
	return nil
}

func FinalCheck(report *models.QCReport) (models.QCStatus, error) {
	if report.Status != models.StatusFinalCheck {
		return report.Status, fmt.Errorf("错误：当前状态 %q 不允许执行终检", report.Status)
	}

	if report.Recheck != nil && report.Recheck.Items {
		report.Status = models.StatusQualified
		report.FinalResult = "合格（复检通过）"
	} else {
		report.Status = models.StatusUnqualified
		report.FinalResult = "不合格（复检未通过）"
	}

	report.UpdatedAt = time.Now().Format(time.RFC3339)
	return report.Status, nil
}

func CreateReworkOrder(report *models.QCReport, reason string) error {
	if report.Status != models.StatusUnqualified {
		return fmt.Errorf("错误：当前状态 %q 不允许创建返工工单，只有不合格批次才能返工", report.Status)
	}

	report.ReworkOrder = &models.ReworkOrder{
		ID:        uuid.New().String(),
		Reason:    reason,
		Completed: false,
	}
	report.UpdatedAt = time.Now().Format(time.RFC3339)
	return nil
}

func CompleteRework(report *models.QCReport) error {
	if report.ReworkOrder == nil {
		return fmt.Errorf("错误：该质检单没有关联的返工工单")
	}

	if report.ReworkOrder.Completed {
		return fmt.Errorf("错误：返工工单已完成")
	}

	report.ReworkOrder.Completed = true
	report.Status = models.StatusPendingSampling
	report.Recheck = nil
	report.UnqualifiedItems = nil
	report.FinalResult = ""

	for i := range report.Items {
		report.Items[i].ActualValue = nil
		report.Items[i].IsQualified = nil
	}

	report.UpdatedAt = time.Now().Format(time.RFC3339)
	return nil
}

func Archive(report *models.QCReport) error {
	if report.Status != models.StatusQualified && report.Status != models.StatusUnqualified {
		return fmt.Errorf("错误：当前状态 %q 不允许归档，只有合格或不合格状态才能归档", report.Status)
	}

	report.Status = models.StatusArchived
	report.UpdatedAt = time.Now().Format(time.RFC3339)
	return nil
}
