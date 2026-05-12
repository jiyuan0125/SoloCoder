package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gateway/pkg/models"
)

type Plugin interface {
	Type() models.PluginType
	Name() string
	Config() map[string]any
	Execute(c *gin.Context) (continueChain bool)
}

type BasePlugin struct {
	name   string
	config map[string]any
}

func (p *BasePlugin) Name() string {
	return p.name
}

func (p *BasePlugin) Config() map[string]any {
	return p.config
}

type AuthPlugin struct {
	BasePlugin
}

func NewAuthPlugin(name string, config map[string]any) *AuthPlugin {
	return &AuthPlugin{BasePlugin{name: name, config: config}}
}

func (p *AuthPlugin) Type() models.PluginType {
	return models.PluginAuth
}

func (p *AuthPlugin) Execute(c *gin.Context) bool {
	token := c.GetHeader("Authorization")
	expected, _ := p.config["token"].(string)

	if expected != "" && token != "Bearer "+expected {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		c.Abort()
		return false
	}
	return true
}

type RatelimitPlugin struct {
	BasePlugin
	window      time.Duration
	maxRequests int
	requests    sync.Map
}

func NewRatelimitPlugin(name string, config map[string]any) *RatelimitPlugin {
	window := 1 * time.Minute
	maxRequests := 100

	if w, ok := config["window_ms"].(float64); ok {
		window = time.Duration(w) * time.Millisecond
	}
	if m, ok := config["max_requests"].(float64); ok {
		maxRequests = int(m)
	}

	return &RatelimitPlugin{
		BasePlugin:  BasePlugin{name: name, config: config},
		window:      window,
		maxRequests: maxRequests,
	}
}

func (p *RatelimitPlugin) Type() models.PluginType {
	return models.PluginRatelim
}

func (p *RatelimitPlugin) getKey(c *gin.Context) string {
	key, _ := p.config["key"].(string)
	switch key {
	case "header":
		h, _ := p.config["header_name"].(string)
		if h == "" {
			h = "X-User-ID"
		}
		return c.GetHeader(h)
	case "path":
		return c.Request.URL.Path
	default:
		return c.ClientIP()
	}
}

type limiterData struct {
	count  int
	expire time.Time
}

func (p *RatelimitPlugin) Execute(c *gin.Context) bool {
	key := p.getKey(c)
	now := time.Now()

	var data *limiterData
	if v, ok := p.requests.Load(key); ok {
		data = v.(*limiterData)
		if now.After(data.expire) {
			data = nil
		}
	}

	if data == nil {
		data = &limiterData{count: 0, expire: now.Add(p.window)}
	}

	data.count++
	p.requests.Store(key, data)

	if data.count > p.maxRequests {
		retryAfter := int(data.expire.Sub(now).Seconds()) + 1
		c.Header("Retry-After", strconv.Itoa(retryAfter))
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		c.Abort()
		return false
	}
	return true
}

type LogPlugin struct {
	BasePlugin
}

func NewLogPlugin(name string, config map[string]any) *LogPlugin {
	return &LogPlugin{BasePlugin{name: name, config: config}}
}

func (p *LogPlugin) Type() models.PluginType {
	return models.PluginLog
}

func (p *LogPlugin) Execute(c *gin.Context) bool {
	start := time.Now()
	rid := uuid.New().String()
	c.Set("request_id", rid)
	c.Header("X-Request-ID", rid)

	level, _ := p.config["level"].(string)
	if level == "debug" {
		fmt.Printf("[DEBUG] [%s] %s %s from %s\n", rid, c.Request.Method, c.Request.URL.Path, c.ClientIP())
	}

	c.Next()

	latency := time.Since(start)
	status := c.Writer.Status()
	fmt.Printf("[INFO] [%s] %s %s -> %d %s\n", rid, c.Request.Method, c.Request.URL.Path, status, latency)
	return true
}

type RewritePlugin struct {
	BasePlugin
	from    string
	to      string
	headers map[string]string
}

func NewRewritePlugin(name string, config map[string]any) *RewritePlugin {
	from, _ := config["from"].(string)
	to, _ := config["to"].(string)

	headers := make(map[string]string)
	if h, ok := config["headers"].(map[string]any); ok {
		for k, v := range h {
			headers[k] = fmt.Sprintf("%v", v)
		}
	}

	return &RewritePlugin{
		BasePlugin: BasePlugin{name: name, config: config},
		from:       from,
		to:         to,
		headers:    headers,
	}
}

func (p *RewritePlugin) Type() models.PluginType {
	return models.PluginRewrite
}

func (p *RewritePlugin) Execute(c *gin.Context) bool {
	if p.from != "" && p.to != "" {
		newPath := strings.Replace(c.Request.URL.Path, p.from, p.to, 1)
		if newPath != c.Request.URL.Path {
			c.Request.URL.Path = newPath
		}
	}

	for k, v := range p.headers {
		if strings.HasPrefix(k, "+") {
			c.Request.Header.Add(strings.TrimPrefix(k, "+"), v)
		} else if strings.HasPrefix(k, "-") {
			c.Request.Header.Del(strings.TrimPrefix(k, "-"))
		} else {
			c.Request.Header.Set(k, v)
		}
	}

	if body, ok := p.config["body"].(map[string]any); ok {
		newBody, err := json.Marshal(body)
		if err == nil {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(newBody))
			c.Request.ContentLength = int64(len(newBody))
			c.Request.Header.Set("Content-Type", "application/json")
		}
	}

	return true
}

var _ = json.Marshal

func BuildPlugins(configs []models.PluginConfig) []Plugin {
	var plugins []Plugin
	for _, cfg := range configs {
		if !cfg.Enable {
			continue
		}
		switch cfg.Type {
		case models.PluginAuth:
			plugins = append(plugins, NewAuthPlugin(cfg.Name, cfg.Config))
		case models.PluginRatelim:
			plugins = append(plugins, NewRatelimitPlugin(cfg.Name, cfg.Config))
		case models.PluginLog:
			plugins = append(plugins, NewLogPlugin(cfg.Name, cfg.Config))
		case models.PluginRewrite:
			plugins = append(plugins, NewRewritePlugin(cfg.Name, cfg.Config))
		}
	}
	return plugins
}
