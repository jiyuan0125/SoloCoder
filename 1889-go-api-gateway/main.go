package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gateway/aggregate"
	"gateway/auth"
	"gateway/config"
	"gateway/router"
)

type App struct {
	cfg          *config.Config
	jwtValidator *auth.JWTValidator
	apiKeyStore  *auth.APIKeyStore
	routeTable   *router.Table
}

func main() {
	cfg := config.Load()
	app := &App{
		cfg:          cfg,
		jwtValidator: auth.NewJWTValidator(cfg.JWTSecret),
		apiKeyStore:  auth.NewAPIKeyStore(),
		routeTable:   router.NewTable(),
	}

	app.routeTable.Add("/api/public", "http://localhost:8903/public", router.AuthNone, false)
	app.routeTable.Add("/api/protected", "http://localhost:8903/protected", router.AuthAPIKey, false)
	app.routeTable.Add("/api/user", "http://localhost:8903/user", router.AuthJWT, false)
	app.routeTable.Add("/api/services/", "http://localhost:8904/services/", router.AuthNone, true)

	r := gin.Default()

	r.POST("/admin/keys", app.handleCreateKey)
	r.DELETE("/admin/keys/:kid", app.handleRevokeKey)

	r.POST("/api/aggregate", app.concurrencyMiddleware(), app.authMiddleware(), app.handleAggregate)

	r.NoRoute(app.authMiddleware(), app.handleProxy)

	fmt.Printf("Gateway running on port %s\n", cfg.Port)
	r.Run(":" + cfg.Port)
}

func (app *App) handleCreateKey(c *gin.Context) {
	var body struct {
		KID string `json:"kid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "kid is required"})
		return
	}
	key, err := app.apiKeyStore.Create(body.KID)
	if err != nil {
		c.JSON(409, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, gin.H{"kid": key.ID, "key": key.Key})
}

func (app *App) handleRevokeKey(c *gin.Context) {
	kid := c.Param("kid")
	app.apiKeyStore.Revoke(kid)
	c.JSON(200, gin.H{"status": "revoked", "kid": kid})
}

func (app *App) handleAggregate(c *gin.Context) {
	var req aggregate.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}
	for _, u := range req.URLs {
		if !app.routeTable.IsBackendURL(u) {
			c.JSON(400, gin.H{"error": fmt.Sprintf("unknown backend URL: %s", u)})
			return
		}
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), aggregate.DefaultTimeout)
	defer cancel()
	results := aggregate.Run(ctx, req.URLs)
	c.JSON(200, gin.H{"results": results})
}

func (app *App) handleProxy(c *gin.Context) {
	path := c.Request.URL.Path
	route := app.routeTable.Match(path)
	if route == nil {
		c.JSON(404, gin.H{"error": "route not found"})
		return
	}
	app.proxyRequest(c, route)
}

func (app *App) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/admin/") {
			c.Next()
			return
		}
		if path == "/api/aggregate" {
			c.Next()
			return
		}
		route := app.routeTable.Match(path)
		if route == nil {
			c.Next()
			return
		}
		app.applyAuth(c, route)
	}
}

func (app *App) concurrencyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := extractAPIKey(c)
		if key == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "api key required for aggregation"})
			return
		}
		kid, ok := app.apiKeyStore.GetIDByKey(key)
		if !ok || !app.apiKeyStore.Validate(key) {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid api key"})
			return
		}
		if !app.apiKeyStore.TryAcquire(kid) {
			c.AbortWithStatusJSON(429, gin.H{"error": "too many concurrent aggregate requests"})
			return
		}
		defer app.apiKeyStore.Release(kid)
		c.Set("api_key_id", kid)
		c.Next()
	}
}

func (app *App) applyAuth(c *gin.Context, route *router.Route) {
	switch route.Auth {
	case router.AuthNone:
		c.Next()
	case router.AuthAPIKey:
		key := extractAPIKey(c)
		if key == "" || !app.apiKeyStore.Validate(key) {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid or missing api key"})
			return
		}
		kid, _ := app.apiKeyStore.GetIDByKey(key)
		c.Set("api_key_id", kid)
		c.Set("api_key", key)
		c.Next()
	case router.AuthJWT:
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing jwt token"})
			return
		}
		claims, err := app.jwtValidator.Validate(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid jwt token"})
			return
		}
		c.Set("jwt_claims", claims)
		c.Next()
	default:
		c.AbortWithStatusJSON(500, gin.H{"error": "unknown auth strategy"})
	}
}

func extractAPIKey(c *gin.Context) string {
	header := c.GetHeader("X-API-Key")
	if header != "" {
		return header
	}
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "ApiKey ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "ApiKey "))
	}
	return ""
}

func extractBearerToken(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	}
	return ""
}

func (app *App) proxyRequest(c *gin.Context, route *router.Route) {
	target := route.BackendURL + strings.TrimPrefix(c.Request.URL.Path, route.Path)
	client := &http.Client{}
	req, err := http.NewRequest(c.Request.Method, target, c.Request.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	for k, v := range c.Request.Header {
		req.Header[k] = v
	}
	q := c.Request.URL.Query()
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(502, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(502, gin.H{"error": err.Error()})
		return
	}
	for k, v := range resp.Header {
		c.Writer.Header()[k] = v
	}
	c.Writer.WriteHeader(resp.StatusCode)
	c.Writer.Write(body)
}
