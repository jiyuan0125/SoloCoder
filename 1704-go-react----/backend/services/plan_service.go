package services

import (
	"fmt"
	"rehab-system/config"
	"rehab-system/models"
	"time"

	"gorm.io/gorm"
)

type PlanService struct{}

func NewPlanService() *PlanService {
	return &PlanService{}
}

func (s *PlanService) Create(plan *models.Plan) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		plan.CreatedAt = time.Now()
		plan.UpdatedAt = time.Now()
		plan.Status = "active"
		plan.Effectiveness = models.EffectivenessEvaluating

		if err := tx.Create(plan).Error; err != nil {
			return err
		}

		if err := s.generateTasks(tx, plan); err != nil {
			return err
		}

		if err := s.scheduleAssessments(tx, plan); err != nil {
			return err
		}

		return nil
	})
}

func (s *PlanService) validateExercise(exercise *models.Exercise) error {
	if exercise.Name == "" {
		return fmt.Errorf("训练项目名称不能为空")
	}
	if exercise.DifficultyLevel < 1 || exercise.DifficultyLevel > 5 {
		return fmt.Errorf("难度等级必须在1到5之间")
	}
	if exercise.FrequencyCount <= 0 {
		return fmt.Errorf("频率必须为正数")
	}
	return nil
}

func (s *PlanService) generateTasks(tx *gorm.DB, plan *models.Plan) error {
	for i, exercise := range plan.Exercises {
		if err := s.validateExercise(&exercise); err != nil {
			return err
		}

		exercise.PlanID = plan.ID
		exercise.CreatedAt = time.Now()
		exercise.UpdatedAt = time.Now()

		if err := tx.Create(&exercise).Error; err != nil {
			return err
		}
		plan.Exercises[i] = exercise

		tasks, err := s.calculateTasks(&exercise, plan)
		if err != nil {
			return err
		}

		for _, task := range tasks {
			if err := tx.Create(&task).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *PlanService) calculateTasks(exercise *models.Exercise, plan *models.Plan) ([]models.Task, error) {
	var tasks []models.Task
	startDate := plan.StartDate
	endDate := startDate.AddDate(0, 0, plan.DurationWeeks*7-1)

	existingTasks := []models.Task{}
	err := config.DB.Where("plan_id = ? AND exercise_id = ?", plan.ID, exercise.ID).Find(&existingTasks).Error
	if err != nil {
		return nil, err
	}
	if len(existingTasks) > 0 {
		return nil, fmt.Errorf("训练任务已存在")
	}

	currentDate := startDate
	for !currentDate.After(endDate) {
		shouldGenerate := false
		switch exercise.FrequencyType {
		case models.FrequencyTypeDaily:
			shouldGenerate = true
		case models.FrequencyTypeWeekly:
			weekday := int(currentDate.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			if weekday <= exercise.FrequencyCount {
				shouldGenerate = true
			}
		}

		if shouldGenerate {
			timesPerDay := 1
			if exercise.FrequencyType == models.FrequencyTypeDaily {
				timesPerDay = exercise.FrequencyCount
			}

			for j := 0; j < timesPerDay; j++ {
				taskTime := fmt.Sprintf("%02d:00", 8+j*3)
				if j*3 > 8 {
					taskTime = fmt.Sprintf("%02d:00", 9+j*2)
				}

				task := models.Task{
					PlanID:     plan.ID,
					ExerciseID: exercise.ID,
					TaskDate:   currentDate,
					TaskTime:   taskTime,
					Status:     models.TaskStatusPending,
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
				tasks = append(tasks, task)
			}
		}
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return tasks, nil
}

func (s *PlanService) scheduleAssessments(tx *gorm.DB, plan *models.Plan) error {
	assessmentInterval := 14
	totalDays := plan.DurationWeeks * 7

	assessmentsNeeded := totalDays / assessmentInterval
	if assessmentsNeeded == 0 {
		assessmentsNeeded = 1
	}

	for i := 0; i < assessmentsNeeded; i++ {
		scheduledDate := plan.StartDate.AddDate(0, 0, i*assessmentInterval)
		if i > 0 {
			scheduledDate = s.adjustForWeekend(scheduledDate)
		}

		assessment := models.Assessment{
			PlanID:         plan.ID,
			PatientID:      plan.PatientID,
			AssessmentType: plan.AssessmentType,
			ScheduledDate:  scheduledDate,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := tx.Create(&assessment).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *PlanService) adjustForWeekend(date time.Time) time.Time {
	weekday := date.Weekday()
	if weekday == time.Saturday {
		return date.AddDate(0, 0, -1)
	}
	if weekday == time.Sunday {
		return date.AddDate(0, 0, 1)
	}
	return date
}

func (s *PlanService) Get(id uint) (*models.Plan, error) {
	var plan models.Plan
	err := config.DB.Preload("Exercises").
		Preload("Exercises.DifficultyLogs").
		First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (s *PlanService) List(patientID uint) ([]models.Plan, error) {
	var plans []models.Plan
	query := config.DB.Preload("Exercises")
	if patientID > 0 {
		query = query.Where("patient_id = ?", patientID)
	}
	err := query.Find(&plans).Error
	return plans, err
}

func (s *PlanService) UpdateTaskStatuses() error {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	threeDaysAgo := today.AddDate(0, 0, -3)

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Task{}).
			Where("status = ? AND task_date < ?", models.TaskStatusPending, today).
			Update("status", models.TaskStatusOverdue).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Task{}).
			Where("status = ? AND task_date < ?", models.TaskStatusOverdue, threeDaysAgo).
			Update("status", models.TaskStatusAbandoned).Error; err != nil {
			return err
		}

		return nil
	})

	return err
}
