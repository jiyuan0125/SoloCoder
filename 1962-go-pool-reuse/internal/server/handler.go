package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"shardpool/internal/pool"
)

type Server struct {
	pool *pool.ShardPool
}

type AcquireRequest struct {
	Key     string        `json:"key" binding:"required"`
	Timeout time.Duration `json:"timeout"`
}

type AcquireResponse struct {
	ConnectionID string    `json:"connection_id"`
	ShardID      int       `json:"shard_id"`
	Key          string    `json:"key"`
	AcquiredAt   time.Time `json:"acquired_at"`
}

type ReleaseRequest struct {
	Key          string `json:"key" binding:"required"`
	ConnectionID string `json:"connection_id" binding:"required"`
}

type ReleaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ShardStatsResponse struct {
	ShardID      int    `json:"shard_id"`
	ActiveConns  int    `json:"active_connections"`
	IdleConns    int    `json:"idle_connections"`
	WaitingCount int64  `json:"waiting_count"`
	TotalRequests int64 `json:"total_requests"`
	AvgLatency   string `json:"average_acquire_latency"`
}

type StatsResponse struct {
	TotalShards        int                 `json:"total_shards"`
	TotalActiveConns   int                 `json:"total_active_connections"`
	TotalIdleConns     int                 `json:"total_idle_connections"`
	TotalRequests      int64               `json:"total_requests"`
	AvgAcquireLatency  string              `json:"average_acquire_latency"`
	ShardDistribution  map[int]float64     `json:"shard_distribution"`
}

func NewServer(p *pool.ShardPool) *Server {
	return &Server{pool: p}
}

func (s *Server) Acquire(c *gin.Context) {
	var req AcquireRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
		return
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, shardID, err := s.pool.Acquire(ctx, req.Key)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":    err.Error(),
			"shard_id": shardID,
		})
		return
	}

	c.JSON(http.StatusOK, AcquireResponse{
		ConnectionID: conn.ID,
		ShardID:      shardID,
		Key:          req.Key,
		AcquiredAt:   time.Now(),
	})
}

func (s *Server) Release(c *gin.Context) {
	var req ReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.pool.Release(req.Key, req.ConnectionID)

	c.JSON(http.StatusOK, ReleaseResponse{
		Success: true,
		Message: "connection released",
	})
}

func (s *Server) ListShards(c *gin.Context) {
	shards := s.pool.AllShardsStats()

	response := make([]ShardStatsResponse, len(shards))
	for i, ss := range shards {
		response[i] = ShardStatsResponse{
			ShardID:        ss.ShardID,
			ActiveConns:    ss.ActiveConns,
			IdleConns:      ss.IdleConns,
			WaitingCount:   ss.WaitingCount,
			TotalRequests:  ss.TotalRequests,
			AvgLatency:     ss.AcquireLatency.String(),
		}
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) Stats(c *gin.Context) {
	stats := s.pool.Stats()

	response := StatsResponse{
		TotalShards:       stats.TotalShards,
		TotalActiveConns:  stats.TotalActiveConns,
		TotalIdleConns:    stats.TotalIdleConns,
		TotalRequests:     stats.TotalRequests,
		AvgAcquireLatency: stats.AvgAcquireLatency.String(),
		ShardDistribution: stats.ShardDistribution,
	}

	c.JSON(http.StatusOK, response)
}
