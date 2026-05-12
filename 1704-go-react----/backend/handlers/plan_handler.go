package handlers

import (
	"net/http"
	"rehab-system/config"
	"rehab-system/models"
	"rehab-system/services"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type PlanHandler struct {
	service *services.PlanService
}

func NewPlanHandler() *PlanHandler {
	return &PlanHandler{service: services.NewPlanService()}
}

func (h *PlanHandler) List(c *gin.Context) {
	patientID, _ := strconv.ParseUint(c.Query("patientId"), 10, 32)
	plans, err := h.service.List(uint(patientID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, plans)
}

func (h *PlanHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	plan, err := h.service.Get(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "计划不存在"})
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *PlanHandler) Create(c *gin.Context) {
	var plan models.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if plan.StartDate.IsZero() {
		plan.StartDate = time.Now()
	}

	if err := h.service.Create(&plan); err != nil {
		if err.Error() == "训练任务已存在" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, plan)
}

func (h *PlanHandler) GetPlanTasks(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var plan models.Plan
	if err := config.DB.First(&plan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "计划不存在"})
		return
	}

	tasks, err := services.NewTrainingService().GetPatientTasks(plan.PatientID, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}
