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

const MaxLevel = 4

type KnowledgeHandler struct{}

func NewKnowledgeHandler() *KnowledgeHandler {
	return &KnowledgeHandler{}
}

func (h *KnowledgeHandler) buildTree(knowledgePoints []models.KnowledgePoint) []models.KnowledgePoint {
	kpMap := make(map[uint]*models.KnowledgePoint)
	var roots []models.KnowledgePoint

	for i := range knowledgePoints {
		kpMap[knowledgePoints[i].ID] = &knowledgePoints[i]
	}

	for i := range knowledgePoints {
		kp := &knowledgePoints[i]
		if kp.ParentID == nil {
			roots = append(roots, *kp)
		} else if parent, exists := kpMap[*kp.ParentID]; exists {
			parent.Children = append(parent.Children, *kp)
		}
	}

	return roots
}

func (h *KnowledgeHandler) List(c *gin.Context) {
	subject := c.Query("subject")

	var kps []models.KnowledgePoint
	query := database.DB

	if subject != "" {
		query = query.Where("subject = ?", subject)
	}

	if err := query.Order("level, id").Find(&kps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tree := h.buildTree(kps)
	c.JSON(http.StatusOK, gin.H{"data": tree})
}

func (h *KnowledgeHandler) Create(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Subject  string `json:"subject" binding:"required"`
		ParentID *uint  `json:"parent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.KnowledgePoint
	if err := database.DB.Where("name = ? AND subject = ?", req.Name, req.Subject).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "knowledge point name already exists in this subject"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	level := 1
	if req.ParentID != nil {
		var parent models.KnowledgePoint
		if err := database.DB.First(&parent, *req.ParentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "parent knowledge point not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		level = parent.Level + 1
		if level > MaxLevel {
			c.JSON(http.StatusBadRequest, gin.H{"error": "maximum level exceeded (max 4)"})
			return
		}
	}

	kp := models.KnowledgePoint{
		Name:     req.Name,
		Subject:  req.Subject,
		ParentID: req.ParentID,
		Level:    level,
	}

	if err := database.DB.Create(&kp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": kp})
}

func (h *KnowledgeHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Name    string `json:"name"`
		Subject string `json:"subject"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var kp models.KnowledgePoint
	if err := database.DB.First(&kp, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge point not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" && req.Name != kp.Name {
		subject := kp.Subject
		if req.Subject != "" {
			subject = req.Subject
		}
		var existing models.KnowledgePoint
		if err := database.DB.Where("name = ? AND subject = ? AND id != ?", req.Name, subject, id).First(&existing).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "knowledge point name already exists in this subject"})
			return
		}
	}

	if req.Name != "" {
		kp.Name = req.Name
	}
	if req.Subject != "" {
		kp.Subject = req.Subject
	}

	if err := database.DB.Save(&kp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": kp})
}

func (h *KnowledgeHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var childrenCount int64
	database.DB.Model(&models.KnowledgePoint{}).Where("parent_id = ?", id).Count(&childrenCount)
	if childrenCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete knowledge point with children"})
		return
	}

	var questionCount int64
	database.DB.Model(&models.Question{}).Where("knowledge_point_id = ?", id).Count(&questionCount)
	if questionCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete knowledge point with associated questions"})
		return
	}

	if err := database.DB.Delete(&models.KnowledgePoint{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted successfully"})
}

func (h *KnowledgeHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var kp models.KnowledgePoint
	if err := database.DB.First(&kp, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "knowledge point not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": kp})
}
