package handlers

import (
	"errors"
	"net/http"
	"sort"
	"time"

	"smart-exam/database"
	"smart-exam/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ReportHandler struct{}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

func (h *ReportHandler) GetStudentReport(c *gin.Context) {
	studentID := c.Query("student_id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student_id is required"})
		return
	}

	var exams []models.Exam
	if err := database.DB.Where("student_id = ? AND status = ?", studentID, models.ExamStatusCompleted).Order("created_at ASC").Find(&exams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	scoreTrend := make([]map[string]interface{}, 0)
	for _, exam := range exams {
		var totalPossible float64
		var examQuestions []models.ExamQuestion
		database.DB.Where("exam_id = ?", exam.ID).Preload("Question").Find(&examQuestions)
		for _, eq := range examQuestions {
			totalPossible += eq.Question.Score
		}

		accuracy := 0.0
		if totalPossible > 0 {
			accuracy = (exam.TotalScore / totalPossible) * 100
		}

		scoreTrend = append(scoreTrend, map[string]interface{}{
			"exam_id": exam.ID,
			"date":    exam.CreatedAt.Format("2006-01-02"),
			"score":   exam.TotalScore,
			"accuracy": accuracy,
			"correct_count": exam.CorrectCount,
			"total_questions": exam.TotalQuestions,
		})
	}

	var masteries []models.KnowledgeMastery
	if err := database.DB.Where("student_id = ?", studentID).Preload("KnowledgePoint").Find(&masteries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	knowledgeMastery := make([]map[string]interface{}, 0)
	for _, m := range masteries {
		accuracy := 0.0
		if m.TotalAnswered > 0 {
			accuracy = float64(m.CorrectAnswered) / float64(m.TotalAnswered) * 100
		}

		knowledgeMastery = append(knowledgeMastery, map[string]interface{}{
			"knowledge_point_id":   m.KnowledgePointID,
			"knowledge_point_name": m.KnowledgePoint.Name,
			"subject":             m.KnowledgePoint.Subject,
			"total_answered":      m.TotalAnswered,
			"correct_answered":    m.CorrectAnswered,
			"accuracy":            accuracy,
			"mastery_level":       m.MasteryLevel,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"student_id":        studentID,
		"score_trend":       scoreTrend,
		"knowledge_mastery": knowledgeMastery,
		"total_exams":       len(exams),
	})
}

func (h *ReportHandler) GetExamHistory(c *gin.Context) {
	studentID := c.Query("student_id")
	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student_id is required"})
		return
	}

	var exams []models.Exam
	if err := database.DB.Where("student_id = ?", studentID).Order("created_at DESC").Find(&exams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": exams})
}
