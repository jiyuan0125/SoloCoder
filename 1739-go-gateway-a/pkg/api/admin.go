package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gateway/pkg/models"
	"gateway/pkg/router"
)

func RegisterAdmin(r *gin.Engine, mgr *router.Manager) {
	admin := r.Group("/api/admin")
	{
		admin.GET("/routes", listRoutes(mgr))
		admin.POST("/routes", createRoute(mgr))
		admin.GET("/routes/:id", getRoute(mgr))
		admin.PUT("/routes/:id", updateRoute(mgr))
		admin.DELETE("/routes/:id", deleteRoute(mgr))
		admin.GET("/routes/:id/versions", listVersions(mgr))
		admin.POST("/routes/:id/rollback", rollbackRoute(mgr))
		admin.GET("/routes/:id/stats", getStats(mgr))
	}
}

func listRoutes(mgr *router.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		routes := mgr.List()
		resp := make([]gin.H, len(routes))
		for i, r := range routes {
			resp[i] = gin.H{
				"id":         r.ID,
				"path":       r.Path,
				"methods":    r.Methods,
				"upstream":   r.Upstream,
				"plugins":    r.Plugins,
				"enabled":    r.Enabled,
				"version":    r.Version,
				"created_at": r.CreatedAt,
				"updated_at": r.UpdatedAt,
			}
		}
		c.JSON(http.StatusOK, gin.H{"data": resp})
	}
}

func createRoute(mgr *router.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.Route
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		entry, err := mgr.Create(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": entry.Route})
	}
}

func getRoute(mgr *router.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		entry, ok := mgr.Get(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": entry.Route})
	}
}

func updateRoute(mgr *router.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var req models.Route
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		entry, err := mgr.Update(id, &req)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": entry.Route})
	}
}

func deleteRoute(mgr *router.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := mgr.Delete(id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	}
}

func listVersions(mgr *router.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		versions := mgr.Versions(id)
		c.JSON(http.StatusOK, gin.H{"data": versions})
	}
}

func rollbackRoute(mgr *router.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		versionStr := c.Query("version")
		if versionStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "version query required"})
			return
		}
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid version"})
			return
		}
		entry, err := mgr.Rollback(id, version)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": entry.Route})
	}
}

func getStats(mgr *router.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		entry, ok := mgr.Get(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
			return
		}
		stats := entry.Stats
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"route_id":       id,
				"total_requests": stats.TotalRequests,
				"total_errors":   stats.TotalErrors,
				"avg_latency_ms": stats.AvgLatency().Milliseconds(),
				"qps":            stats.QPS(),
				"error_rate":     stats.ErrorRate(),
			},
		})
	}
}
