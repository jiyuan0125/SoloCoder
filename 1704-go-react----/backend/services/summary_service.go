package services

import (
	"rehab-system/config"
	"rehab-system/models"
	"time"
)

type SummaryService struct{}

func NewSummaryService() *SummaryService {
	return &SummaryService{}
}

type ExerciseSummary struct {
	ExerciseID          uint    `json:"exerciseId"`
	ExerciseName        string  `json:"exerciseName"`
	AvgQualityScore     float64 `json:"avgQualityScore"`
	PlannedDuration     int     `json:"plannedDuration"`
	AvgActualDuration   float64 `json:"avgActualDuration"`
	DurationDeviation   float64 `json:"durationDeviation"`
	TotalCompleted      int     `json:"totalCompleted"`
	DifficultyChanges   []models.DifficultyLog `json:"difficultyChanges"`
}

type AssessmentTrend struct {
	AssessmentDate     time.Time              `json:"assessmentDate"`
	Score              float64                `json:"score"`
	ImprovementPercent float64                `json:"improvementPercent"`
	Indicators         map[string]float64     `json:"indicators"`
}

type PatientSummary struct {
	PatientID        uint               `json:"patientId"`
	PatientName      string             `json:"patientName"`
	PlanName         string             `json:"planName"`
	PlanStartDate    time.Time          `json:"planStartDate"`
	PlanEndDate      time.Time          `json:"planEndDate"`
	PlanProgress     float64            `json:"planProgress"`
	Effectiveness    string             `json:"effectiveness"`
	ExerciseSummaries []ExerciseSummary `json:"exerciseSummaries"`
	AssessmentTrends  []AssessmentTrend `json:"assessmentTrends"`
}

func (s *SummaryService) GetPatientSummary(patientID uint, startDate, endDate time.Time) (*PatientSummary, error) {
	planService := NewPlanService()
	trainingService := NewTrainingService()
	assessmentService := NewAssessmentService()

	plans, err := planService.List(patientID)
	if err != nil {
		return nil, err
	}

	if len(plans) == 0 {
		return nil, nil
	}

	plan := plans[0]
	planEndDate := plan.StartDate.AddDate(0, 0, plan.DurationWeeks*7-1)

	exerciseSummaries, err := s.getExerciseSummaries(&plan, startDate, endDate, trainingService)
	if err != nil {
		return nil, err
	}

	assessmentTrends, err := s.getAssessmentTrends(plan.ID, startDate, endDate, assessmentService)
	if err != nil {
		return nil, err
	}

	progress := s.calculatePlanProgress(&plan)

	patient, err := NewPatientService().Get(patientID)
	if err != nil {
		return nil, err
	}

	return &PatientSummary{
		PatientID:         patientID,
		PatientName:       patient.Name,
		PlanName:          plan.Name,
		PlanStartDate:     plan.StartDate,
		PlanEndDate:       planEndDate,
		PlanProgress:      progress,
		Effectiveness:     plan.Effectiveness,
		ExerciseSummaries: exerciseSummaries,
		AssessmentTrends:  assessmentTrends,
	}, nil
}

func (s *SummaryService) getExerciseSummaries(plan *models.Plan, startDate, endDate time.Time, trainingService *TrainingService) ([]ExerciseSummary, error) {
	var summaries []ExerciseSummary

	for _, exercise := range plan.Exercises {
		records, err := s.getRecordsInRange(exercise.ID, startDate, endDate)
		if err != nil {
			return nil, err
		}

		if len(records) == 0 {
			continue
		}

		var totalQuality int
		var totalDuration int
		for _, r := range records {
			totalQuality += r.QualityScore
			totalDuration += r.ActualDuration
		}

		avgQuality := float64(totalQuality) / float64(len(records))
		avgDuration := float64(totalDuration) / float64(len(records))
		deviation := avgDuration - float64(exercise.DurationMinutes)

		difficultyLogs, err := trainingService.GetDifficultyHistory(exercise.ID)
		if err != nil {
			return nil, err
		}

		summaries = append(summaries, ExerciseSummary{
			ExerciseID:        exercise.ID,
			ExerciseName:      exercise.Name,
			AvgQualityScore:   avgQuality,
			PlannedDuration:   exercise.DurationMinutes,
			AvgActualDuration: avgDuration,
			DurationDeviation: deviation,
			TotalCompleted:    len(records),
			DifficultyChanges: difficultyLogs,
		})
	}

	return summaries, nil
}

func (s *SummaryService) getRecordsInRange(exerciseID uint, startDate, endDate time.Time) ([]models.TrainingRecord, error) {
	var records []models.TrainingRecord
	query := config.DB.Where("exercise_id = ?", exerciseID)

	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	err := query.Order("created_at ASC").Find(&records).Error
	return records, err
}

func (s *SummaryService) getAssessmentTrends(planID uint, startDate, endDate time.Time, assessmentService *AssessmentService) ([]AssessmentTrend, error) {
	assessments, err := assessmentService.List(0, planID)
	if err != nil {
		return nil, err
	}

	var trends []AssessmentTrend
	for _, a := range assessments {
		if a.AssessmentDate.IsZero() {
			continue
		}

		if !startDate.IsZero() && a.AssessmentDate.Before(startDate) {
			continue
		}
		if !endDate.IsZero() && a.AssessmentDate.After(endDate) {
			continue
		}

		indicators := make(map[string]float64)
		for _, ind := range a.Indicators {
			indicators[ind.Name] = ind.Value
		}

		trends = append(trends, AssessmentTrend{
			AssessmentDate:     a.AssessmentDate,
			Score:              a.CurrentScore,
			ImprovementPercent: a.ImprovementPercent,
			Indicators:         indicators,
		})
	}

	return trends, nil
}

func (s *SummaryService) calculatePlanProgress(plan *models.Plan) float64 {
	var totalTasks, completedTasks int64

	config.DB.Model(&models.Task{}).Where("plan_id = ?", plan.ID).Count(&totalTasks)
	config.DB.Model(&models.Task{}).Where("plan_id = ? AND status = ?", plan.ID, models.TaskStatusCompleted).Count(&completedTasks)

	if totalTasks == 0 {
		return 0
	}
	return float64(completedTasks) / float64(totalTasks) * 100
}
