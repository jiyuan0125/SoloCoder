package handlers

import (
	"net/http"

	"jwt-gateway/storage"

	"github.com/gin-gonic/gin"
)

func GenerateKey(km *storage.KeyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := km.GenerateKey()
		c.JSON(http.StatusOK, gin.H{"key_id": id})
	}
}

func RotateKey(km *storage.KeyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		oldID, newID := km.RotateKey()
		c.JSON(http.StatusOK, gin.H{
			"old_key_id": oldID,
			"new_key_id": newID,
		})
	}
}

func ListKeys(km *storage.KeyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		keys := km.ListKeys()
		result := make([]map[string]interface{}, 0, len(keys))
		for _, key := range keys {
			result = append(result, map[string]interface{}{
				"id":         key.ID,
				"status":     key.Status,
				"created_at": key.CreatedAt,
				"rotated_at": key.RotatedAt,
			})
		}
		c.JSON(http.StatusOK, gin.H{"keys": result})
	}
}

func GetAuditLogs(al *storage.AuditLog) gin.HandlerFunc {
	return func(c *gin.Context) {
		entries := al.List()
		c.JSON(http.StatusOK, gin.H{"logs": entries})
	}
}
