package services

import (
	"fmt"
	"rehab-system/config"
	"rehab-system/models"
	"time"

	"gorm.io/gorm"
)

type TrainingService struct{}

func NewTrainingService() *TrainingService {
	return &TrainingService{}
}

func (s *TrainingService) GetPatientTasks(patientID uint, date time.Time) ([]models.Task, error) {
	today := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	tomorrow := today.AddDate(0, 0, 1)

	planService := NewPlanService()
	_ = planService.UpdateTaskStatuses()

	var tasks []models.Task
	err := config.DB.Preload("Exercise").
		Where("plan_id IN (SELECT id FROM plans WHERE patient_id = ? AND status = 'active') AND status IN (?, ?, ?) AND task_date < ?",
			patientID, models.TaskStatusPending, models.TaskStatusOverdue, models.TaskStatusCompleted, tomorrow).
		Order("CASE status WHEN 'completed' THEN 2 WHEN 'overdue' THEN 1 ELSE 0 END, task_date ASC, task_time ASC").
		Find(&tasks).Error

	return tasks, err
}

func (s *TrainingService) GetTask(id uint) (*models.Task, error) {
	var task models.Task
	err := config.DB.Preload("Exercise").Preload("Plan").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *TrainingService) CreateRecord(record *models.TrainingRecord) error {
	if err := s.validateRecord(record); err != nil {
		return err
	}

	return config.DB.Transaction(func(tx *gorm.DB) error {
		record.CreatedAt = time.Now()
		record.UpdatedAt = time.Now()

		if err := tx.Create(record).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Task{}).
			Where("id = ?", record.TaskID).
			Updates(map[string]interface{}{
				"status":     models.TaskStatusCompleted,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return err
		}

		if err := s.checkAndAdjustDifficulty(tx, record); err != nil {
			return err
		}

		return nil
	})
}

func (s *TrainingService) validateRecord(record *models.TrainingRecord) error {
	if record.QualityScore < 1 || record.QualityScore > 10 {
		return fmt.Errorf("完成质量评分必须在1到10之间")
	}
	if !models.ValidSubjectiveFeelings[record.SubjectiveFeeling] {
		return fmt.Errorf("主观感受必须是: easy, normal, hard, very_hard 之一")
	}
	return nil
}

func (s *TrainingService) checkAndAdjustDifficulty(tx *gorm.DB, record *models.TrainingRecord) error {
	var exercise models.Exercise
	if err := tx.First(&exercise, record.ExerciseID).Error; err != nil {
		return err
	}

	var recentRecords []models.TrainingRecord
	err := tx.Where("exercise_id = ? AND id <= ?", record.ExerciseID, record.ID).
		Order("created_at DESC").
		Limit(5).
		Find(&recentRecords).Error
	if err != nil {
		return err
	}

	if len(recentRecords) >= 2 &&
		recentRecords[0].SubjectiveFeeling == models.FeelingVeryHard &&
		recentRecords[1].SubjectiveFeeling == models.FeelingVeryHard {
		if exercise.DifficultyLevel > 1 {
			if err := s.adjustDifficulty(tx, &exercise, exercise.DifficultyLevel-1, models.ReasonDifficultyDown); err != nil {
				return err
			}
		}
	}

	if len(recentRecords) >= 3 &&
		recentRecords[0].SubjectiveFeeling == models.FeelingEasy &&
		recentRecords[1].SubjectiveFeeling == models.FeelingEasy &&
		recentRecords[2].SubjectiveFeeling == models.FeelingEasy {
		if exercise.DifficultyLevel < 5 {
			if err := s.adjustDifficulty(tx, &exercise, exercise.DifficultyLevel+1, models.ReasonDifficultyUp); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *TrainingService) adjustDifficulty(tx *gorm.DB, exercise *models.Exercise, newLevel int, reason string) error {
	oldLevel := exercise.DifficultyLevel

	log := models.DifficultyLog{
		ExerciseID:    exercise.ID,
		OldDifficulty: oldLevel,
		NewDifficulty: newLevel,
		AdjustReason:  reason,
		AdjustedAt:    time.Now(),
		CreatedAt:     time.Now(),
	}

	if err := tx.Create(&log).Error; err != nil {
		return err
	}

	exercise.DifficultyLevel = newLevel
	exercise.UpdatedAt = time.Now()
	return tx.Save(exercise).Error
}

func (s *TrainingService) GetExerciseRecords(exerciseID uint) ([]models.TrainingRecord, error) {
	var records []models.TrainingRecord
	err := config.DB.Where("exercise_id = ?", exerciseID).
		Order("created_at DESC").
		Find(&records).Error
	return records, err
}

func (s *TrainingService) GetDifficultyHistory(exerciseID uint) ([]models.DifficultyLog, error) {
	var logs []models.DifficultyLog
	err := config.DB.Where("exercise_id = ?", exerciseID).
		Order("adjusted_at ASC").
		Find(&logs).Error
	return logs, err
}
