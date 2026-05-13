package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"pool-health/internal/pool"
)

type Handler struct {
	manager *pool.Manager
}

func NewHandler(manager *pool.Manager) *Handler {
	return &Handler{manager: manager}
}

type RegisterPoolRequest struct {
	Name              string              `json:"name" binding:"required"`
	MaxConns          int               `json:"maxConns"`
	MinIdle           int               `json:"minIdle"`
	GetTimeout        int               `json:"getTimeout"`
	IdleTimeout       int               `json:"idleTimeout"`
	LeakThreshold     int               `json:"leakThreshold"`
	HealthCheckType   string            `json:"healthCheckType"`
	HealthCheckAddr   string            `json:"healthCheckAddr"`
	HealthCheckCmd    string            `json:"healthCheckCmd"`
	HealthCheckInterval int             `json:"healthCheckInterval"`
}

type UpdatePoolConfigRequest struct {
	MaxConns          *int              `json:"maxConns,omitempty"`
	MinIdle           *int              `json:"minIdle,omitempty"`
	GetTimeout        *int              `json:"getTimeout,omitempty"`
	IdleTimeout       *int              `json:"idleTimeout,omitempty"`
	LeakThreshold     *int              `json:"leakThreshold,omitempty"`
	HealthCheckType   *string           `json:"healthCheckType,omitempty"`
	HealthCheckAddr   *string           `json:"healthCheckAddr,omitempty"`
	HealthCheckCmd    *string            `json:"healthCheckCmd,omitempty"`
	HealthCheckInterval *int            `json:"healthCheckInterval,omitempty"`
}

func (h *Handler) RegisterPool(c *gin.Context) {
	var req RegisterPoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := pool.DefaultPoolConfig()
	config.HealthCheckAddr = req.HealthCheckAddr
	config.HealthCheckCmd = req.HealthCheckCmd

	if req.MaxConns > 0 {
		config.MaxConns = req.MaxConns
	}
	if req.MinIdle > 0 {
		config.MinIdle = req.MinIdle
	}
	if req.GetTimeout > 0 {
		config.GetTimeout = time.Duration(req.GetTimeout) * time.Second
	}
	if req.IdleTimeout > 0 {
		config.IdleTimeout = time.Duration(req.IdleTimeout) * time.Second
	}
	if req.LeakThreshold > 0 {
		config.LeakThreshold = time.Duration(req.LeakThreshold) * time.Second
	}
	if req.HealthCheckType != "" {
		config.HealthCheckType = pool.HealthCheckType(req.HealthCheckType)
	}
	if req.HealthCheckInterval > 0 {
		config.HealthCheckInterval = time.Duration(req.HealthCheckInterval) * time.Second
	}

	if err := h.manager.Register(pool.RegisterRequest{
		Name:   req.Name,
		Config: config,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pool registered successfully", "name": req.Name})
}

func (h *Handler) UnregisterPool(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pool name is required"})
		return
	}

	if err := h.manager.Unregister(name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "pool unregistered successfully", "name": name})
}

func (h *Handler) ListPools(c *gin.Context) {
	pools := h.manager.List()
	names := make([]string, 0, len(pools))
	for _, p := range pools {
		names = append(names, p.Name())
	}
	c.JSON(http.StatusOK, gin.H{"pools": names})
}

func (h *Handler) GetPoolStats(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pool name is required"})
		return
	}

	p, err := h.manager.Get(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	stats := p.GetStats()
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetAllStats(c *gin.Context) {
	stats := h.manager.GetAllStats()
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) UpdatePoolConfig(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pool name is required"})
		return
	}

	p, err := h.manager.Get(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	var req UpdatePoolConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := p.GetConfig()

	if req.MaxConns != nil && *req.MaxConns > 0 {
		config.MaxConns = *req.MaxConns
	}
	if req.MinIdle != nil && *req.MinIdle > 0 {
		config.MinIdle = *req.MinIdle
	}
	if req.GetTimeout != nil && *req.GetTimeout > 0 {
		config.GetTimeout = time.Duration(*req.GetTimeout) * time.Second
	}
	if req.IdleTimeout != nil && *req.IdleTimeout > 0 {
		config.IdleTimeout = time.Duration(*req.IdleTimeout) * time.Second
	}
	if req.LeakThreshold != nil && *req.LeakThreshold > 0 {
		config.LeakThreshold = time.Duration(*req.LeakThreshold) * time.Second
	}
	if req.HealthCheckType != nil && *req.HealthCheckType != "" {
		config.HealthCheckType = pool.HealthCheckType(*req.HealthCheckType)
	}
	if req.HealthCheckAddr != nil {
		config.HealthCheckAddr = *req.HealthCheckAddr
	}
	if req.HealthCheckCmd != nil {
		config.HealthCheckCmd = *req.HealthCheckCmd
	}
	if req.HealthCheckInterval != nil && *req.HealthCheckInterval > 0 {
		config.HealthCheckInterval = time.Duration(*req.HealthCheckInterval) * time.Second
	}

	p.UpdateConfig(config)
	c.JSON(http.StatusOK, gin.H{"message": "config updated successfully", "config": config})
}

func (h *Handler) GetPoolConfig(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pool name is required"})
		return
	}

	p, err := h.manager.Get(name)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	config := p.GetConfig()
	c.JSON(http.StatusOK, config)
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
