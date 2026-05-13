package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type BulkheadState string

const (
	StateActive   BulkheadState = "active"
	StateBreaking BulkheadState = "breaking"
	StateRecovery BulkheadState = "recovery"
)

type BulkheadConfig struct {
	BackendAddress   string        `json:"backend_address"`
	MaxConcurrent    int           `json:"max_concurrent"`
	FailureThreshold float64       `json:"failure_threshold"`
	BreakDuration    time.Duration `json:"break_duration"`
	WindowSize       int           `json:"window_size"`
}

type BulkheadStats struct {
	Name              string        `json:"name"`
	State             BulkheadState `json:"state"`
	ActiveConnections int32         `json:"active_connections"`
	TotalRequests     int64         `json:"total_requests"`
	FailedRequests    int64         `json:"failed_requests"`
	FailureRate       float64       `json:"failure_rate"`
}

type Bulkhead struct {
	Name              string
	config            atomic.Value
	backendURL        *url.URL
	proxy             *httputil.ReverseProxy
	state             atomic.Value
	activeConnections int32
	totalRequests     int64
	failedRequests    int64
	mu                sync.RWMutex
	lastFailureTime   time.Time
	lastSuccessTime   time.Time
	recoveryStart     time.Time
}

type Manager struct {
	bulkheads sync.Map
}

var manager = &Manager{}

func NewBulkhead(name string, config BulkheadConfig) (*Bulkhead, error) {
	if name == "" {
		return nil, errors.New("name cannot be empty")
	}
	if config.BackendAddress == "" {
		return nil, errors.New("backend_address cannot be empty")
	}
	if config.MaxConcurrent <= 0 {
		return nil, errors.New("max_concurrent must be greater than 0")
	}
	if config.FailureThreshold <= 0 || config.FailureThreshold > 1 {
		return nil, errors.New("failure_threshold must be between 0 and 1")
	}
	if config.BreakDuration <= 0 {
		return nil, errors.New("break_duration must be greater than 0")
	}
	if config.WindowSize <= 0 {
		config.WindowSize = 20
	}

	backendURL, err := url.Parse(config.BackendAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid backend_address: %w", err)
	}

	b := &Bulkhead{
		Name:       name,
		backendURL: backendURL,
	}
	b.config.Store(config)
	b.state.Store(StateActive)

	b.proxy = httputil.NewSingleHostReverseProxy(backendURL)
	b.proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"error": "backend unavailable", "details": "%s"}`, err.Error())
	}

	return b, nil
}

func (b *Bulkhead) GetConfig() BulkheadConfig {
	return b.config.Load().(BulkheadConfig)
}

func (b *Bulkhead) SetConfig(config BulkheadConfig) {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if config.BackendAddress != "" && config.BackendAddress != b.backendURL.String() {
		backendURL, err := url.Parse(config.BackendAddress)
		if err == nil {
			b.backendURL = backendURL
			b.proxy = httputil.NewSingleHostReverseProxy(backendURL)
			b.proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintf(w, `{"error": "backend unavailable", "details": "%s"}`, err.Error())
			}
		}
	}
	
	oldConfig := b.GetConfig()
	if config.MaxConcurrent <= 0 {
		config.MaxConcurrent = oldConfig.MaxConcurrent
	}
	if config.FailureThreshold <= 0 || config.FailureThreshold > 1 {
		config.FailureThreshold = oldConfig.FailureThreshold
	}
	if config.BreakDuration <= 0 {
		config.BreakDuration = oldConfig.BreakDuration
	}
	if config.WindowSize <= 0 {
		config.WindowSize = oldConfig.WindowSize
	}
	if config.BackendAddress == "" {
		config.BackendAddress = oldConfig.BackendAddress
	}
	
	b.config.Store(config)
}

func (b *Bulkhead) GetState() BulkheadState {
	return b.state.Load().(BulkheadState)
}

func (b *Bulkhead) SetState(state BulkheadState) {
	currentState := b.GetState()
	if currentState != state {
		b.state.Store(state)
		if state == StateBreaking {
			atomic.StoreInt64(&b.failedRequests, 0)
			atomic.StoreInt64(&b.totalRequests, 0)
			b.recoveryStart = time.Now().Add(b.GetConfig().BreakDuration)
		}
	}
}

func (b *Bulkhead) recordSuccess() {
	atomic.AddInt64(&b.totalRequests, 1)
	b.mu.Lock()
	b.lastSuccessTime = time.Now()
	b.mu.Unlock()
	
	if b.GetState() == StateRecovery {
		b.SetState(StateActive)
	}
}

func (b *Bulkhead) recordFailure() {
	total := atomic.AddInt64(&b.totalRequests, 1)
	failed := atomic.AddInt64(&b.failedRequests, 1)
	
	b.mu.Lock()
	b.lastFailureTime = time.Now()
	b.mu.Unlock()
	
	config := b.GetConfig()
	if total >= int64(config.WindowSize) {
		failureRate := float64(failed) / float64(total)
		if failureRate >= config.FailureThreshold {
			b.SetState(StateBreaking)
		}
	}
}

func (b *Bulkhead) tryAcquire() bool {
	config := b.GetConfig()
	current := atomic.LoadInt32(&b.activeConnections)
	if current >= int32(config.MaxConcurrent) {
		return false
	}
	return atomic.CompareAndSwapInt32(&b.activeConnections, current, current+1)
}

func (b *Bulkhead) release() {
	atomic.AddInt32(&b.activeConnections, -1)
}

func (b *Bulkhead) ShouldBreak() bool {
	state := b.GetState()
	if state == StateActive {
		return false
	}
	
	if state == StateBreaking {
		if time.Now().After(b.recoveryStart) {
			b.SetState(StateRecovery)
			return false
		}
		return true
	}
	
	return false
}

func (b *Bulkhead) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	state := b.GetState()
	
	if state == StateBreaking {
		if time.Now().After(b.recoveryStart) {
			b.SetState(StateRecovery)
		} else {
			b.handleFallback(w, r)
			return
		}
	}
	
	if state == StateRecovery {
		if !b.tryAcquire() {
			b.handleFallback(w, r)
			return
		}
		defer b.release()
	} else {
		if !b.tryAcquire() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, `{"error": "bulkhead capacity exceeded", "bulkhead": "%s"}`, b.Name)
			return
		}
		defer b.release()
	}
	
	rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
	
	modifiedPath := "/" + b.Name + r.URL.Path
	originalPath := r.URL.Path
	r.URL.Path = modifiedPath
	
	b.proxy.ServeHTTP(rw, r)
	
	r.URL.Path = originalPath
	
	if rw.status >= 500 {
		b.recordFailure()
	} else {
		b.recordSuccess()
	}
}

func (b *Bulkhead) handleFallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	fmt.Fprintf(w, `{"error": "bulkhead in circuit breaker state", "bulkhead": "%s", "state": "%s"}`, b.Name, b.GetState())
}

func (b *Bulkhead) GetStats() BulkheadStats {
	total := atomic.LoadInt64(&b.totalRequests)
	failed := atomic.LoadInt64(&b.failedRequests)
	
	var failureRate float64
	if total > 0 {
		failureRate = float64(failed) / float64(total)
	}
	
	return BulkheadStats{
		Name:              b.Name,
		State:             b.GetState(),
		ActiveConnections: atomic.LoadInt32(&b.activeConnections),
		TotalRequests:     total,
		FailedRequests:    failed,
		FailureRate:       failureRate,
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rw *statusRecorder) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (m *Manager) Create(name string, config BulkheadConfig) error {
	_, loaded := m.bulkheads.LoadOrStore(name, nil)
	if loaded {
		return errors.New("bulkhead already exists")
	}
	
	b, err := NewBulkhead(name, config)
	if err != nil {
		return err
	}
	
	m.bulkheads.Store(name, b)
	return nil
}

func (m *Manager) Delete(name string) error {
	_, loaded := m.bulkheads.LoadAndDelete(name)
	if !loaded {
		return errors.New("bulkhead not found")
	}
	return nil
}

func (m *Manager) Get(name string) (*Bulkhead, bool) {
	value, ok := m.bulkheads.Load(name)
	if !ok || value == nil {
		return nil, false
	}
	return value.(*Bulkhead), true
}

func (m *Manager) List() []BulkheadStats {
	var stats []BulkheadStats
	m.bulkheads.Range(func(key, value interface{}) bool {
		if b, ok := value.(*Bulkhead); ok {
			stats = append(stats, b.GetStats())
		}
		return true
	})
	return stats
}

type CreateBulkheadRequest struct {
	Name             string  `json:"name" binding:"required"`
	BackendAddress   string  `json:"backend_address" binding:"required"`
	MaxConcurrent    int     `json:"max_concurrent" binding:"required,gt=0"`
	FailureThreshold float64 `json:"failure_threshold" binding:"required,gt=0,lte=1"`
	BreakDuration    string  `json:"break_duration" binding:"required"`
}

type UpdateConfigRequest struct {
	BackendAddress   string  `json:"backend_address"`
	MaxConcurrent    int     `json:"max_concurrent"`
	FailureThreshold float64 `json:"failure_threshold"`
	BreakDuration    string  `json:"break_duration"`
}

func createBulkhead(c *gin.Context) {
	var req CreateBulkheadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	breakDuration, err := time.ParseDuration(req.BreakDuration)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid break_duration: " + err.Error()})
		return
	}
	
	config := BulkheadConfig{
		BackendAddress:   req.BackendAddress,
		MaxConcurrent:    req.MaxConcurrent,
		FailureThreshold: req.FailureThreshold,
		BreakDuration:    breakDuration,
		WindowSize:       20,
	}
	
	if err := manager.Create(req.Name, config); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{"message": "bulkhead created", "name": req.Name})
}

func deleteBulkhead(c *gin.Context) {
	name := c.Param("name")
	if err := manager.Delete(name); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "bulkhead deleted", "name": name})
}

func listBulkheads(c *gin.Context) {
	stats := manager.List()
	c.JSON(http.StatusOK, stats)
}

func updateConfig(c *gin.Context) {
	name := c.Param("name")
	b, ok := manager.Get(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "bulkhead not found"})
		return
	}
	
	var req UpdateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	oldConfig := b.GetConfig()
	newConfig := oldConfig
	
	if req.BackendAddress != "" {
		newConfig.BackendAddress = req.BackendAddress
	}
	if req.MaxConcurrent > 0 {
		newConfig.MaxConcurrent = req.MaxConcurrent
	}
	if req.FailureThreshold > 0 && req.FailureThreshold <= 1 {
		newConfig.FailureThreshold = req.FailureThreshold
	}
	if req.BreakDuration != "" {
		breakDuration, err := time.ParseDuration(req.BreakDuration)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid break_duration: " + err.Error()})
			return
		}
		newConfig.BreakDuration = breakDuration
	}
	
	b.SetConfig(newConfig)
	c.JSON(http.StatusOK, gin.H{"message": "config updated", "name": name})
}

func proxyHandler(c *gin.Context) {
	name := c.Param("name")
	b, ok := manager.Get(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "bulkhead not found"})
		return
	}
	
	b.ServeHTTP(c.Writer, c.Request)
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8608"
	}
	if _, err := strconv.Atoi(port); err != nil {
		port = "8608"
	}
	return ":" + port
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	
	r.POST("/bulkheads", createBulkhead)
	r.DELETE("/bulkheads/:name", deleteBulkhead)
	r.GET("/bulkheads", listBulkheads)
	r.PUT("/bulkheads/:name/config", updateConfig)
	r.Any("/:name/*path", proxyHandler)
	
	port := getPort()
	fmt.Printf("Server starting on port %s\n", port)
	if err := r.Run(port); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
