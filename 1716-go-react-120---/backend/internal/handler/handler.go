package handler

import (
	"net/http"
	"sort"
	"strings"

	"ambulance-scheduler/internal/models"
	"ambulance-scheduler/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	callService     *service.CallService
	vehicleService  *service.VehicleService
	dispatchService *service.DispatchService
	triageService   *service.TriageService
	statsService    *service.StatsService
}

func New(
	callService *service.CallService,
	vehicleService *service.VehicleService,
	dispatchService *service.DispatchService,
	triageService *service.TriageService,
	statsService *service.StatsService,
) *Handler {
	return &Handler{
		callService:     callService,
		vehicleService:  vehicleService,
		dispatchService: dispatchService,
		triageService:   triageService,
		statsService:    statsService,
	}
}

func (h *Handler) ReceiveCall(c *gin.Context) {
	var req service.CreateCallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	call, err := h.callService.CreateCall(&req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "主诉症状") || strings.Contains(errMsg, "病情严重程度") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusCreated, call)
}

func (h *Handler) ListCalls(c *gin.Context) {
	calls := h.callService.ListCalls()
	c.JSON(http.StatusOK, calls)
}

func (h *Handler) GetPendingDispatchCalls(c *gin.Context) {
	calls := h.callService.ListCalls()
	
	pending := make([]*models.EmergencyCall, 0)
	for _, call := range calls {
		if call.Status == models.CallStatusPendingAccept || 
		   call.Status == models.CallStatusAccepted ||
		   call.Status == models.CallStatusProcessing {
			if call.AssignedVehicleID == "" {
				pending = append(pending, call)
			}
		}
	}

	severityOrder := map[models.SeverityLevel]int{
		models.SeverityLevel1: 0,
		models.SeverityLevel2: 1,
		models.SeverityLevel3: 2,
		models.SeverityLevel4: 3,
	}

	sort.Slice(pending, func(i, j int) bool {
		return severityOrder[pending[i].SeverityLevel] < severityOrder[pending[j].SeverityLevel]
	})

	c.JSON(http.StatusOK, pending)
}

func (h *Handler) AcceptCall(c *gin.Context) {
	id := c.Param("id")
	call, err := h.callService.AcceptCall(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, call)
}

func (h *Handler) ProcessCall(c *gin.Context) {
	id := c.Param("id")
	call, err := h.callService.ProcessCall(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, call)
}

func (h *Handler) SubmitForReview(c *gin.Context) {
	id := c.Param("id")
	call, err := h.callService.SubmitForReview(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, call)
}

func (h *Handler) CompleteCall(c *gin.Context) {
	id := c.Param("id")
	call, err := h.callService.CompleteCall(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, call)
}

func (h *Handler) RejectCall(c *gin.Context) {
	id := c.Param("id")
	call, err := h.callService.RejectCall(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, call)
}

func (h *Handler) CreateVehicle(c *gin.Context) {
	var req service.CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vehicle, err := h.vehicleService.CreateVehicle(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, vehicle)
}

func (h *Handler) ListVehicles(c *gin.Context) {
	vehicles := h.vehicleService.ListVehicles()
	c.JSON(http.StatusOK, vehicles)
}

func (h *Handler) UpdateVehicleStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status models.VehicleStatus `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.vehicleService.UpdateStatus(id, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "状态更新成功"})
}

func (h *Handler) SetVehicleMaintenance(c *gin.Context) {
	id := c.Param("id")
	if err := h.vehicleService.SetMaintenance(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "车辆已进入维护状态"})
}

func (h *Handler) RecommendVehicle(c *gin.Context) {
	callID := c.Param("callId")
	recommended, err := h.dispatchService.RecommendVehicle(callID)
	if err != nil {
		if err.Error() == "无合适车辆" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, recommended)
}

func (h *Handler) DispatchVehicle(c *gin.Context) {
	var req service.DispatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := h.dispatchService.DispatchVehicle(&req)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "无合适车辆") {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": errMsg})
			return
		}
		if strings.Contains(errMsg, "维护中") || strings.Contains(errMsg, "已达最大承载") {
			c.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": errMsg})
		return
	}

	c.JSON(http.StatusCreated, record)
}

func (h *Handler) ListDispatchRecords(c *gin.Context) {
	records := h.dispatchService.ListDispatchRecords()
	c.JSON(http.StatusOK, records)
}

func (h *Handler) CreateTriageRecord(c *gin.Context) {
	var req service.CreateTriageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	triage, err := h.triageService.CreateTriageRecord(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, triage)
}

func (h *Handler) ListTriageRecords(c *gin.Context) {
	records := h.triageService.ListTriageRecords()
	c.JSON(http.StatusOK, records)
}

func (h *Handler) GetStatistics(c *gin.Context) {
	stats := h.statsService.GetStatistics()
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
