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

type QuestionHandler struct{}

func NewQuestionHandler() *QuestionHandler {
	return &QuestionHandler{}
}

func isValidDifficulty(difficulty int) bool {
	return difficulty >= 1 && difficulty <= 5
}

func isValidQuestionType(qType string) bool {
	switch models.QuestionType(qType) {
	case models.QuestionTypeSingleChoice,
		models.QuestionTypeMultipleChoice,
		models.QuestionTypeTrueFalse,
		models.QuestionTypeFillBlank,
		models.QuestionTypeEssay:
		return true
	}
	return false
}

func (h *QuestionHandler) List(c *gin.Context) {
	subject := c.Query("subject")
	knowledgePointID := c.Query("knowledge_point_id")
	difficulty := c.Query("difficulty")
	status := c.Query("status")
	qType := c.Query("type")

	var questions []models.Question
	query := database.DB.Preload("KnowledgePoint")

	if subject != "" {
		query = query.Joins("JOIN knowledge_points ON knowledge_points.id = questions.knowledge_point_id").Where("knowledge_points.subject = ?", subject)
	}
	if knowledgePointID != "" {
		query = query.Where("questions.knowledge_point_id = ?", knowledgePointID)
	}
	if difficulty != "" {
		diff, err := strconv.Atoi(difficulty)
		if err == nil {
			query = query.Where("questions.difficulty = ?", diff)
		}
	}
	if status != "" {
		query = query.Where("questions.status = ?", status)
	}
	if qType != "" {
		query = query.Where("questions.type = ?", qType)
	}

	if err := query.Order("questions.id DESC").Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": questions})
}

func (h *QuestionHandler) Create(c *gin.Context) {
	var req struct {
		QuestionNumber   string               `json:"question_number" binding:"required"`
		Content        string               `json:"content" binding:"required"`
		Type           string               `json:"type" binding:"required"`
		KnowledgePointID uint             `json:"knowledge_point_id" binding:"required"`
		Difficulty     int                  `json:"difficulty" binding:"required"`
		CorrectAnswer  string              `json:"correct_answer" binding:"required"`
		Options        string              `json:"options"`
		Explanation    string              `json:"explanation"`
		Score          float64             `json:"score"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !isValidDifficulty(req.Difficulty) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "difficulty must be between 1 and 5"})
		return
	}

	if !isValidQuestionType(req.Type) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question type"})
		return
	}

	var kp models.KnowledgePoint
	if err := database.DB.First(&kp, req.KnowledgePointID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge point not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var childrenCount int64
	database.DB.Model(&models.KnowledgePoint{}).Where("parent_id = ?", kp.ID).Count(&childrenCount)
	if childrenCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "can only associate questions with leaf knowledge points only"})
		return
	}

	var existing models.Question
	if err := database.DB.Where("question_number = ?", req.QuestionNumber).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "question number already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	score := req.Score
	if score <= 0 {
		score = 1
	}

	question := models.Question{
		QuestionNumber:   req.QuestionNumber,
		Content:        req.Content,
		Type:           models.QuestionType(req.Type),
		KnowledgePointID: req.KnowledgePointID,
		Difficulty:     req.Difficulty,
		CorrectAnswer:  req.CorrectAnswer,
		Options:        req.Options,
		Explanation:    req.Explanation,
		Score:          score,
		Status:         models.QuestionStatusDraft,
	}

	if err := database.DB.Create(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": question})
}

func (h *QuestionHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var question models.Question
	if err := database.DB.Preload("KnowledgePoint").First(&question, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": question})
}

func (h *QuestionHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		QuestionNumber   string               `json:"question_number"`
		Content        string               `json:"content"`
		Type           string               `json:"type"`
		KnowledgePointID *uint             `json:"knowledge_point_id"`
		Difficulty     *int                  `json:"difficulty"`
		CorrectAnswer  string              `json:"correct_answer"`
		Options        string              `json:"options"`
		Explanation    string              `json:"explanation"`
		Score          *float64             `json:"score"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var question models.Question
	if err := database.DB.First(&question, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Difficulty != nil && !isValidDifficulty(*req.Difficulty) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "difficulty must be between 1 and 5"})
		return
	}

	if req.Type != "" && !isValidQuestionType(req.Type) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question type"})
		return
	}

	if req.KnowledgePointID != nil {
		var kp models.KnowledgePoint
		if err := database.DB.First(&kp, *req.KnowledgePointID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "knowledge point not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		question.KnowledgePointID = *req.KnowledgePointID
	}

	if req.QuestionNumber != "" && req.QuestionNumber != question.QuestionNumber {
		var existing models.Question
		if err := database.DB.Where("question_number = ? AND id != ?", req.QuestionNumber, id).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "question number already exists"})
			return
		}
	}

	if req.QuestionNumber != "" {
		question.QuestionNumber = req.QuestionNumber
	}
	if req.Content != "" {
		question.Content = req.Content
	}
	if req.Type != "" {
		question.Type = models.QuestionType(req.Type)
	}
	if req.Difficulty != nil {
		question.Difficulty = *req.Difficulty
	}
	if req.CorrectAnswer != "" {
		question.CorrectAnswer = req.CorrectAnswer
	}
	if req.Options != "" {
		question.Options = req.Options
	}
	if req.Explanation != "" {
		question.Explanation = req.Explanation
	}
	if req.Score != nil {
		question.Score = *req.Score
	}

	if err := database.DB.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": question})
}

func (h *QuestionHandler) SubmitForReview(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var question models.Question
	if err := database.DB.First(&question, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if question.Status != models.QuestionStatusDraft && question.Status != models.QuestionStatusRejected {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only draft or rejected questions can be submitted for review"})
		return
	}

	question.Status = models.QuestionStatusPendingReview
	if err := database.DB.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": question})
}

func (h *QuestionHandler) Approve(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var question models.Question
	if err := database.DB.First(&question, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if question.Status != models.QuestionStatusPendingReview {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only pending review questions can be approved"})
		return
	}

	question.Status = models.QuestionStatusApproved
	if err := database.DB.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": question})
}

func (h *QuestionHandler) Publish(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var question models.Question
	if err := database.DB.First(&question, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if question.Status != models.QuestionStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only approved questions can be published"})
		return
	}

	question.Status = models.QuestionStatusPublished
	if err := database.DB.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": question})
}

func (h *QuestionHandler) Offline(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var question models.Question
	if err := database.DB.First(&question, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if question.Status != models.QuestionStatusPublished {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only published questions can be taken offline"})
		return
	}

	question.Status = models.QuestionStatusOffline
	if err := database.DB.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": question})
}

func (h *QuestionHandler) Reject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var question models.Question
	if err := database.DB.First(&question, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "question not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if question.Status != models.QuestionStatusPendingReview {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only pending review questions can be rejected"})
		return
	}

	question.Status = models.QuestionStatusRejected
	if err := database.DB.Save(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": question})
}

func (h *QuestionHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := database.DB.Delete(&models.Question{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted successfully"})
}
