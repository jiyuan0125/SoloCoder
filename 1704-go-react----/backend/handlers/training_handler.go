package handlers

import (
	"net/http"
	"rehab-system/models"
	"rehab-system/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type TrainingHandler struct {
	service *services.TrainingService
}

func NewTrainingHandler() *TrainingHandler {
	return &TrainingHandler{service: services.NewTrainingService()}
}

func (h *TrainingHandler) GetPatientTasks(c *gin.Context) {
	patientID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	dateStr := c.Query("date")

	var date time.Time
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式错误，应为 YYYY-MM-DD"})
			return
		}
		date = parsed
	} else {
		date = time.Now()
	}

	tasks, err := h.service.GetPatientTasks(uint(patientID), date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func (h *TrainingHandler) GetTask(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	task, err := h.service.GetTask(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *TrainingHandler) CreateRecord(c *gin.Context) {
	var record models.TrainingRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if record.CompletedAt.IsZero() {
		record.CompletedAt = time.Now()
	}

	if err := h.service.CreateRecord(&record); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, record)
}

func (h *TrainingHandler) GetDifficultyHistory(c *gin.Context) {
	exerciseID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	logs, err := h.service.GetDifficultyHistory(uint(exerciseID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}
