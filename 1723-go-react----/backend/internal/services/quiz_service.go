package services

import (
	"learning-platform/internal/models"
	"learning-platform/internal/storage"
	"learning-platform/internal/utils"
	"time"
)

type QuizService struct {
	storage        *storage.Storage
	achievementSvc *AchievementService
}

func NewQuizService(s *storage.Storage, as *AchievementService) *QuizService {
	return &QuizService{
		storage:        s,
		achievementSvc: as,
	}
}

func (s *QuizService) CanTakeQuiz(studentID, courseID string) bool {
	progress, exists := s.storage.GetLearningProgress(studentID, courseID)
	if !exists {
		return false
	}

	course, exists := s.storage.GetCourse(courseID)
	if !exists {
		return false
	}

	return progress.ProgressPercentage >= 100 || len(progress.CompletedUnitIDs) == len(course.Units)
}

func (s *QuizService) GenerateQuiz(studentID, courseID string) ([]models.Question, error) {
	if !s.CanTakeQuiz(studentID, courseID) {
		return nil, &QuizLockedError{CourseID: courseID}
	}

	course, exists := s.storage.GetCourse(courseID)
	if !exists {
		return nil, &storage.NotFoundError{ID: courseID, Type: "course"}
	}

	prevAttempts := s.storage.GetQuizAttempts(studentID, courseID)
	usedQuestions := []string{}
	for _, attempt := range prevAttempts {
		usedQuestions = append(usedQuestions, attempt.QuestionIDs...)
	}

	selected := ShuffleQuestions(course.Quiz.Questions, usedQuestions)
	return selected, nil
}

func (s *QuizService) SubmitQuiz(studentID, courseID string, answers []int, selectedQuestions []models.Question) (*models.QuizAttempt, error) {
	if !s.CanTakeQuiz(studentID, courseID) {
		return nil, &QuizLockedError{CourseID: courseID}
	}

	course, exists := s.storage.GetCourse(courseID)
	if !exists {
		return nil, &storage.NotFoundError{ID: courseID, Type: "course"}
	}

	if len(answers) != len(selectedQuestions) {
		return nil, &InvalidAnswerCountError{Expected: len(selectedQuestions), Actual: len(answers)}
	}

	correctCount := 0
	questionIDs := []string{}
	for i, q := range selectedQuestions {
		questionIDs = append(questionIDs, q.ID)
		if answers[i] == q.CorrectIndex {
			correctCount++
		}
	}

	totalCount := len(selectedQuestions)
	passed := correctCount >= 8
	score := float64(correctCount) / float64(totalCount) * 100

	attempt := &models.QuizAttempt{
		ID:           utils.GenerateUUID(),
		StudentID:    studentID,
		CourseID:     courseID,
		QuizID:       course.Quiz.ID,
		QuestionIDs:  questionIDs,
		Answers:      answers,
		CorrectCount: correctCount,
		TotalCount:   totalCount,
		Passed:       passed,
		Score:        score,
		AttemptedAt:  time.Now(),
	}

	if err := s.storage.CreateQuizAttempt(attempt); err != nil {
		return nil, err
	}

	if passed {
		enrollments := s.storage.GetEnrollmentsByStudent(studentID)
		for _, e := range enrollments {
			if e.CourseID == courseID && e.Status == "active" {
				now := time.Now()
				e.Status = "completed"
				e.CompletedAt = &now
				s.storage.UpdateOrder(nil)
			}
		}

		s.achievementSvc.CheckPerfectScore(studentID, score)
	}

	return attempt, nil
}

func (s *QuizService) GetQuizAttempts(studentID, courseID string) []*models.QuizAttempt {
	return s.storage.GetQuizAttempts(studentID, courseID)
}

type QuizLockedError struct {
	CourseID string
}

func (e *QuizLockedError) Error() string {
	return "quiz is locked for course " + e.CourseID + " - complete all units first"
}

type InvalidAnswerCountError struct {
	Expected int
	Actual   int
}

func (e *InvalidAnswerCountError) Error() string {
	return "invalid answer count"
}
