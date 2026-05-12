package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"health-checker/pkg/manager"
	"health-checker/pkg/models"
)

type Handler struct {
	manager *manager.Manager
}

func New(m *manager.Manager) *Handler {
	return &Handler{manager: m}
}

func (h *Handler) RegisterComponent(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.manager.Register(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "component registered", "name": req.Name})
}

func (h *Handler) DeregisterComponent(c *gin.Context) {
	name := c.Param("name")

	if !h.manager.Deregister(name) {
		c.JSON(http.StatusNotFound, gin.H{"error": "component not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "component deregistered", "name": name})
}

func (h *Handler) GetComponent(c *gin.Context) {
	name := c.Param("name")

	component, ok := h.manager.Get(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "component not found"})
		return
	}

	displayStatus := component.Status
	if displayStatus == models.StatusUnknown {
		displayStatus = models.StatusUnhealthy
	}

	detail := &models.ComponentDetail{
		Name:          component.Name,
		CheckType:     component.CheckType,
		Address:       component.Address,
		Timeout:       component.Timeout,
		Interval:      component.Interval,
		Status:        displayStatus,
		RegisterTime:  component.RegisterTime,
		LastCheckTime: component.LastCheckTime,
		CheckResults:  component.CheckResults,
	}

	c.JSON(http.StatusOK, detail)
}

func (h *Handler) GetHealth(c *gin.Context) {
	overallStatus := h.manager.GetOverallStatus()
	components := h.manager.GetAll()

	componentDetails := make([]*models.ComponentDetail, 0, len(components))
	for _, comp := range components {
		displayStatus := comp.Status
		if displayStatus == models.StatusUnknown {
			displayStatus = models.StatusUnhealthy
		}

		componentDetails = append(componentDetails, &models.ComponentDetail{
			Name:          comp.Name,
			CheckType:     comp.CheckType,
			Address:       comp.Address,
			Timeout:       comp.Timeout,
			Interval:      comp.Interval,
			Status:        displayStatus,
			RegisterTime:  comp.RegisterTime,
			LastCheckTime: comp.LastCheckTime,
			CheckResults:  comp.CheckResults,
		})
	}

	summary := models.HealthSummary{
		Status:     overallStatus,
		Components: componentDetails,
	}

	if overallStatus == models.StatusUnhealthy {
		c.JSON(http.StatusServiceUnavailable, summary)
		return
	}

	c.JSON(http.StatusOK, summary)
}
