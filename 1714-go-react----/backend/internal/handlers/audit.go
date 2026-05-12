package handlers

import (
	"net/http"

	"health-supervision-system/internal/models"
	"health-supervision-system/internal/storage"

	"github.com/gin-gonic/gin"
)

func ListAuditLogs(c *gin.Context) {
	action := c.Query("action")
	resource := c.Query("resource")

	var logs []models.AuditLog
	query := storage.DB.Model(&models.AuditLog{})

	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}

	if err := query.Order("created_at DESC").Limit(1000).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func GetAuditLog(c *gin.Context) {
	id := c.Param("id")
	var log models.AuditLog

	if err := storage.DB.First(&log, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "audit log not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": log})
}

func BlockAuditModification(c *gin.Context) {
	c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "审计日志只能查看不能修改"})
}
