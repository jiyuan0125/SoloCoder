package service

import (
	"errors"
	"net/http"
	"time"

	"medical-quality-system/internal/middleware"
	"medical-quality-system/internal/model"
	"medical-quality-system/pkg/db"

	"gorm.io/gorm"
)

type PDCAService struct{}

func NewPDCAService() *PDCAService {
	return &PDCAService{}
}

var phaseOrder = map[model.PDCAPhase]int{
	model.PhasePlan:  1,
	model.PhaseDo:    2,
	model.PhaseCheck: 3,
	model.PhaseAct:   4,
}

var nextPhase = map[model.PDCAPhase]model.PDCAPhase{
	model.PhasePlan:  model.PhaseDo,
	model.PhaseDo:    model.PhaseCheck,
	model.PhaseCheck: model.PhaseAct,
}

var phaseToStatus = map[model.PDCAPhase]model.PDCAStatus{
	model.PhasePlan:  model.StatusPlanning,
	model.PhaseDo:    model.StatusExecuting,
	model.PhaseCheck: model.StatusChecking,
	model.PhaseAct:   model.StatusCompleted,
}

func (s *PDCAService) Create(pdca *model.PDCA) error {
	if pdca.Name == "" || pdca.Responsible == "" || pdca.StartDate == "" {
		return middleware.NewAppError(http.StatusBadRequest, "项目名称、负责人和开始日期不能为空")
	}

	pdca.CurrentPhase = model.PhasePlan
	pdca.Status = model.StatusPlanning

	return db.DB.Create(pdca).Error
}

func (s *PDCAService) List() ([]model.PDCA, error) {
	var pdcaList []model.PDCA
	err := db.DB.Preload("Indicator").Preload("ParentPDCA").Order("created_at DESC").Find(&pdcaList).Error
	return pdcaList, err
}

func (s *PDCAService) Get(id uint) (*model.PDCA, error) {
	var pdca model.PDCA
	err := db.DB.Preload("Indicator").Preload("ParentPDCA").First(&pdca, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, middleware.NewAppError(http.StatusNotFound, "PDCA项目不存在")
	}
	return &pdca, err
}

func (s *PDCAService) Update(id uint, pdca *model.PDCA) error {
	existing, err := s.Get(id)
	if err != nil {
		return err
	}
	return db.DB.Model(existing).Updates(pdca).Error
}

func (s *PDCAService) Delete(id uint) error {
	_, err := s.Get(id)
	if err != nil {
		return err
	}
	return db.DB.Delete(&model.PDCA{}, id).Error
}

func (s *PDCAService) CreatePhaseDetail(detail *model.PDCAPhaseDetail) error {
	if detail.Content == "" {
		return middleware.NewAppError(http.StatusBadRequest, "阶段内容描述不能为空")
	}

	var existing model.PDCAPhaseDetail
	err := db.DB.Where("pdca_id = ? AND phase = ?", detail.PDCAID, detail.Phase).First(&existing).Error
	if err == nil {
		return middleware.NewAppError(http.StatusConflict, "该阶段已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	detail.Completed = false
	return db.DB.Create(detail).Error
}

func (s *PDCAService) UpdatePhaseDetail(pdcaID uint, phase model.PDCAPhase, updates *model.PDCAPhaseDetail) error {
	var existing model.PDCAPhaseDetail
	err := db.DB.Where("pdca_id = ? AND phase = ?", pdcaID, phase).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return middleware.NewAppError(http.StatusNotFound, "阶段不存在")
	}

	if updates.Content != "" {
		existing.Content = updates.Content
	}
	if updates.CompleteDate != "" {
		existing.CompleteDate = updates.CompleteDate
	}
	if updates.Evidence != "" {
		existing.Evidence = updates.Evidence
	}

	return db.DB.Save(&existing).Error
}

func (s *PDCAService) NextPhase(pdcaID uint) error {
	pdca, err := s.Get(pdcaID)
	if err != nil {
		return err
	}

	if pdca.CurrentPhase == model.PhaseAct {
		return middleware.NewAppError(http.StatusBadRequest, "已完成所有阶段，无法继续推进")
	}

	var currentDetail model.PDCAPhaseDetail
	if err := db.DB.Where("pdca_id = ? AND phase = ?", pdcaID, pdca.CurrentPhase).First(&currentDetail).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return middleware.NewAppError(http.StatusBadRequest, "当前阶段未填写内容")
		}
		return err
	}

	if currentDetail.Content == "" {
		return middleware.NewAppError(http.StatusBadRequest, "当前阶段内容描述不能为空")
	}

	currentDetail.Completed = true
	if err := db.DB.Save(&currentDetail).Error; err != nil {
		return err
	}

	next, ok := nextPhase[pdca.CurrentPhase]
	if !ok {
		return middleware.NewAppError(http.StatusBadRequest, "无效的阶段")
	}

	pdca.CurrentPhase = next
	if status, ok := phaseToStatus[next]; ok {
		pdca.Status = status
	}

	if next == model.PhaseCheck && pdca.IndicatorID != nil {
		isImproved, err := s.checkImprovement(*pdca.IndicatorID)
		if err != nil {
			return err
		}
		if isImproved {
			pdca.IsImproved = true
		}
	}

	return db.DB.Save(pdca).Error
}

func (s *PDCAService) checkImprovement(indicatorID uint) (bool, error) {
	now := time.Now()
	currentMonth := now.Format("2006-01")
	prevMonth := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, time.UTC).Format("2006-01")

	var currentData, prevData model.IndicatorData
	err := db.DB.Where("indicator_id = ? AND month = ?", indicatorID, currentMonth).First(&currentData).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return false, err
		}
	} else {
		if currentData.IsTargetMet {
			return true, nil
		}
	}

	err = db.DB.Where("indicator_id = ? AND month = ?", indicatorID, prevMonth).First(&prevData).Error
	if err != nil {
		return false, nil
	}

	indicatorSvc := NewIndicatorService()
	target, err := indicatorSvc.GetTargetForMonth(indicatorID, currentMonth)
	if err != nil {
		return false, err
	}

	return (currentData.Value > prevData.Value && currentData.Value >= target), nil
}

func (s *PDCAService) StartNextCycle(parentID uint) (*model.PDCA, error) {
	parent, err := s.Get(parentID)
	if err != nil {
		return nil, err
	}

	if parent.CurrentPhase != model.PhaseAct {
		return nil, middleware.NewAppError(http.StatusBadRequest, "只有完成所有阶段的项目才能启动下一轮循环")
	}

	newPDCA := &model.PDCA{
		Name:         parent.Name + " (下一轮)",
		IndicatorID:  parent.IndicatorID,
		Responsible:  parent.Responsible,
		StartDate:    time.Now().Format("2006-01-02"),
		CurrentPhase: model.PhasePlan,
		Status:       model.StatusPlanning,
		ParentPDCAID: &parentID,
	}

	if err := db.DB.Create(newPDCA).Error; err != nil {
		return nil, err
	}

	parent.Status = model.StatusClosed
	if err := db.DB.Save(parent).Error; err != nil {
		return nil, err
	}

	return newPDCA, nil
}

func (s *PDCAService) GetPhaseDetails(pdcaID uint) ([]model.PDCAPhaseDetail, error) {
	var details []model.PDCAPhaseDetail
	err := db.DB.Where("pdca_id = ?", pdcaID).Find(&details).Error
	return details, err
}
