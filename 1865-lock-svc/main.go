package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type AcquireRequest struct {
	Resource string `json:"resource" binding:"required"`
	ClientID string `json:"client_id" binding:"required"`
	Timeout  int    `json:"timeout" binding:"required,min=1"`
}

type ReleaseRequest struct {
	Resource string `json:"resource" binding:"required"`
	ClientID string `json:"client_id" binding:"required"`
}

type RenewRequest struct {
	Resource string `json:"resource" binding:"required"`
	ClientID string `json:"client_id" binding:"required"`
}

func main() {
	manager := NewLockManager()
	go manager.cleanupExpiredLocks()

	r := gin.Default()

	r.POST("/locks/acquire", func(c *gin.Context) {
		var req AcquireRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := manager.Acquire(req.Resource, req.ClientID, time.Duration(req.Timeout)*time.Second)
		if err != nil {
			if err == ErrQueueFull || err == ErrWaitTimeout {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": err.Error()})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "acquired"})
	})

	r.POST("/locks/release", func(c *gin.Context) {
		var req ReleaseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := manager.Release(req.Resource, req.ClientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "released"})
	})

	r.POST("/locks/renew", func(c *gin.Context) {
		var req RenewRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := manager.Renew(req.Resource, req.ClientID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "renewed"})
	})

	r.GET("/locks", func(c *gin.Context) {
		statuses := manager.GetAllStatuses()
		c.JSON(http.StatusOK, statuses)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r.Run(":" + port)
}
