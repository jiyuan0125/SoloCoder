package handlers

import (
	"net/http"
	"tdm-system/pkg/models"
	"tdm-system/pkg/storage"

	"github.com/gin-gonic/gin"
)

type StatsResponse struct {
	TrainingHours map[models.Title]HoursStats `json:"training_hours"`
	Evaluations   EvaluationStats             `json:"evaluations"`
	ReviewRate    ReviewStats                 `json:"review_rate"`
}

type HoursStats struct {
	Completed float64 `json:"completed"`
	Required  float64 `json:"required"`
	Count     int     `json:"count"`
}

type EvaluationStats struct {
	Total     int     `json:"total"`
	Qualified int     `json:"qualified"`
	AvgScore  float64 `json:"avg_score"`
}

type ReviewStats struct {
	Total  int     `json:"total"`
	Passed int     `json:"passed"`
	Rate   float64 `json:"rate"`
}

func calcRegistrationHours(t models.Training, r models.Registration) (valid bool, hours float64) {
	if r.Attendance == nil || len(r.Attendance) == 0 || r.Score == nil {
		return false, 0
	}
	if *r.Score < 60 {
		return false, 0
	}
	lateCount := 0
	leaveCount := 0
	absentCount := 0
	for _, item := range r.Attendance {
		switch item.Status {
		case models.ASLate:
			lateCount++
		case models.ASLeave:
			leaveCount++
		case models.ASAbsent:
			absentCount++
		}
	}
	effectiveAbsent := absentCount + (lateCount+1)/2
	leaveExemption := false
	if r.StudyReport != nil && *r.StudyReport {
		maxLeave := float64(t.Hours) / 4.0
		if float64(leaveCount) <= maxLeave {
			leaveExemption = true
		}
	}
	total := len(r.Attendance)
	if total == 0 {
		return false, 0
	}
	attended := total - effectiveAbsent - leaveCount
	if leaveExemption {
		attended = attended + leaveCount
	}
	rate := float64(attended) / float64(total)
	if rate < 0.8 {
		return false, 0
	}
	return true, float64(t.Hours)
}

func GetStats(c *gin.Context) {
	s := storage.Get()
	teacherMap := make(map[string]models.Teacher)
	for _, t := range s.AllTeachers() {
		teacherMap[t.ID] = t
	}
	trainingMap := make(map[string]models.Training)
	for _, t := range s.AllTrainings() {
		trainingMap[t.ID] = t
	}

	hoursStats := map[models.Title]HoursStats{
		models.TitleAssistant: {Required: 60},
		models.TitleLecturer:  {Required: 40},
		models.TitleAssociate: {Required: 20},
		models.TitleProfessor: {Required: 20},
	}

	allRegs := make([]models.Registration, 0)
	for _, t := range s.AllTeachers() {
		for _, r := range s.RegistrationsByTeacher(t.ID) {
			allRegs = append(allRegs, r)
		}
	}

	teacherHours := make(map[string]float64)
	for _, r := range allRegs {
		t, ok := trainingMap[r.TrainingID]
		if !ok {
			continue
		}
		valid, h := calcRegistrationHours(t, r)
		if valid {
			teacherHours[r.TeacherID] += h
		}
	}

	for tid, hours := range teacherHours {
		teacher, ok := teacherMap[tid]
		if !ok {
			continue
		}
		hs := hoursStats[teacher.CurrentTitle]
		hs.Completed += hours
		hs.Count++
		hoursStats[teacher.CurrentTitle] = hs
	}

	evals := s.AllEvaluations()
	evalStats := EvaluationStats{Total: len(evals)}
	sumScore := 0.0
	for _, e := range evals {
		if e.Qualified {
			evalStats.Qualified++
		}
		sumScore += e.FinalScore
	}
	if len(evals) > 0 {
		evalStats.AvgScore = sumScore / float64(len(evals))
	}

	apps := s.AllApplications()
	reviewStats := ReviewStats{Total: len(apps)}
	for _, a := range apps {
		if a.FinalStatus != nil && *a.FinalStatus {
			reviewStats.Passed++
		}
	}
	if len(apps) > 0 {
		reviewStats.Rate = float64(reviewStats.Passed) / float64(len(apps))
	}

	c.JSON(http.StatusOK, StatsResponse{
		TrainingHours: hoursStats,
		Evaluations:   evalStats,
		ReviewRate:    reviewStats,
	})
}

func CreateTeacher(c *gin.Context) {
	var t models.Teacher
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s := storage.Get()
	if t.ID == "" {
		t.ID = s.NextID()
	}
	s.SaveTeacher(t)
	c.JSON(http.StatusCreated, t)
}

func ListTeachers(c *gin.Context) {
	s := storage.Get()
	c.JSON(http.StatusOK, s.AllTeachers())
}

func GetTeacher(c *gin.Context) {
	id := c.Param("id")
	s := storage.Get()
	t, ok := s.GetTeacher(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "教师不存在"})
		return
	}
	c.JSON(http.StatusOK, t)
}
