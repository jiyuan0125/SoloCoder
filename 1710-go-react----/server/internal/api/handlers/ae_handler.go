package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"trial-management-system/internal/models"
	"trial-management-system/internal/services"
)

type CreateAERequest struct {
	SubjectID    string             `json:"subject_id" binding:"required"`
	EventName    string             `json:"event_name" binding:"required"`
	StartDate    time.Time          `json:"start_date" binding:"required"`
	EndDate      *time.Time         `json:"end_date"`
	Severity     string             `json:"severity" binding:"required"`
	Relationship string             `json:"relationship" binding:"required"`
	IsSAE        bool               `json:"is_sae"`
	Treatment    string             `json:"treatment"`
	Outcome      string             `json:"outcome"`
}

func CreateAdverseEvent(c *gin.Context) {
	var req CreateAERequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subjectID, err := uuid.Parse(req.SubjectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的受试者ID"})
		return
	}

	if !services.SubjectExists(subjectID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "受试者不存在"})
		return
	}

	ae := &models.AdverseEvent{
		SubjectID:    subjectID,
		EventName:    req.EventName,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Severity:     models.AESeverity(req.Severity),
		Relationship: models.AERelationship(req.Relationship),
		IsSAE:        req.IsSAE,
		Treatment:    req.Treatment,
		Outcome:      models.AEOutcome(req.Outcome),
	}

	if err := services.CreateAdverseEvent(ae); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if req.IsSAE {
		deadline := time.Now().Add(24 * time.Hour)
		todo := &models.Todo{
			Title:       "SAE报告",
			Description: "需要提交严重不良事件报告: " + req.EventName,
			AeID:        ae.ID,
			Status:      models.TodoPending,
			DueDate:     &deadline,
		}
		services.CreateTodo(todo)
	}

	c.JSON(http.StatusCreated, ae)
}

func ListAdverseEvents(c *gin.Context) {
	subjectIDStr := c.Query("subject_id")
	protocolIDStr := c.Query("protocol_id")
	isSAE := c.Query("is_sae") == "true"

	var aes []models.AdverseEvent
	db := services.GetDB()

	if subjectIDStr != "" {
		subjectID, err := uuid.Parse(subjectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的受试者ID"})
			return
		}
		db = db.Where("subject_id = ?", subjectID)
	}

	if protocolIDStr != "" {
		protocolID, err := uuid.Parse(protocolIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的方案ID"})
			return
		}
		db = db.Joins("JOIN subjects ON subjects.id = adverse_events.subject_id").Where("subjects.protocol_id = ?", protocolID)
	}

	if isSAE {
		db = db.Where("is_sae = ?", true)
	}

	if err := db.Find(&aes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, aes)
}

func GetAdverseEvent(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	ae, err := services.GetAdverseEventByID(uuidID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "不良事件不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ae)
}

func UpdateAdverseEvent(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	_, err = services.GetAdverseEventByID(uuidID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "不良事件不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req CreateAERequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.EventName != "" {
		updates["event_name"] = req.EventName
	}
	if !req.StartDate.IsZero() {
		updates["start_date"] = req.StartDate
	}
	if req.EndDate != nil {
		updates["end_date"] = req.EndDate
	}
	if req.Severity != "" {
		updates["severity"] = req.Severity
	}
	if req.Relationship != "" {
		updates["relationship"] = req.Relationship
	}
	if req.Treatment != "" {
		updates["treatment"] = req.Treatment
	}
	if req.Outcome != "" {
		updates["outcome"] = req.Outcome
	}

	if len(updates) > 0 {
		if err := services.UpdateAdverseEvent(uuidID, updates); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	updated, _ := services.GetAdverseEventByID(uuidID)
	c.JSON(http.StatusOK, updated)
}

func SubmitSAEReport(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	ae, err := services.GetAdverseEventByID(uuidID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "不良事件不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !ae.IsSAE {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不是严重不良事件"})
		return
	}

	if err := services.UpdateAdverseEvent(uuidID, map[string]interface{}{"report_submitted": true}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var todos []models.Todo
	services.GetDB().Where("ae_id = ?", uuidID).Find(&todos)
	for _, todo := range todos {
		services.UpdateTodo(todo.ID, map[string]interface{}{"status": models.TodoCompleted})
	}

	c.JSON(http.StatusOK, gin.H{"message": "SAE报告已提交"})
}

func ListTodos(c *gin.Context) {
	services.CheckOverdueTodos()
	todos, err := services.ListTodos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, todos)
}

func UpdateTodoStatus(c *gin.Context) {
	id := c.Param("id")
	uuidID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := services.UpdateTodo(uuidID, map[string]interface{}{"status": req.Status}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "待办状态已更新"})
}
