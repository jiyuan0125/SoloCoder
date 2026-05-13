package handlers

import (
	"net/http"
	"time"

	"jwt-gateway/storage"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

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
				"secret":     storage.EncodeKey(key.Secret),
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

func IssueTestToken(km *storage.KeyManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		activeKey := km.GetActiveKey()
		if activeKey == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No active key"})
			return
		}

		role := c.DefaultQuery("role", "service")
		sub := c.DefaultQuery("sub", "test-client")

		claims := Claims{
			Role: role,
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   sub,
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		token.Header["kid"] = activeKey.ID

		tokenStr, err := token.SignedString(activeKey.Secret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": tokenStr,
			"kid":   activeKey.ID,
			"role":  role,
			"sub":   sub,
		})
	}
}
