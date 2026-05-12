package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type ServiceInstance struct {
	ServiceName    string                 `json:"service_name"`
	IP             string                 `json:"ip"`
	Port           int                    `json:"port"`
	Metadata       map[string]string      `json:"metadata"`
	LastHeartbeat  time.Time              `json:"last_heartbeat"`
	UnhealthySince *time.Time             `json:"unhealthy_since,omitempty"`
	MissedCount    int                    `json:"missed_count"`
}

type Registry struct {
	instances map[string]*ServiceInstance
	mu        sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		instances: make(map[string]*ServiceInstance),
	}
}

func makeKey(serviceName, ip string, port int) string {
	return fmt.Sprintf("%s:%s:%d", serviceName, ip, port)
}

type RegisterRequest struct {
	ServiceName string            `json:"service_name"`
	IP          string            `json:"ip"`
	Port        int               `json:"port"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type HeartbeatRequest struct {
	ServiceName string `json:"service_name"`
	IP          string `json:"ip"`
	Port        int    `json:"port"`
}

const (
	heartbeatTimeout    = 15 * time.Second
	maxMissedHeartbeats = 2
	unhealthyTimeout    = 30 * time.Second
	defaultPageSize     = 20
	maxPageSize         = 100
)

func (r *Registry) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var missingFields []string
	if req.ServiceName == "" {
		missingFields = append(missingFields, "service_name")
	}
	if req.IP == "" {
		missingFields = append(missingFields, "ip")
	}
	if req.Port == 0 {
		missingFields = append(missingFields, "port")
	}

	if len(missingFields) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "missing required fields",
			"missing_fields": missingFields,
		})
		return
	}

	key := makeKey(req.ServiceName, req.IP, req.Port)
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	instance, exists := r.instances[key]
	if exists {
		instance.LastHeartbeat = now
		instance.Metadata = req.Metadata
		instance.UnhealthySince = nil
		instance.MissedCount = 0
		c.JSON(http.StatusOK, gin.H{"message": "updated"})
		return
	}

	r.instances[key] = &ServiceInstance{
		ServiceName:   req.ServiceName,
		IP:            req.IP,
		Port:          req.Port,
		Metadata:      req.Metadata,
		LastHeartbeat: now,
		MissedCount:   0,
	}
	c.JSON(http.StatusCreated, gin.H{"message": "registered"})
}

func (r *Registry) Heartbeat(c *gin.Context) {
	var req HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.ServiceName == "" || req.IP == "" || req.Port == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields"})
		return
	}

	key := makeKey(req.ServiceName, req.IP, req.Port)
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.instances[key]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "service not found, please register first"})
		return
	}

	r.instances[key].LastHeartbeat = now
	r.instances[key].UnhealthySince = nil
	r.instances[key].MissedCount = 0

	c.JSON(http.StatusOK, gin.H{"message": "heartbeat received"})
}

func (r *Registry) Discover(c *gin.Context) {
	serviceName := c.Query("service_name")

	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")

	page := 1
	pageSize := defaultPageSize

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			if ps > maxPageSize {
				ps = maxPageSize
			}
			pageSize = ps
		}
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []*ServiceInstance
	for _, inst := range r.instances {
		if inst.UnhealthySince == nil {
			if serviceName == "" || inst.ServiceName == serviceName {
				filtered = append(filtered, inst)
			}
		}
	}

	total := len(filtered)
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	if page > totalPages {
		c.JSON(http.StatusOK, gin.H{
			"services":   []ServiceInstance{},
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": totalPages,
		})
		return
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	result := make([]ServiceInstance, 0, end-start)
	for i := start; i < end; i++ {
		result = append(result, *filtered[i])
	}

	c.JSON(http.StatusOK, gin.H{
		"services":    result,
		"page":        page,
		"page_size":   pageSize,
		"total":       total,
		"total_pages": totalPages,
	})
}

func (r *Registry) healthChecker() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		r.checkHealth()
	}
}

func (r *Registry) checkHealth() {
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	for key, inst := range r.instances {
		if inst.UnhealthySince != nil {
			if now.Sub(*inst.UnhealthySince) >= unhealthyTimeout {
				delete(r.instances, key)
				continue
			}
		}

		timeSinceHeartbeat := now.Sub(inst.LastHeartbeat)
		if timeSinceHeartbeat >= heartbeatTimeout {
			inst.MissedCount++
			if inst.MissedCount >= maxMissedHeartbeats && inst.UnhealthySince == nil {
				inst.UnhealthySince = &now
			}
			inst.LastHeartbeat = now
		}
	}
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8103"
	}
	return port
}

func main() {
	registry := NewRegistry()
	go registry.healthChecker()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.POST("/register", registry.Register)
	router.POST("/heartbeat", registry.Heartbeat)
	router.GET("/discover", registry.Discover)

	port := getPort()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	fmt.Printf("Registry server starting on port %s\n", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}
