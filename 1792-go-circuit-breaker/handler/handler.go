package handler

import (
	"net/http"
	"time"

	"circuit-breaker/circuitbreaker"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	manager *circuitbreaker.Manager
}

func NewHandler(manager *circuitbreaker.Manager) *Handler {
	return &Handler{manager: manager}
}

type ServiceStatusResponse struct {
	ServiceName string                     `json:"service_name"`
	State       string                     `json:"state"`
	Configured  bool                       `json:"configured"`
	Records     []CallRecordResponse       `json:"recent_records"`
}

type CallRecordResponse struct {
	Timestamp string `json:"timestamp"`
	Success   bool   `json:"success"`
	DurationMs int64 `json:"duration_ms"`
}

func (h *Handler) GetStatus(c *gin.Context) {
	services := h.manager.List()
	serviceNames := make(map[string]bool)
	response := make([]ServiceStatusResponse, 0, len(services))

	for _, cb := range services {
		serviceNames[cb.ServiceName()] = true
		response = append(response, ServiceStatusResponse{
			ServiceName: cb.ServiceName(),
			State:       string(cb.State()),
			Configured:  true,
			Records:     convertRecords(cb.CallRecords()),
		})
	}

	queryNames := c.QueryArray("service")
	for _, name := range queryNames {
		if !serviceNames[name] {
			response = append(response, ServiceStatusResponse{
				ServiceName: name,
				State:       "未配置",
				Configured:  false,
				Records:     []CallRecordResponse{},
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"services": response,
	})
}

func (h *Handler) ResetService(c *gin.Context) {
	serviceName := c.Param("service")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	ip := c.ClientIP()
	success := h.manager.Reset(serviceName, ip)

	if !success {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "service reset successfully",
		"service":     serviceName,
		"reset_by":    ip,
		"reset_time":  time.Now().Format(time.RFC3339),
	})
}

func (h *Handler) SimulateCall(c *gin.Context) {
	var req struct {
		Service string `json:"service" binding:"required"`
		Success bool   `json:"success"`
		DelayMs int    `json:"delay_ms"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !h.manager.Allow(req.Service) {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":   "circuit breaker is open",
			"service": req.Service,
		})
		return
	}

	start := time.Now()
	duration := time.Duration(req.DelayMs) * time.Millisecond
	if duration > 0 {
		time.Sleep(duration)
	}

	actualDuration := time.Since(start)
	success := req.Success

	cb := h.manager.Get(req.Service)
	if cb != nil {
		config := cb.Config()
		timeout := config.HalfOpenTimeout
		if timeout <= 0 {
			timeout = 5 * time.Second
		}

		if cb.State() == circuitbreaker.StateHalfOpen && actualDuration > timeout {
			success = false
		}
	}

	h.manager.Record(req.Service, success, actualDuration)

	if success {
		c.JSON(http.StatusOK, gin.H{
			"service":  req.Service,
			"success":  true,
			"duration": actualDuration.Milliseconds(),
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"service":  req.Service,
			"success":  false,
			"duration": actualDuration.Milliseconds(),
		})
	}
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func convertRecords(records []*circuitbreaker.CallRecord) []CallRecordResponse {
	result := make([]CallRecordResponse, 0, len(records))
	for _, r := range records {
		result = append(result, CallRecordResponse{
			Timestamp:  r.Timestamp.Format(time.RFC3339),
			Success:    r.Success,
			DurationMs: r.Duration.Milliseconds(),
		})
	}
	return result
}
