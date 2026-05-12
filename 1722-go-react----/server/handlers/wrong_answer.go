package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"smart-exam/database"
	"smart-exam/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WrongAnswerHandler struct{}

func NewWrongAnswerHandler() *WrongAnswerHandler {
	return &WrongAnswerHandler{}
}

func (h *WrongAnswerHandler) List(c *gin.Context) {
	studentID := c.Query("student_id")
	knowledgePointID := c.Query("knowledge_point_id")
	sortBy := c.DefaultQuery("sort_by", "error_count")

	if studentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "student_id is required"})
		return
	}

	var wrongAnswers []models.WrongAnswer
	query := database.DB.Where("student_id = ?", studentID).Preload("Question.KnowledgePoint")

	if knowledgePointID != "" {
		kpID, err := strconv.ParseUint(knowledgePointID, 10, 64)
		if err == nil {
			query = query.Joins("JOIN questions ON questions.id = wrong_answers.question_id").Where("questions.knowledge_point_id = ?", kpID)
		}
	}

	if sortBy == "error_count" {
		query = query.Order("error_count DESC")
	} else if sortBy == "date" {
		query = query.Order("created_at DESC")
	}

	if err := query.Find(&wrongAnswers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wrongAnswers})
}

func (h *WrongAnswerHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		IsReviewed *bool `json:"is_reviewed"`
		IsPriority *bool `json:"is_priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var wrongAnswer models.WrongAnswer
	if err := database.DB.First(&wrongAnswer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "wrong answer not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.IsReviewed != nil {
		wrongAnswer.IsReviewed = *req.IsReviewed
	}
	if req.IsPriority != nil {
		wrongAnswer.IsPriority = *req.IsPriority
	}

	if err := database.DB.Save(&wrongAnswer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": wrongAnswer})
}

func (h *WrongAnswerHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := database.DB.Delete(&models.WrongAnswer{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted successfully"})
}

func (h *WrongAnswerHandler) Add(c *gin.Context) {
	var req struct {
		StudentID  string `json:"student_id" binding:"required"`
		QuestionID uint   `json:"question_id" binding:"required"`
		ExamID     *uint  `json:"exam_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.WrongAnswer
	if err := database.DB.Where("student_id = ? AND question_id = ?", req.StudentID, req.QuestionID).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "question already in wrong answers"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	wrongAnswer := models.WrongAnswer{
		StudentID:      req.StudentID,
		QuestionID:     req.QuestionID,
		ErrorCount:    1,
		LastWrongExamID: req.ExamID,
	}

	if err := database.DB.Create(&wrongAnswer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": wrongAnswer})
}
