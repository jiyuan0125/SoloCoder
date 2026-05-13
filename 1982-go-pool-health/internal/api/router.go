package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(handler *Handler) *gin.Engine {
	r := gin.Default()

	r.GET("/health", handler.HealthCheck)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/pools", handler.RegisterPool)
		v1.DELETE("/pools/:name", handler.UnregisterPool)
		v1.GET("/pools", handler.ListPools)
		v1.GET("/pools/:name/stats", handler.GetPoolStats)
		v1.GET("/pools/stats", handler.GetAllStats)
		v1.PATCH("/pools/:name/config", handler.UpdatePoolConfig)
		v1.GET("/pools/:name/config", handler.GetPoolConfig)
	}

	return r
}
