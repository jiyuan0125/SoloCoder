package service

import (
	"testing"

	"drg-system/internal/model"
	"drg-system/internal/repository"
)

func TestCalculateStandardPayment(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	repo.InitPresetData()
	service := NewPaymentService(repo)

	tests := []struct {
		name     string
		weight   float64
		rate     float64
		expected int64
	}{
		{"基础测试", 1.2345, 6850.123456, int64(845648)},
		{"权重1.0", 1.0, 6850.123456, int64(685012)},
		{"权重0.5", 0.5, 6850.123456, int64(342506)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateStandardPayment(tt.weight, tt.rate)
			if result != tt.expected {
				t.Errorf("期望 %d, 实际 %d", tt.expected, result)
			}
		})
	}
}

func TestCalculateEfficiencyIndex(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	repo.InitPresetData()
	service := NewPaymentService(repo)

	tests := []struct {
		name            string
		avgActualCost   int64
		standardPayment int64
		expected        float64
	}{
		{"正常计算", 685000, 685012, 0.999982481628},
		{"标准支付为零", 100000, 0, 0.0},
		{"效率指数大于1", 800000, 685012, 1.167862752},
		{"效率指数小于1", 500000, 685012, 0.729914220},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateEfficiencyIndex(tt.avgActualCost, tt.standardPayment)
			diff := result - tt.expected
			if diff < -0.0001 || diff > 0.0001 {
				t.Errorf("期望 %.6f, 实际 %.6f", tt.expected, result)
			}
		})
	}
}

func TestCalculateAdjustedPayment(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	repo.InitPresetData()
	service := NewPaymentService(repo)

	standard := int64(685012)

	tests := []struct {
		name            string
		efficiencyIndex float64
		expected        int64
	}{
		{"低于0.85上浮5%", 0.80, int64(719263)},
		{"等于0.85正常支付", 0.85, standard},
		{"正常区间", 1.0, standard},
		{"等于1.15正常支付", 1.15, standard},
		{"高于1.15下浮10%", 1.20, int64(616511)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateAdjustedPayment(standard, tt.efficiencyIndex)
			if result != tt.expected {
				t.Errorf("期望 %d, 实际 %d", tt.expected, result)
			}
		})
	}
}

func TestGroupingClassification(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	repo.InitPresetData()
	service := NewGroupingService(repo)

	tests := []struct {
		name          string
		record        model.MedicalRecord
		expectedGroup string
	}{
		{
			"急性心梗伴MCC",
			model.MedicalRecord{MainDiagnosis: "I21.0", MainProcedure: "", CCFlag: model.CCFlagMCC},
			"GRG19",
		},
		{
			"急性心梗无CC",
			model.MedicalRecord{MainDiagnosis: "I21.0", MainProcedure: "", CCFlag: model.CCFlagNone},
			"GRG20",
		},
		{
			"肺炎伴CC",
			model.MedicalRecord{MainDiagnosis: "J18.9", MainProcedure: "", CCFlag: model.CCFlagCC},
			"GRG23",
		},
		{
			"胆囊切除手术",
			model.MedicalRecord{MainDiagnosis: "K80.2", MainProcedure: "51.23", CCFlag: model.CCFlagNone},
			"GRG12",
		},
		{
			"心衰伴MCC",
			model.MedicalRecord{MainDiagnosis: "I50.9", MainProcedure: "", CCFlag: model.CCFlagMCC},
			"GRG21",
		},
		{
			"未匹配的诊断",
			model.MedicalRecord{MainDiagnosis: "A00.0", MainProcedure: "", CCFlag: model.CCFlagNone},
			"未分组",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ClassifyRecord(tt.record)
			if result != tt.expectedGroup {
				t.Errorf("期望 %s, 实际 %s", tt.expectedGroup, result)
			}
		})
	}
}

func TestSettlementStatusFlow(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	repo.InitPresetData()
	groupingService := NewGroupingService(repo)
	paymentService := NewPaymentService(repo)
	settlementService := NewSettlementService(repo, groupingService, paymentService)

	records := []model.MedicalRecord{
		{
			CaseNo:        "TEST001",
			MainDiagnosis: "I21.0",
			MainProcedure: "",
			CCFlag:        model.CCFlagMCC,
			Age:           65,
			Gender:        "男",
			ActualCost:    850000,
		},
	}

	settlement, err := settlementService.CreateSettlement("H001", "2024-01", records)
	if err != nil {
		t.Fatalf("创建结算失败: %v", err)
	}

	if settlement.Status != model.SettlementStatusDraft {
		t.Errorf("期望状态 draft, 实际 %s", settlement.Status)
	}

	settlement, err = settlementService.SubmitSettlement(settlement.ID)
	if err != nil {
		t.Fatalf("提交失败: %v", err)
	}
	if settlement.Status != model.SettlementStatusSubmitted {
		t.Errorf("期望状态 submitted, 实际 %s", settlement.Status)
	}

	settlement, err = settlementService.ApproveSettlement(settlement.ID)
	if err != nil {
		t.Fatalf("审核通过失败: %v", err)
	}
	if settlement.Status != model.SettlementStatusApproved {
		t.Errorf("期望状态 approved, 实际 %s", settlement.Status)
	}

	settlement, err = settlementService.PublishSettlement(settlement.ID)
	if err != nil {
		t.Fatalf("发布失败: %v", err)
	}
	if settlement.Status != model.SettlementStatusPublished {
		t.Errorf("期望状态 published, 实际 %s", settlement.Status)
	}
}

func TestDRGGroupValidation(t *testing.T) {
	repo := repository.NewInMemoryRepository()
	repo.InitPresetData()
	service := NewGroupingService(repo)

	_, err := service.CreateDRGGroup(model.DRGGroup{Code: "GRG01", Weight: 1.5})
	if err == nil {
		t.Error("期望返回冲突错误，但未返回")
	}

	_, err = service.CreateDRGGroup(model.DRGGroup{Code: "TEST01", Weight: 0})
	if err == nil {
		t.Error("期望返回零权重错误，但未返回")
	}

	_, err = service.CreateDRGGroup(model.DRGGroup{Code: "TEST01", Weight: -1.0})
	if err == nil {
		t.Error("期望返回负权重错误，但未返回")
	}
}
