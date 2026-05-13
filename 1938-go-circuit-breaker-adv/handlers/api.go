package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"circuit-breaker/circuit"
	"circuit-breaker/models"
	"circuit-breaker/subscribers"
)

type Server struct {
	breakerManager   *circuit.BreakerManager
	subscriberManager *subscribers.Manager
}

func NewServer(bm *circuit.BreakerManager, sm *subscribers.Manager) *Server {
	return &Server{
		breakerManager:   bm,
		subscriberManager: sm,
	}
}

type CreateBreakerRequest struct {
	Name            string  `json:"name" binding:"required"`
	WindowSize      string  `json:"window_size,omitempty"`
	FailureThreshold float64 `json:"failure_threshold,omitempty"`
	SlowThreshold   string  `json:"slow_threshold,omitempty"`
	SlowRatio       float64 `json:"slow_ratio,omitempty"`
	CoolDownPeriod  string  `json:"cooldown_period,omitempty"`
	ProbeCount      int     `json:"probe_count,omitempty"`
	MinRequests     int     `json:"min_requests,omitempty"`
}

type UpdateConfigRequest struct {
	WindowSize       string  `json:"window_size,omitempty"`
	FailureThreshold float64 `json:"failure_threshold,omitempty"`
	SlowThreshold    string  `json:"slow_threshold,omitempty"`
	SlowRatio        float64 `json:"slow_ratio,omitempty"`
	CoolDownPeriod   string  `json:"cooldown_period,omitempty"`
	ProbeCount       int     `json:"probe_count,omitempty"`
	MinRequests      int     `json:"min_requests,omitempty"`
}

type SubscribeRequest struct {
	URL string `json:"url" binding:"required"`
}

type BreakerStatusResponse struct {
	Name             string              `json:"name"`
	State            models.State        `json:"state"`
	Config           ConfigResponse      `json:"config"`
	Metrics          MetricsResponse     `json:"metrics"`
	LastTriggeredAt  *string             `json:"last_triggered_at,omitempty"`
	ProbeSuccess     int                 `json:"probe_success,omitempty"`
}

type ConfigResponse struct {
	WindowSize       string  `json:"window_size"`
	FailureThreshold float64 `json:"failure_threshold"`
	SlowThreshold    string  `json:"slow_threshold"`
	SlowRatio        float64 `json:"slow_ratio"`
	CoolDownPeriod   string  `json:"cooldown_period"`
	ProbeCount       int     `json:"probe_count"`
	MinRequests      int     `json:"min_requests"`
}

type MetricsResponse struct {
	Total         int64   `json:"total"`
	Failures      int64   `json:"failures"`
	FailureRate   float64 `json:"failure_rate"`
	SlowCalls     int64   `json:"slow_calls"`
	SlowCallRate  float64 `json:"slow_call_rate"`
}

func parseDuration(s string, defaultVal time.Duration) time.Duration {
	if s == "" {
		return defaultVal
	}
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	if n, err := strconv.Atoi(s); err == nil {
		return time.Duration(n) * time.Second
	}
	return defaultVal
}

func (s *Server) parseConfig(req CreateBreakerRequest) models.BreakerConfig {
	config := models.DefaultConfig()
	
	config.WindowSize = parseDuration(req.WindowSize, config.WindowSize)
	if req.FailureThreshold > 0 {
		config.FailureThreshold = req.FailureThreshold
	}
	config.SlowThreshold = parseDuration(req.SlowThreshold, config.SlowThreshold)
	if req.SlowRatio > 0 {
		config.SlowRatio = req.SlowRatio
	}
	config.CoolDownPeriod = parseDuration(req.CoolDownPeriod, config.CoolDownPeriod)
	if req.ProbeCount > 0 {
		config.ProbeCount = req.ProbeCount
	}
	if req.MinRequests > 0 {
		config.MinRequests = req.MinRequests
	}
	
	return config
}

func (s *Server) parseUpdateConfig(req UpdateConfigRequest, current models.BreakerConfig) models.BreakerConfig {
	config := current
	
	if req.WindowSize != "" {
		config.WindowSize = parseDuration(req.WindowSize, config.WindowSize)
	}
	if req.FailureThreshold > 0 {
		config.FailureThreshold = req.FailureThreshold
	}
	if req.SlowThreshold != "" {
		config.SlowThreshold = parseDuration(req.SlowThreshold, config.SlowThreshold)
	}
	if req.SlowRatio > 0 {
		config.SlowRatio = req.SlowRatio
	}
	if req.CoolDownPeriod != "" {
		config.CoolDownPeriod = parseDuration(req.CoolDownPeriod, config.CoolDownPeriod)
	}
	if req.ProbeCount > 0 {
		config.ProbeCount = req.ProbeCount
	}
	if req.MinRequests > 0 {
		config.MinRequests = req.MinRequests
	}
	
	return config
}

func toBreakerStatusResponse(status models.BreakerState) BreakerStatusResponse {
	var lastTriggeredAt *string
	if status.LastTriggeredAt != nil {
		t := status.LastTriggeredAt.Format(time.RFC3339)
		lastTriggeredAt = &t
	}
	
	var failureRate, slowCallRate float64
	if status.Metrics.Total > 0 {
		failureRate = float64(status.Metrics.Failures) / float64(status.Metrics.Total)
		slowCallRate = float64(status.Metrics.SlowCalls) / float64(status.Metrics.Total)
	}
	
	return BreakerStatusResponse{
		Name:  status.Name,
		State: status.State,
		Config: ConfigResponse{
			WindowSize:       status.Config.WindowSize.String(),
			FailureThreshold: status.Config.FailureThreshold,
			SlowThreshold:    status.Config.SlowThreshold.String(),
			SlowRatio:        status.Config.SlowRatio,
			CoolDownPeriod:   status.Config.CoolDownPeriod.String(),
			ProbeCount:       status.Config.ProbeCount,
			MinRequests:      status.Config.MinRequests,
		},
		Metrics: MetricsResponse{
			Total:        status.Metrics.Total,
			Failures:     status.Metrics.Failures,
			FailureRate:  failureRate,
			SlowCalls:    status.Metrics.SlowCalls,
			SlowCallRate: slowCallRate,
		},
		LastTriggeredAt: lastTriggeredAt,
		ProbeSuccess:    status.ProbeSuccess,
	}
}

func (s *Server) RegisterRoutes(r *gin.Engine) {
	r.GET("/breakers", s.ListBreakers)
	r.POST("/breakers", s.CreateBreaker)
	r.GET("/breakers/:name", s.GetBreaker)
	r.PUT("/breakers/:name/config", s.UpdateConfig)
	r.POST("/breakers/:name/subscribe", s.Subscribe)
	r.DELETE("/breakers/:name/subscribe/:id", s.Unsubscribe)
	r.GET("/breakers/:name/subscribers", s.ListSubscribers)
}

func (s *Server) ListBreakers(c *gin.Context) {
	breakers := s.breakerManager.List()
	
	response := make([]BreakerStatusResponse, 0, len(breakers))
	for _, b := range breakers {
		response = append(response, toBreakerStatusResponse(b.GetStatus()))
	}
	
	c.JSON(http.StatusOK, response)
}

func (s *Server) CreateBreaker(c *gin.Context) {
	var req CreateBreakerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	
	config := s.parseConfig(req)
	
	breaker, err := s.breakerManager.Create(req.Name, config)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	
	breaker.AddEventListener(func(event models.Event) {
		s.subscriberManager.Broadcast(event.BreakerName, event)
	})
	
	c.JSON(http.StatusCreated, toBreakerStatusResponse(breaker.GetStatus()))
}

func (s *Server) GetBreaker(c *gin.Context) {
	name := c.Param("name")
	
	breaker, ok := s.breakerManager.Get(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "breaker not found"})
		return
	}
	
	c.JSON(http.StatusOK, toBreakerStatusResponse(breaker.GetStatus()))
}

func (s *Server) UpdateConfig(c *gin.Context) {
	name := c.Param("name")
	
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	breaker, ok := s.breakerManager.Get(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "breaker not found"})
		return
	}
	
	currentConfig := breaker.Config()
	newConfig := s.parseUpdateConfig(req, currentConfig)
	
	if err := s.breakerManager.UpdateConfig(name, newConfig); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "config updated"})
}

func (s *Server) Subscribe(c *gin.Context) {
	name := c.Param("name")
	
	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if _, ok := s.breakerManager.Get(name); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "breaker not found"})
		return
	}
	
	subscriber := models.Subscriber{
		ID:      uuid.New().String(),
		URL:     req.URL,
		Created: time.Now(),
	}
	
	s.subscriberManager.Subscribe(name, subscriber)
	
	c.JSON(http.StatusCreated, gin.H{
		"id":  subscriber.ID,
		"url": subscriber.URL,
	})
}

func (s *Server) Unsubscribe(c *gin.Context) {
	name := c.Param("name")
	subID := c.Param("id")
	
	s.subscriberManager.Unsubscribe(name, subID)
	
	c.JSON(http.StatusOK, gin.H{"message": "unsubscribed"})
}

func (s *Server) ListSubscribers(c *gin.Context) {
	name := c.Param("name")
	
	subs := s.subscriberManager.List(name)
	
	response := make([]gin.H, 0, len(subs))
	for _, s := range subs {
		response = append(response, gin.H{
			"id":      s.ID,
			"url":     s.URL,
			"created": s.Created.Format(time.RFC3339),
		})
	}
	
	c.JSON(http.StatusOK, response)
}
