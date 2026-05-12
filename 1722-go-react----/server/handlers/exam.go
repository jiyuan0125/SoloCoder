package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"smart-exam/database"
	"smart-exam/models"
	"smart-exam/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ExamHandler struct {
	adaptive *services.AdaptiveService
	billing  *services.BillingService
}

func NewExamHandler() *ExamHandler {
	return &ExamHandler{
		adaptive: services.NewAdaptiveService(),
		billing:  services.NewBillingService(),
	}
}

func (h *ExamHandler) StartExam(c *gin.Context) {
	var req struct {
		StudentID string `json:"student_id" binding:"required"`
		BillID    *uint  `json:"bill_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var inProgressExam models.Exam
	if err := database.DB.Where("student_id = ? AND status = ?", req.StudentID, models.ExamStatusInProgress).First(&inProgressExam).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student already has an exam in progress"})
		return
	}

	now := time.Now()
	exam := models.Exam{
		StudentID:       req.StudentID,
		Status:          models.ExamStatusInProgress,
		CurrentQuestion: 0,
		CurrentDifficulty: services.InitialDifficulty,
		ConsecutiveCorrect: 0,
		ConsecutiveWrong: 0,
		TotalQuestions: services.TotalQuestions,
		StartTime:       &now,
		BillID:        req.BillID,
	}

	if err := database.DB.Create(&exam).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pool, err := h.adaptive.GetQuestionPool(exam.CurrentDifficulty, []uint{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	question, err := h.adaptive.PickRandomQuestion(pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	examQuestion := models.ExamQuestion{
		ExamID:          exam.ID,
		QuestionID:    question.ID,
		OrderNumber:   1,
		DifficultyAtTime: exam.CurrentDifficulty,
	}

	if err := database.DB.Create(&examQuestion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	exam.CurrentQuestion = 1
	database.DB.Save(&exam)

	c.JSON(http.StatusCreated, gin.H{
		"exam":      exam,
		"question":  question,
		"question_number": 1,
		"total_questions": exam.TotalQuestions,
	})
}

func (h *ExamHandler) GetCurrentQuestion(c *gin.Context) {
	examIDStr := c.Param("exam_id")
	examID, err := strconv.ParseUint(examIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	var exam models.Exam
	if err := database.DB.First(&exam, examID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "exam not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if exam.Status != models.ExamStatusInProgress {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exam is not in progress"})
		return
	}

	var examQuestion models.ExamQuestion
	if err := database.DB.Where("exam_id = ? AND order_number = ?", examID, exam.CurrentQuestion).Preload("Question").First(&examQuestion).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "current question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exam":              exam,
		"question":         examQuestion.Question,
		"question_number":  exam.CurrentQuestion,
		"total_questions": exam.TotalQuestions,
	})
}

func (h *ExamHandler) SubmitAnswer(c *gin.Context) {
	examIDStr := c.Param("exam_id")
	examID, err := strconv.ParseUint(examIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	var req struct {
		StudentAnswer string `json:"student_answer" binding:"required"`
		TimeSpent    int    `json:"time_spent"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exam models.Exam
	if err := database.DB.First(&exam, examID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "exam not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if exam.Status != models.ExamStatusInProgress {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exam is not in progress"})
		return
	}

	var examQuestion models.ExamQuestion
	if err := database.DB.Where("exam_id = ? AND order_number = ?", examID, exam.CurrentQuestion).Preload("Question").First(&examQuestion).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "current question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	question := examQuestion.Question
	isObjective := h.adaptive.IsObjectiveQuestion(question.Type)
	var isCorrect *bool
	var scoreEarned float64
	needsGrading := false

	if isObjective {
		correct, err := h.adaptive.CheckAnswer(&question, req.StudentAnswer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		isCorrect = &correct
		if correct {
			scoreEarned = question.Score
			exam.CorrectCount++
		}

		if !correct {
			if err := h.addToWrongAnswers(exam.StudentID, question.ID, exam.ID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	} else {
		needsGrading = true
	}

	now := time.Now()
	examQuestion.StudentAnswer = req.StudentAnswer
	examQuestion.IsCorrect = isCorrect
	examQuestion.ScoreEarned = scoreEarned
	examQuestion.TimeSpent = req.TimeSpent
	examQuestion.SubmittedAt = &now
	examQuestion.NeedsGrading = needsGrading
	if err := database.DB.Save(&examQuestion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	exam.TotalScore += scoreEarned

	if isObjective {
		if isCorrect != nil && *isCorrect {
			exam.ConsecutiveCorrect++
			exam.ConsecutiveWrong = 0
			if exam.ConsecutiveCorrect >= 2 {
				if exam.CurrentDifficulty < services.MaxDifficulty {
					exam.CurrentDifficulty++
				}
				exam.ConsecutiveCorrect = 0
			}
		} else {
			exam.ConsecutiveWrong++
			exam.ConsecutiveCorrect = 0
			if exam.ConsecutiveWrong >= 2 {
				if exam.CurrentDifficulty > services.MinDifficulty {
					exam.CurrentDifficulty--
				}
				exam.ConsecutiveWrong = 0
			}
		}
	}

	if err := h.updateKnowledgeMastery(exam.StudentID, question.KnowledgePointID, isCorrect != nil && *isCorrect); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if exam.CurrentQuestion >= exam.TotalQuestions {
		exam.Status = models.ExamStatusCompleted
		endTime := time.Now()
		exam.EndTime = &endTime

		if exam.BillID != nil {
			if err := h.billing.ChargeExam(*exam.BillID, exam.ID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if err := database.DB.Save(&exam).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"exam":        exam,
			"completed": true,
			"next_question": nil,
		})
		return
	}

	var usedQuestionIDs []uint
	database.DB.Where("exam_id = ?", examID).Pluck("question_id", &usedQuestionIDs)

	pool, err := h.adaptive.GetQuestionPool(exam.CurrentDifficulty, usedQuestionIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nextQuestion, err := h.adaptive.PickRandomQuestion(pool)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	nextExamQuestion := models.ExamQuestion{
		ExamID:          exam.ID,
		QuestionID:    nextQuestion.ID,
		OrderNumber:   exam.CurrentQuestion + 1,
		DifficultyAtTime: exam.CurrentDifficulty,
	}

	if err := database.DB.Create(&nextExamQuestion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	exam.CurrentQuestion++
	if err := database.DB.Save(&exam).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exam": exam,
		"current_answer_result": gin.H{
			"is_correct":   isCorrect,
			"score_earned": scoreEarned,
			"needs_grading": needsGrading,
			"explanation": question.Explanation,
		},
		"next_question": nextQuestion,
		"question_number": exam.CurrentQuestion,
		"total_questions": exam.TotalQuestions,
	})
}

func (h *ExamHandler) addToWrongAnswers(studentID string, questionID uint, examID uint) error {
	var wrongAnswer models.WrongAnswer
	err := database.DB.Where("student_id = ? AND question_id = ?", studentID, questionID).First(&wrongAnswer).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		wrongAnswer = models.WrongAnswer{
			StudentID:      studentID,
			QuestionID:     questionID,
			ErrorCount:    1,
			LastWrongExamID: &examID,
		}
		return database.DB.Create(&wrongAnswer).Error
	} else if err != nil {
		return err
	}

	wrongAnswer.ErrorCount++
	wrongAnswer.LastWrongExamID = &examID
	return database.DB.Save(&wrongAnswer).Error
}

func (h *ExamHandler) updateKnowledgeMastery(studentID string, knowledgePointID uint, isCorrect bool) error {
	ancestors, err := h.adaptive.GetKnowledgePointAncestors(knowledgePointID)
	if err != nil {
		return err
	}

	for _, kpID := range ancestors {
		var mastery models.KnowledgeMastery
		err := database.DB.Where("student_id = ? AND knowledge_point_id = ?", studentID, kpID).First(&mastery).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			mastery = models.KnowledgeMastery{
				StudentID:        studentID,
				KnowledgePointID: kpID,
				TotalAnswered:     1,
			}
			if isCorrect {
				mastery.CorrectAnswered = 1
			}
			mastery.MasteryLevel = h.adaptive.CalculateMasteryLevel(mastery.CorrectAnswered, mastery.TotalAnswered)
			mastery.LastUpdated = time.Now()
			return database.DB.Create(&mastery).Error
		} else if err != nil {
			return err
		}

		mastery.TotalAnswered++
		if isCorrect {
			mastery.CorrectAnswered++
		}
		mastery.MasteryLevel = h.adaptive.CalculateMasteryLevel(mastery.CorrectAnswered, mastery.TotalAnswered)
		mastery.LastUpdated = time.Now()
		if err := database.DB.Save(&mastery).Error; err != nil {
			return err
		}
	}

	return nil
}

func (h *ExamHandler) MarkExamIncomplete(c *gin.Context) {
	examIDStr := c.Param("exam_id")
	examID, err := strconv.ParseUint(examIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	var exam models.Exam
	if err := database.DB.First(&exam, examID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "exam not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if exam.Status != models.ExamStatusInProgress {
		c.JSON(http.StatusBadRequest, gin.H{"error": "exam is not in progress"})
		return
	}

	exam.Status = models.ExamStatusIncomplete
	now := time.Now()
	exam.EndTime = &now

	if err := database.DB.Save(&exam).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "exam marked as incomplete"})
}

func (h *ExamHandler) GetExamReport(c *gin.Context) {
	examIDStr := c.Param("exam_id")
	examID, err := strconv.ParseUint(examIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid exam id"})
		return
	}

	var exam models.Exam
	if err := database.DB.First(&exam, examID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "exam not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var examQuestions []models.ExamQuestion
	if err := database.DB.Where("exam_id = ?", examID).Preload("Question").Find(&examQuestions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"exam":     exam,
		"questions": examQuestions,
	})
}
