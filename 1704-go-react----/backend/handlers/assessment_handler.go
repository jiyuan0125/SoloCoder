package handlers

import (
	"net/http"
	"rehab-system/models"
	"rehab-system/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type AssessmentHandler struct {
	service *services.AssessmentService
}

func NewAssessmentHandler() *AssessmentHandler {
	return &AssessmentHandler{service: services.NewAssessmentService()}
}

func (h *AssessmentHandler) List(c *gin.Context) {
	patientID, _ := strconv.ParseUint(c.Query("patientId"), 10, 32)
	planID, _ := strconv.ParseUint(c.Query("planId"), 10, 32)

	assessments, err := h.service.List(uint(patientID), uint(planID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assessments)
}

func (h *AssessmentHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	assessment, err := h.service.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "评估不存在"})
		return
	}
	c.JSON(http.StatusOK, assessment)
}

func (h *AssessmentHandler) Record(c *gin.Context) {
	var assessment models.Assessment
	if err := c.ShouldBindJSON(&assessment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if assessment.AssessmentDate.IsZero() {
		assessment.AssessmentDate = time.Now()
	}

	if err := h.service.Record(&assessment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "评估记录成功"})
}
