package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateRule struct {
	Dimension   string `json:"dimension"`
	Window      int    `json:"window"`
	Limit       int    `json:"limit"`
	DailyQuota  int    `json:"daily_quota"`
}

type APIKey struct {
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
	Revoked   bool      `json:"revoked"`
}

type BlacklistEntry struct {
	Entry     string    `json:"entry"`
	CreatedAt time.Time `json:"created_at"`
}

type RateLog struct {
	ID        int       `json:"id"`
	IP        string    `json:"ip"`
	APIKey    string    `json:"api_key"`
	Path      string    `json:"path"`
	Triggered time.Time `json:"triggered_at"`
}

type RateCounter struct {
	mu           sync.Mutex
	windowCount  int
	windowStart  time.Time
	dailyCount   int
	dailyStart   time.Time
}

var (
	rulesMu     sync.RWMutex
	rules       map[string]*RateRule
	keysMu      sync.RWMutex
	apiKeys     map[string]*APIKey
	blacklistMu sync.RWMutex
	blacklist   map[string]*BlacklistEntry
	countersMu  sync.RWMutex
	counters    map[string]*RateCounter
	logsMu      sync.Mutex
	logs        []*RateLog
	logID       int
)

func init() {
	rules = make(map[string]*RateRule)
	apiKeys = make(map[string]*APIKey)
	blacklist = make(map[string]*BlacklistEntry)
	counters = make(map[string]*RateCounter)
	logs = make([]*RateLog, 0)
}

func generateKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func getCounter(key string) *RateCounter {
	countersMu.Lock()
	defer countersMu.Unlock()
	if c, ok := counters[key]; ok {
		return c
	}
	c := &RateCounter{
		windowStart: time.Now(),
		dailyStart:  time.Now().Truncate(24 * time.Hour),
	}
	counters[key] = c
	return c
}

func isBlacklisted(ip, apiKey string) bool {
	blacklistMu.RLock()
	defer blacklistMu.RUnlock()
	if _, ok := blacklist[ip]; ok {
		return true
	}
	if apiKey != "" {
		if _, ok := blacklist[apiKey]; ok {
			return true
		}
	}
	return false
}

func isKeyRevoked(key string) bool {
	keysMu.RLock()
	defer keysMu.RUnlock()
	if k, ok := apiKeys[key]; ok {
		return k.Revoked
	}
	return false
}

func checkRateLimit(dimension, identifier string, rule *RateRule) (bool, time.Duration) {
	if rule == nil {
		return true, 0
	}
	c := getCounter(dimension + ":" + identifier)
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if now.Sub(c.windowStart) >= time.Duration(rule.Window)*time.Minute {
		c.windowCount = 0
		c.windowStart = now
	}

	if now.Sub(c.dailyStart) >= 24*time.Hour {
		c.dailyCount = 0
		c.dailyStart = now.Truncate(24 * time.Hour)
	}

	if c.windowCount >= rule.Limit {
		retryAfter := time.Duration(rule.Window)*time.Minute - now.Sub(c.windowStart)
		return false, retryAfter
	}

	if rule.DailyQuota > 0 && c.dailyCount >= rule.DailyQuota {
		retryAfter := 24*time.Hour - now.Sub(c.dailyStart)
		return false, retryAfter
	}

	c.windowCount++
	c.dailyCount++
	return true, 0
}

func logRateLimit(ip, apiKey, path string) {
	logsMu.Lock()
	defer logsMu.Unlock()
	logID++
	logs = append(logs, &RateLog{
		ID:        logID,
		IP:        ip,
		APIKey:    apiKey,
		Path:      path,
		Triggered: time.Now(),
	})
}

func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		apiKey := c.GetHeader("X-API-Key")
		path := c.Request.URL.Path

		if isBlacklisted(ip, apiKey) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		if apiKey != "" && isKeyRevoked(apiKey) {
			c.Header("Retry-After", strconv.Itoa(int(24*time.Hour.Seconds())))
			c.AbortWithStatus(http.StatusTooManyRequests)
			logRateLimit(ip, apiKey, path)
			return
		}

		var maxRetryAfter time.Duration
		limited := false

		rulesMu.RLock()
		ipRule := rules["ip"]
		keyRule := rules["apikey"]
		rulesMu.RUnlock()

		if ipRule != nil {
			allowed, retryAfter := checkRateLimit("ip", ip, ipRule)
			if !allowed {
				limited = true
				maxRetryAfter = retryAfter
			}
		}

		if apiKey != "" && keyRule != nil && !limited {
			allowed, retryAfter := checkRateLimit("apikey", apiKey, keyRule)
			if !allowed {
				limited = true
				if retryAfter > maxRetryAfter {
					maxRetryAfter = retryAfter
				}
			}
		}

		if limited {
			c.Header("Retry-After", strconv.Itoa(int(maxRetryAfter.Seconds())))
			c.AbortWithStatus(http.StatusTooManyRequests)
			logRateLimit(ip, apiKey, path)
			return
		}

		c.Next()
	}
}

func createKey(c *gin.Context) {
	key := generateKey()
	keysMu.Lock()
	apiKeys[key] = &APIKey{
		Key:       key,
		CreatedAt: time.Now(),
		Revoked:   false,
	}
	keysMu.Unlock()
	c.JSON(http.StatusCreated, gin.H{"key": key})
}

func deleteKey(c *gin.Context) {
	key := c.Param("key")
	keysMu.Lock()
	if k, ok := apiKeys[key]; ok {
		k.Revoked = true
		c.Status(http.StatusNoContent)
	} else {
		c.Status(http.StatusNotFound)
	}
	keysMu.Unlock()
}

func listKeys(c *gin.Context) {
	keysMu.RLock()
	result := make([]*APIKey, 0, len(apiKeys))
	for _, k := range apiKeys {
		result = append(result, k)
	}
	keysMu.RUnlock()
	c.JSON(http.StatusOK, result)
}

func createRule(c *gin.Context) {
	var rule RateRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if rule.Dimension != "ip" && rule.Dimension != "apikey" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dimension must be 'ip' or 'apikey'"})
		return
	}
	if rule.Window <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "window must be positive"})
		return
	}
	if rule.Limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be positive"})
		return
	}
	rulesMu.Lock()
	rules[rule.Dimension] = &rule
	rulesMu.Unlock()
	c.JSON(http.StatusCreated, rule)
}

func listRules(c *gin.Context) {
	rulesMu.RLock()
	result := make([]*RateRule, 0, len(rules))
	for _, r := range rules {
		result = append(result, r)
	}
	rulesMu.RUnlock()
	c.JSON(http.StatusOK, result)
}

func addBlacklist(c *gin.Context) {
	var req struct {
		Entry string `json:"entry" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	blacklistMu.Lock()
	blacklist[req.Entry] = &BlacklistEntry{
		Entry:     req.Entry,
		CreatedAt: time.Now(),
	}
	blacklistMu.Unlock()
	c.Status(http.StatusNoContent)
}

func removeBlacklist(c *gin.Context) {
	entry := c.Param("entry")
	blacklistMu.Lock()
	if _, ok := blacklist[entry]; ok {
		delete(blacklist, entry)
		c.Status(http.StatusNoContent)
	} else {
		c.Status(http.StatusNotFound)
	}
	blacklistMu.Unlock()
}

func listBlacklist(c *gin.Context) {
	blacklistMu.RLock()
	result := make([]*BlacklistEntry, 0, len(blacklist))
	for _, b := range blacklist {
		result = append(result, b)
	}
	blacklistMu.RUnlock()
	c.JSON(http.StatusOK, result)
}

func queryLogs(c *gin.Context) {
	startStr := c.Query("start")
	endStr := c.Query("end")
	ip := c.Query("ip")
	apiKey := c.Query("api_key")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse(time.RFC3339, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start time"})
			return
		}
	}

	if endStr != "" {
		end, err = time.Parse(time.RFC3339, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end time"})
			return
		}
	}

	logsMu.Lock()
	result := make([]*RateLog, 0)
	for _, l := range logs {
		if !start.IsZero() && l.Triggered.Before(start) {
			continue
		}
		if !end.IsZero() && l.Triggered.After(end) {
			continue
		}
		if ip != "" && l.IP != ip {
			continue
		}
		if apiKey != "" && l.APIKey != apiKey {
			continue
		}
		result = append(result, l)
	}
	logsMu.Unlock()
	c.JSON(http.StatusOK, result)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8304"
	}

	r := gin.Default()

	admin := r.Group("/")
	{
		admin.POST("/keys", createKey)
		admin.DELETE("/keys/:key", deleteKey)
		admin.GET("/keys", listKeys)

		admin.POST("/rules", createRule)
		admin.GET("/rules", listRules)

		admin.POST("/blacklist", addBlacklist)
		admin.DELETE("/blacklist/:entry", removeBlacklist)
		admin.GET("/blacklist", listBlacklist)

		admin.GET("/logs", queryLogs)
	}

	api := r.Group("/", rateLimitMiddleware())
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		api.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	log.Printf("Rate Gate server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
