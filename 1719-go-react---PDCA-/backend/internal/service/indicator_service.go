package service

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"medical-quality-system/internal/middleware"
	"medical-quality-system/internal/model"
	"medical-quality-system/pkg/db"

	"gorm.io/gorm"
)

type IndicatorService struct{}

func NewIndicatorService() *IndicatorService {
	return &IndicatorService{}
}

func (s *IndicatorService) Create(indicator *model.Indicator) error {
	if indicator.TargetValue == 0 || indicator.WarningValue == 0 {
		return middleware.NewAppError(http.StatusBadRequest, "目标值和预警阈值不能为空")
	}

	var existing model.Indicator
	if err := db.DB.Where("code = ?", indicator.Code).First(&existing).Error; err == nil {
		return middleware.NewAppError(http.StatusConflict, "指标编号已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return db.DB.Create(indicator).Error
}

func (s *IndicatorService) List() ([]model.Indicator, error) {
	var indicators []model.Indicator
	err := db.DB.Find(&indicators).Error
	return indicators, err
}

func (s *IndicatorService) Get(id uint) (*model.Indicator, error) {
	var indicator model.Indicator
	err := db.DB.First(&indicator, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, middleware.NewAppError(http.StatusNotFound, "指标不存在")
	}
	return &indicator, err
}

func (s *IndicatorService) Update(id uint, indicator *model.Indicator) error {
	existing, err := s.Get(id)
	if err != nil {
		return err
	}

	if indicator.Code != "" && indicator.Code != existing.Code {
		var dup model.Indicator
		if err := db.DB.Where("code = ? AND id != ?", indicator.Code, id).First(&dup).Error; err == nil {
			return middleware.NewAppError(http.StatusConflict, "指标编号已存在")
		}
	}

	return db.DB.Model(existing).Updates(indicator).Error
}

func (s *IndicatorService) UpdateTarget(id uint, newTarget float64, reason, approvedBy string) error {
	indicator, err := s.Get(id)
	if err != nil {
		return err
	}

	if reason == "" || approvedBy == "" {
		return middleware.NewAppError(http.StatusBadRequest, "调整原因和审批人不能为空")
	}

	now := time.Now()
	nextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	effectiveMonth := nextMonth.Format("2006-01")

	history := model.IndicatorTargetHistory{
		IndicatorID:    id,
		OldTargetValue: indicator.TargetValue,
		NewTargetValue: newTarget,
		Reason:         reason,
		ApprovedBy:     approvedBy,
		EffectiveMonth: effectiveMonth,
	}

	tx := db.DB.Begin()
	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Model(indicator).Update("target_value", newTarget).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *IndicatorService) Delete(id uint) error {
	_, err := s.Get(id)
	if err != nil {
		return err
	}
	return db.DB.Delete(&model.Indicator{}, id).Error
}

func (s *IndicatorService) GetTargetForMonth(indicatorID uint, month string) (float64, error) {
	var history model.IndicatorTargetHistory
	err := db.DB.Where("indicator_id = ? AND effective_month <= ?", indicatorID, month).
		Order("effective_month DESC").
		First(&history).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		var indicator model.Indicator
		if err := db.DB.First(&indicator, indicatorID).Error; err != nil {
			return 0, err
		}
		return indicator.TargetValue, nil
	}
	return history.NewTargetValue, err
}

func (s *IndicatorService) CreateData(data *model.IndicatorData) error {
	var indicator model.Indicator
	if err := db.DB.First(&indicator, data.IndicatorID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return middleware.NewAppError(http.StatusNotFound, "指标不存在")
		}
		return err
	}

	if data.Value < 0 {
		return middleware.NewAppError(http.StatusBadRequest, "指标值不在合理范围")
	}

	target, err := s.GetTargetForMonth(data.IndicatorID, data.Month)
	if err != nil {
		return err
	}

	data.IsTargetMet = data.Value >= target
	data.IsWarning = data.Value < indicator.WarningValue

	if err := db.DB.Create(data).Error; err != nil {
		return err
	}

	if !data.IsTargetMet {
		if err := s.createImprovementTodo(data); err != nil {
			fmt.Printf("Warning: Failed to create improvement todo: %v\n", err)
		}
	}

	return nil
}

func (s *IndicatorService) createImprovementTodo(data *model.IndicatorData) error {
	var indicator model.Indicator
	if err := db.DB.First(&indicator, data.IndicatorID).Error; err != nil {
		return err
	}

	var existing model.Todo
	err := db.DB.Where(
		"type = ? AND indicator_id = ? AND status IN ? AND deleted_at IS NULL",
		model.TodoTypeImprovement,
		data.IndicatorID,
		[]model.TodoStatus{model.TodoStatusPending, model.TodoStatusProgress},
	).First(&existing).Error
	if err == nil {
		return nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	dueDate := time.Now().AddDate(0, 0, 30).Format("2006-01-02")

	todo := model.Todo{
		Type:        model.TodoTypeImprovement,
		Title:       fmt.Sprintf("改进指标: %s", indicator.Name),
		Description: fmt.Sprintf("指标数据未达标。月份: %s, 数值: %.2f", data.Month, data.Value),
		IndicatorID: &data.IndicatorID,
		Responsible: "质量管理办公室",
		Department:  indicator.SourceDept,
		DueDate:     dueDate,
		Status:      model.TodoStatusPending,
	}

	return db.DB.Create(&todo).Error
}

func (s *IndicatorService) ListData(indicatorID uint) ([]model.IndicatorData, error) {
	var dataList []model.IndicatorData
	err := db.DB.Preload("Indicator").Where("indicator_id = ?", indicatorID).Order("month DESC").Find(&dataList).Error
	return dataList, err
}

func (s *IndicatorService) GetTrend(indicatorID uint, months int) ([]map[string]interface{}, error) {
	var dataList []model.IndicatorData
	err := db.DB.Where("indicator_id = ?", indicatorID).Order("month DESC").Limit(months).Find(&dataList).Error
	if err != nil {
		return nil, err
	}

	target, err := s.GetTargetForMonth(indicatorID, time.Now().Format("2006-01"))
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(dataList))
	for i, data := range dataList {
		result[i] = map[string]interface{}{
			"month":        data.Month,
			"value":        data.Value,
			"target":       target,
			"is_target_met": data.IsTargetMet,
		}
	}

	return result, nil
}

func (s *IndicatorService) GetAverageValue(indicatorID uint, fromMonth, toMonth string) (float64, int, error) {
	var dataList []model.IndicatorData
	err := db.DB.Where("indicator_id = ? AND month >= ? AND month <= ?", indicatorID, fromMonth, toMonth).Find(&dataList).Error
	if err != nil {
		return 0, 0, err
	}

	if len(dataList) == 0 {
		return 0, 0, nil
	}

	var sum float64
	count := 0
	for _, data := range dataList {
		if !math.IsNaN(data.Value) && !math.IsInf(data.Value, 0) {
			sum += data.Value
			count++
		}
	}

	if count == 0 {
		return 0, 0, nil
	}

	return sum / float64(count), count, nil
}
