package proxy

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ProxyService struct {
	BackendManager *BackendManager
	StatsCollector *StatsCollector
}

func NewProxyService(bm *BackendManager) *ProxyService {
	return &ProxyService{
		BackendManager: bm,
		StatsCollector: NewStatsCollector(),
	}
}

func (ps *ProxyService) Forward(c *gin.Context) {
	backend := ps.BackendManager.MatchBackend(c.Request.URL.Path)
	if backend == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "No healthy backend available for this path",
		})
		return
	}

	start := time.Now()

	c.Request.URL.Path = c.Request.URL.Path[len(backend.Path):]
	if c.Request.URL.Path == "" {
		c.Request.URL.Path = "/"
	}

	backend.Proxy.ServeHTTP(c.Writer, c.Request)

	duration := time.Since(start)
	ps.StatsCollector.RecordRequest(backend.Name, duration)
}

func (ps *ProxyService) GetStatus(c *gin.Context) {
	backends := ps.BackendManager.GetAllBackends()
	stats := ps.StatsCollector.GetAllStats()

	type BackendStatus struct {
		Name        string  `json:"name"`
		URL         string  `json:"url"`
		Path        string  `json:"path"`
		Healthy     bool    `json:"healthy"`
		RequestCount int64  `json:"request_count"`
		AvgDurationMs float64 `json:"avg_duration_ms"`
	}

	result := make([]BackendStatus, 0, len(backends))
	for name, backend := range backends {
		backend.mu.RLock()
		healthy := backend.Healthy
		backend.mu.RUnlock()

		var count int64 = 0
		var avg float64 = 0.0
		if stat, ok := stats[name]; ok {
			count = stat.RequestCount
			avg = stat.AvgDuration
		}

		result = append(result, BackendStatus{
			Name:           name,
			URL:            backend.URL.String(),
			Path:           backend.Path,
			Healthy:        healthy,
			RequestCount:   count,
			AvgDurationMs:  avg,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"backends": result,
	})
}
