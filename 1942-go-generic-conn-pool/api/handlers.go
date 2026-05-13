package api

import (
	"net/http"
	"time"

	"poolapp/pool"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	manager  *pool.Manager
	registry *pool.FactoryRegistry
}

func NewHandler(manager *pool.Manager, registry *pool.FactoryRegistry) *Handler {
	return &Handler{
		manager:  manager,
		registry: registry,
	}
}

type CreatePoolRequest struct {
	Name         string        `json:"name" binding:"required"`
	FactoryType  string        `json:"factory_type" binding:"required"`
	MinIdle      int           `json:"min_idle"`
	MaxTotal     int           `json:"max_total"`
	MaxLifetime  time.Duration `json:"max_lifetime"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

type PoolStatsResponse struct {
	Name              string  `json:"name"`
	FactoryType       string  `json:"factory_type"`
	ActiveConnections int     `json:"active_connections"`
	IdleConnections   int     `json:"idle_connections"`
	WaitingRequests   int     `json:"waiting_requests"`
	TotalBorrows      int64   `json:"total_borrows"`
	AvgBorrowTimeMs   float64 `json:"avg_borrow_time_ms"`
	TotalBorrowTimeMs float64 `json:"total_borrow_time_ms"`
	MinIdle           int     `json:"min_idle"`
	MaxTotal          int     `json:"max_total"`
	MaxLifetimeSec    int64   `json:"max_lifetime_sec"`
	IdleTimeoutSec    int64   `json:"idle_timeout_sec"`
	CreatedAt         string  `json:"created_at"`
}

func (h *Handler) GetPools(c *gin.Context) {
	statsList := h.manager.GetAllStats()
	response := make([]PoolStatsResponse, 0, len(statsList))
	for _, stats := range statsList {
		response = append(response, toResponse(stats))
	}
	c.JSON(http.StatusOK, gin.H{
		"pools": response,
		"count": len(response),
	})
}

func (h *Handler) CreatePool(c *gin.Context) {
	var req CreatePoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	factory, exists := h.registry.Get(req.FactoryType)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unknown factory type",
			"available_types": h.registry.List(),
		})
		return
	}

	cfg := pool.PoolConfig{
		Name:        req.Name,
		Factory:     factory,
		FactoryType: req.FactoryType,
		MinIdle:     req.MinIdle,
		MaxTotal:    req.MaxTotal,
		MaxLifetime: req.MaxLifetime,
		IdleTimeout: req.IdleTimeout,
	}

	if err := h.manager.CreatePool(cfg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stats, _ := h.manager.GetStats(req.Name)
	c.JSON(http.StatusCreated, gin.H{
		"message": "pool created successfully",
		"pool":    toResponse(stats),
	})
}

func (h *Handler) GetPoolStats(c *gin.Context) {
	name := c.Param("name")
	stats, exists := h.manager.GetStats(name)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "pool not found"})
		return
	}
	c.JSON(http.StatusOK, toResponse(stats))
}

func (h *Handler) GetFactoryTypes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"factory_types": h.registry.List(),
	})
}

func toResponse(stats pool.PoolStats) PoolStatsResponse {
	return PoolStatsResponse{
		Name:              stats.Name,
		FactoryType:       stats.FactoryType,
		ActiveConnections: stats.ActiveConnections,
		IdleConnections:   stats.IdleConnections,
		WaitingRequests:   stats.WaitingRequests,
		TotalBorrows:      stats.TotalBorrows,
		AvgBorrowTimeMs:   float64(stats.AvgBorrowTime) / float64(time.Millisecond),
		TotalBorrowTimeMs: float64(stats.TotalBorrowTime) / float64(time.Millisecond),
		MinIdle:           stats.MinIdle,
		MaxTotal:          stats.MaxTotal,
		MaxLifetimeSec:    int64(stats.MaxLifetime / time.Second),
		IdleTimeoutSec:    int64(stats.IdleTimeout / time.Second),
		CreatedAt:         stats.CreatedAt.Format(time.RFC3339),
	}
}
