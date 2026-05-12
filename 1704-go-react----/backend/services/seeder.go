package services

import (
	"log"
	"rehab-system/config"
	"rehab-system/models"
	"time"

	"gorm.io/gorm"
)

func SeedDatabase() {
	var patientCount int64
	config.DB.Model(&models.Patient{}).Count(&patientCount)
	if patientCount > 0 {
		log.Println("Database already seeded, skipping...")
		return
	}

	log.Println("Seeding database with example data...")

	config.DB.Transaction(func(tx *gorm.DB) error {
		patients := []models.Patient{
			{
				Name:      "张三",
				BirthDate: time.Date(1985, 5, 15, 0, 0, 0, 0, time.Local),
				Gender:    "男",
				Phone:     "13800138001",
				Status:    "active",
			},
			{
				Name:      "李四",
				BirthDate: time.Date(1990, 8, 22, 0, 0, 0, 0, time.Local),
				Gender:    "女",
				Phone:     "13800138002",
				Status:    "active",
			},
			{
				Name:      "王五",
				BirthDate: time.Date(1978, 12, 3, 0, 0, 0, 0, time.Local),
				Gender:    "男",
				Phone:     "13800138003",
				Status:    "active",
			},
		}

		for i := range patients {
			if err := tx.Create(&patients[i]).Error; err != nil {
				return err
			}
		}

		plans := []models.Plan{
			{
				PatientID:      patients[0].ID,
				Name:           "膝关节术后康复计划",
				AssessmentType: "joint_range",
				StartDate:      time.Now().AddDate(0, 0, -14),
				DurationWeeks:  8,
				Status:         "active",
				Effectiveness:  "evaluating",
			},
			{
				PatientID:      patients[1].ID,
				Name:           "腰部康复训练计划",
				AssessmentType: "pain_scale",
				StartDate:      time.Now().AddDate(0, 0, -7),
				DurationWeeks:  4,
				Status:         "active",
				Effectiveness:  "evaluating",
			},
		}

		for i := range plans {
			if err := tx.Create(&plans[i]).Error; err != nil {
				return err
			}
		}

		exercises := []models.Exercise{
			{
				PlanID:          plans[0].ID,
				Name:            "膝关节屈伸训练",
				GoalDescription: "恢复膝关节活动范围，增强股四头肌力量",
				FrequencyType:   "daily",
				FrequencyCount:  3,
				DurationMinutes: 15,
				DifficultyLevel: 2,
			},
			{
				PlanID:          plans[0].ID,
				Name:            "直腿抬高训练",
				GoalDescription: "增强大腿前侧肌肉力量",
				FrequencyType:   "daily",
				FrequencyCount:  2,
				DurationMinutes: 10,
				DifficultyLevel: 1,
			},
			{
				PlanID:          plans[0].ID,
				Name:            "靠墙静蹲训练",
				GoalDescription: "增强膝关节稳定性",
				FrequencyType:   "weekly",
				FrequencyCount:  3,
				DurationMinutes: 5,
				DifficultyLevel: 3,
			},
			{
				PlanID:          plans[1].ID,
				Name:            "猫式伸展",
				GoalDescription: "缓解腰部紧张，增加脊柱灵活性",
				FrequencyType:   "daily",
				FrequencyCount:  2,
				DurationMinutes: 10,
				DifficultyLevel: 1,
			},
			{
				PlanID:          plans[1].ID,
				Name:            "桥式运动",
				GoalDescription: "增强核心和臀部肌肉力量",
				FrequencyType:   "daily",
				FrequencyCount:  2,
				DurationMinutes: 8,
				DifficultyLevel: 2,
			},
		}

		for i := range exercises {
			if err := tx.Create(&exercises[i]).Error; err != nil {
				return err
			}
		}

		planService := NewPlanService()

		for planIdx, plan := range plans {
			for exerciseIdx, exercise := range exercises {
				if exercise.PlanID != plan.ID {
					continue
				}

				tasks, err := planService.CalculateTasksForSeeder(&exercise, &plan)
				if err != nil {
					continue
				}

				for _, task := range tasks {
					if err := tx.Create(&task).Error; err != nil {
						continue
					}

					if planIdx == 0 && exerciseIdx == 0 && task.TaskDate.Before(time.Now()) {
						if err := createSampleRecords(tx, &task, &exercises[exerciseIdx], patients[0].ID); err != nil {
							continue
						}
					}
				}
			}

			for i := 0; i < 2; i++ {
				scheduledDate := plan.StartDate.AddDate(0, 0, i*14)
				assessment := models.Assessment{
					PlanID:         plan.ID,
					PatientID:      plan.PatientID,
					AssessmentType: plan.AssessmentType,
					ScheduledDate:  scheduledDate,
					AssessmentDate: scheduledDate,
					InitialScore:   40,
					CurrentScore:   40 + float64(i*10),
					Notes:          "定期评估记录",
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}
				if i > 0 {
					assessment.InitialScore = 40
					assessment.ImprovementPercent = ((assessment.CurrentScore - 40) / 40) * 100
				}
				tx.Create(&assessment)

				indicators := []models.AssessmentIndicator{
					{
						AssessmentID: assessment.ID,
						Name:         "活动范围",
						Value:        60 + float64(i*10),
						PreviousValue: 50 + float64(i*5),
						Change:       10 + float64(i*5),
						Unit:         "度",
					},
					{
						AssessmentID: assessment.ID,
						Name:         "肌力评分",
						Value:        3 + float64(i),
						PreviousValue: 2 + float64(i),
						Change:       1,
						Unit:         "级",
					},
				}
				tx.Create(&indicators)
			}
		}

		return nil
	})

	log.Println("Database seeded successfully!")
}

func createSampleRecords(tx *gorm.DB, task *models.Task, exercise *models.Exercise, patientID uint) error {
	feelings := []string{models.FeelingNormal, models.FeelingEasy, models.FeelingNormal, models.FeelingHard, models.FeelingNormal}
	scores := []int{7, 8, 9, 6, 8}

	for i := 0; i < 1; i++ {
		record := models.TrainingRecord{
			TaskID:            task.ID,
			ExerciseID:        task.ExerciseID,
			PlanID:            task.PlanID,
			PatientID:         patientID,
			CompletedAt:       task.TaskDate,
			ActualDuration:    exercise.DurationMinutes,
			QualityScore:      scores[i%len(scores)],
			SubjectiveFeeling: feelings[i%len(feelings)],
			Notes:             "训练完成",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
	}

	return nil
}

func (s *PlanService) CalculateTasksForSeeder(exercise *models.Exercise, plan *models.Plan) ([]models.Task, error) {
	var tasks []models.Task
	startDate := plan.StartDate
	endDate := startDate.AddDate(0, 0, plan.DurationWeeks*7-1)

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
				taskTime := "08:00"
				if j == 1 {
					taskTime = "14:00"
				} else if j == 2 {
					taskTime = "20:00"
				}

				status := models.TaskStatusPending
				if currentDate.Before(time.Now()) {
					status = models.TaskStatusCompleted
				}

				task := models.Task{
					PlanID:     plan.ID,
					ExerciseID: exercise.ID,
					TaskDate:   currentDate,
					TaskTime:   taskTime,
					Status:     status,
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
