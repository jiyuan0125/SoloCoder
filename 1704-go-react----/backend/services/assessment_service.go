package services

import (
	"encoding/json"
	"fmt"
	"rehab-system/config"
	"rehab-system/models"
	"time"

	"gorm.io/gorm"
)

type AssessmentService struct{}

func NewAssessmentService() *AssessmentService {
	return &AssessmentService{}
}

func (s *AssessmentService) List(patientID uint, planID uint) ([]models.Assessment, error) {
	var assessments []models.Assessment
	query := config.DB.Preload("Indicators").Order("scheduled_date ASC")

	if patientID > 0 {
		query = query.Where("patient_id = ?", patientID)
	}
	if planID > 0 {
		query = query.Where("plan_id = ?", planID)
	}

	err := query.Find(&assessments).Error
	return assessments, err
}

func (s *AssessmentService) Get(id uint) (*models.Assessment, error) {
	var assessment models.Assessment
	err := config.DB.Preload("Indicators").First(&assessment, id).Error
	if err != nil {
		return nil, err
	}
	return &assessment, nil
}

func (s *AssessmentService) Record(assessment *models.Assessment) error {
	if err := s.validateAssessment(assessment); err != nil {
		return err
	}

	return config.DB.Transaction(func(tx *gorm.DB) error {
		var existing models.Assessment
		if err := tx.First(&existing, assessment.ID).Error; err != nil {
			return err
		}

		if assessment.AssessmentDate.IsZero() {
			assessment.AssessmentDate = time.Now()
		}

		if err := s.validateAssessmentDate(&existing, assessment.AssessmentDate); err != nil {
			return err
		}

		if err := s.calculateMetrics(tx, assessment, &existing); err != nil {
			return err
		}

		rawData, _ := json.Marshal(assessment.Indicators)
		assessment.RawData = string(rawData)

		if err := tx.Model(&existing).Updates(map[string]interface{}{
			"assessment_date":     assessment.AssessmentDate,
			"initial_score":       assessment.InitialScore,
			"current_score":       assessment.CurrentScore,
			"improvement_percent": assessment.ImprovementPercent,
			"raw_data":            assessment.RawData,
			"notes":               assessment.Notes,
			"updated_at":          time.Now(),
		}).Error; err != nil {
			return err
		}

		if err := s.saveIndicators(tx, existing.ID, assessment.Indicators); err != nil {
			return err
		}

		if err := s.checkPlanEffectiveness(tx, existing.PlanID); err != nil {
			return err
		}

		return nil
	})
}

func (s *AssessmentService) validateAssessment(assessment *models.Assessment) error {
	if assessment.CurrentScore < 0 || assessment.CurrentScore > 100 {
		return fmt.Errorf("评估评分必须在0到100之间")
	}
	for _, indicator := range assessment.Indicators {
		if indicator.Value < 0 || indicator.Value > 100 {
			return fmt.Errorf("指标评分 %s 必须在0到100之间", indicator.Name)
		}
	}
	return nil
}

func (s *AssessmentService) validateAssessmentDate(existing *models.Assessment, assessmentDate time.Time) error {
	scheduled := existing.ScheduledDate
	minDate := scheduled.AddDate(0, 0, -2)
	maxDate := scheduled.AddDate(0, 0, 2)

	if assessmentDate.Before(minDate) || assessmentDate.After(maxDate) {
		return fmt.Errorf("评估日期只能在计划日期前后两天内浮动")
	}
	return nil
}

func (s *AssessmentService) calculateMetrics(tx *gorm.DB, assessment *models.Assessment, existing *models.Assessment) error {
	var previousAssessments []models.Assessment
	err := tx.Where("plan_id = ? AND id < ? AND assessment_date IS NOT NULL", existing.PlanID, existing.ID).
		Order("assessment_date DESC").
		Limit(1).
		Find(&previousAssessments).Error
	if err != nil {
		return err
	}

	if len(previousAssessments) > 0 {
		assessment.InitialScore = previousAssessments[0].InitialScore
	} else {
		assessment.InitialScore = assessment.CurrentScore
	}

	if assessment.InitialScore > 0 {
		assessment.ImprovementPercent = ((assessment.CurrentScore - assessment.InitialScore) / assessment.InitialScore) * 100
	} else {
		assessment.ImprovementPercent = 0
	}

	for i := range assessment.Indicators {
		assessment.Indicators[i].Change = assessment.Indicators[i].Value - assessment.Indicators[i].PreviousValue
	}

	return nil
}

func (s *AssessmentService) saveIndicators(tx *gorm.DB, assessmentID uint, indicators []models.AssessmentIndicator) error {
	if err := tx.Where("assessment_id = ?", assessmentID).Delete(&models.AssessmentIndicator{}).Error; err != nil {
		return err
	}

	for i := range indicators {
		indicators[i].AssessmentID = assessmentID
		indicators[i].CreatedAt = time.Now()
		if err := tx.Create(&indicators[i]).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *AssessmentService) checkPlanEffectiveness(tx *gorm.DB, planID uint) error {
	var assessments []models.Assessment
	err := tx.Where("plan_id = ? AND assessment_date IS NOT NULL", planID).
		Order("assessment_date ASC").
		Find(&assessments).Error
	if err != nil {
		return err
	}

	if len(assessments) < 3 {
		return nil
	}

	lastTwo := assessments[len(assessments)-2:]
	noImprovementCount := 0

	for _, a := range lastTwo {
		if a.ImprovementPercent <= 0 {
			noImprovementCount++
		}
	}

	if noImprovementCount >= 2 {
		if err := tx.Model(&models.Plan{}).
			Where("id = ?", planID).
			Update("effectiveness", models.EffectivenessPoor).Error; err != nil {
			return err
		}
	}

	return nil
}
