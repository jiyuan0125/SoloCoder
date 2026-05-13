package api

import (
	"breaker/internal/breaker"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type MetricConfigDTO struct {
	Type         string  `json:"type"`
	Threshold    float64 `json:"threshold"`
	TimeoutLimit float64 `json:"timeout_limit_ms,omitempty"`
	CustomName   string  `json:"custom_name,omitempty"`
}

type UpdateMetricsRequest struct {
	Metrics []MetricConfigDTO `json:"metrics"`
}

type MetricsStatusDTO struct {
	Type       string  `json:"type"`
	Name       string  `json:"name,omitempty"`
	Current    float64 `json:"current"`
	Threshold  float64 `json:"threshold"`
	Exceeded   bool    `json:"exceeded"`
	Configured bool    `json:"configured"`
}

type CircuitBreakerStatusDTO struct {
	Name      string            `json:"name"`
	State     string            `json:"state"`
	Metrics   []MetricsStatusDTO `json:"metrics"`
	Timestamp time.Time         `json:"timestamp"`
}

type ReportCustomMetricRequest struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type Handler struct {
	manager *breaker.Manager
}

func NewHandler(manager *breaker.Manager) *Handler {
	return &Handler{
		manager: manager,
	}
}

func (h *Handler) GetStatus(c *gin.Context) {
	name := c.Param("name")
	cb, ok := h.manager.Get(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "breaker not found"})
		return
	}

	status := cb.Status()
	c.JSON(http.StatusOK, toStatusDTO(status))
}

func (h *Handler) UpdateMetrics(c *gin.Context) {
	name := c.Param("name")

	var req UpdateMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	configs, err := dtoToConfigs(req.Metrics)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cb := h.manager.GetOrCreate(name, nil)
	cb.UpdateMetrics(configs)

	c.JSON(http.StatusOK, gin.H{
		"message": "metrics updated",
		"name":    name,
	})
}

func (h *Handler) ReportCustomMetric(c *gin.Context) {
	name := c.Param("name")

	var req ReportCustomMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "custom metric name required"})
		return
	}

	cb, ok := h.manager.Get(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "breaker not found"})
		return
	}

	cb.RecordCustom(req.Name, req.Value)

	c.JSON(http.StatusOK, gin.H{
		"message": "custom metric reported",
		"breaker": name,
		"metric":  req.Name,
		"value":   req.Value,
	})
}

func (h *Handler) ListBreakers(c *gin.Context) {
	breakers := h.manager.All()
	result := make([]CircuitBreakerStatusDTO, 0, len(breakers))
	for _, cb := range breakers {
		result = append(result, *toStatusDTO(cb.Status()))
	}
	c.JSON(http.StatusOK, result)
}

func toStatusDTO(s *breaker.CircuitBreakerStatus) *CircuitBreakerStatusDTO {
	metrics := make([]MetricsStatusDTO, 0, len(s.Metrics))
	for _, m := range s.Metrics {
		metrics = append(metrics, MetricsStatusDTO{
			Type:       string(m.Type),
			Name:       m.Name,
			Current:    m.Current,
			Threshold:  m.Threshold,
			Exceeded:   m.Exceeded,
			Configured: m.Configured,
		})
	}
	return &CircuitBreakerStatusDTO{
		Name:      s.Name,
		State:     string(s.State),
		Metrics:   metrics,
		Timestamp: s.Timestamp,
	}
}

func dtoToConfigs(dtos []MetricConfigDTO) ([]*breaker.MetricConfig, error) {
	configs := make([]*breaker.MetricConfig, 0, len(dtos))
	for _, dto := range dtos {
		cfg := &breaker.MetricConfig{
			Type:       breaker.MetricType(dto.Type),
			Threshold:  dto.Threshold,
			CustomName: dto.CustomName,
		}
		if dto.TimeoutLimit > 0 {
			cfg.TimeoutLimit = time.Duration(dto.TimeoutLimit) * time.Millisecond
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}
