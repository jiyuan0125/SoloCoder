package services

import (
	"learning-platform/internal/models"
	"learning-platform/internal/storage"
	"learning-platform/internal/utils"
	"sort"
	"time"
)

type AchievementService struct {
	storage *storage.Storage
}

func NewAchievementService(s *storage.Storage) *AchievementService {
	return &AchievementService{storage: s}
}

const (
	AchievementTypeFirstCourse   = "first_course"
	AchievementTypeConsecutive7  = "consecutive_7_days"
	AchievementTypeTotal100Hours = "total_100_hours"
	AchievementTypePerfectScore  = "perfect_score"
)

func (s *AchievementService) CheckFirstCourse(studentID string) {
	if s.storage.HasAchievement(studentID, AchievementTypeFirstCourse) {
		return
	}

	enrollments := s.storage.GetEnrollmentsByStudent(studentID)
	if len(enrollments) == 1 {
		s.createAchievement(studentID, AchievementTypeFirstCourse, "第一门课", "完成第一门课程的学习")
	}
}

func (s *AchievementService) CheckConsecutiveDays(studentID string) {
	if s.storage.HasAchievement(studentID, AchievementTypeConsecutive7) {
		return
	}

	records := s.storage.GetLearningRecordsByStudent(studentID)
	if len(records) == 0 {
		return
	}

	days := make(map[string]bool)
	for _, r := range records {
		day := r.CompletedAt.Truncate(24 * time.Hour).Format("2006-01-02")
		days[day] = true
	}

	dayList := make([]time.Time, 0, len(days))
	for dayStr := range days {
		day, _ := time.Parse("2006-01-02", dayStr)
		dayList = append(dayList, day)
	}

	sort.Slice(dayList, func(i, j int) bool {
		return dayList[i].Before(dayList[j])
	})

	maxConsecutive := 1
	current := 1
	for i := 1; i < len(dayList); i++ {
		if dayList[i].Sub(dayList[i-1]).Hours() == 24 {
			current++
			if current > maxConsecutive {
				maxConsecutive = current
			}
		} else {
			current = 1
		}
	}

	if maxConsecutive >= 7 {
		s.createAchievement(studentID, AchievementTypeConsecutive7, "连续学习7天", "连续7天坚持学习")
	}
}

func (s *AchievementService) CheckTotalHours(studentID string) {
	if s.storage.HasAchievement(studentID, AchievementTypeTotal100Hours) {
		return
	}

	records := s.storage.GetLearningRecordsByStudent(studentID)
	totalMinutes := 0
	for _, r := range records {
		totalMinutes += r.TimeSpent
	}

	if float64(totalMinutes)/60.0 >= 100.0 {
		s.createAchievement(studentID, AchievementTypeTotal100Hours, "累计100小时", "累计学习时间达到100小时")
	}
}

func (s *AchievementService) CheckPerfectScore(studentID string, score float64) {
	if score >= 100.0 {
		if !s.storage.HasAchievement(studentID, AchievementTypePerfectScore) {
			s.createAchievement(studentID, AchievementTypePerfectScore, "满分通过测验", "测验获得满分")
		}
	}
}

func (s *AchievementService) createAchievement(studentID, typ, name, description string) {
	achievement := &models.Achievement{
		ID:          utils.GenerateUUID(),
		StudentID:   studentID,
		Type:        typ,
		Name:        name,
		Description: description,
		AchievedAt:  time.Now(),
	}
	s.storage.CreateAchievement(achievement)
}

func (s *AchievementService) GetStudentAchievements(studentID string) []*models.Achievement {
	return s.storage.GetAchievementsByStudent(studentID)
}

func (s *AchievementService) GetLearningAnalytics(studentID string) *models.LearningAnalytics {
	records := s.storage.GetLearningRecordsByStudent(studentID)
	enrollments := s.storage.GetEnrollmentsByStudent(studentID)

	totalMinutes := 0
	for _, r := range records {
		totalMinutes += r.TimeSpent
	}

	completedCourses := 0
	activeCourses := 0
	for _, e := range enrollments {
		if e.Status == "completed" {
			completedCourses++
		} else if e.Status == "active" {
			activeCourses++
		}
	}

	learningSpeed := 1.0
	if totalMinutes > 0 {
		learningSpeed = float64(completedCourses) / (float64(totalMinutes) / 60.0)
	}

	attempts := s.storage.GetQuizAttempts(studentID, "")
	totalScore := 0.0
	count := 0
	for _, a := range attempts {
		if a.StudentID == studentID {
			totalScore += a.Score
			count++
		}
	}

	avgScore := 0.0
	if count > 0 {
		avgScore = totalScore / float64(count)
	}

	analytics := &models.LearningAnalytics{
		StudentID:              studentID,
		TotalLearningHours:     float64(totalMinutes) / 60.0,
		LearningSpeed:          utils.RoundFloat(learningSpeed, 2),
		MasteryLevel:           utils.RoundFloat(avgScore, 2),
		AverageQuizScore:       utils.RoundFloat(avgScore, 2),
		CompletedCourses:       completedCourses,
		CurrentActiveCourses:   activeCourses,
		ConsecutiveLearningDays: 0,
	}

	return analytics
}
