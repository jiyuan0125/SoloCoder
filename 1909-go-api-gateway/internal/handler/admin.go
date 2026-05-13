package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gateway/internal/store"
)

type CreateKeyRequest struct {
	RateLimit int `json:"rate_limit" binding:"required,min=1"`
}

type KeyResponse struct {
	ID        string `json:"key_id"`
	Secret    string `json:"secret,omitempty"`
	RateLimit int    `json:"rate_limit"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func CreateKey(keyStore *store.KeyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateKeyRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		key, err := keyStore.Create(req.RateLimit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create key"})
			return
		}

		c.JSON(http.StatusCreated, KeyResponse{
			ID:        key.ID,
			Secret:    key.Secret,
			RateLimit: key.RateLimit,
			Status:    string(key.Status),
			CreatedAt: key.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
}

func RevokeKey(keyStore *store.KeyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		keyID := c.Param("kid")
		if keyStore.Revoke(keyID) {
			c.JSON(http.StatusOK, gin.H{"message": "key revoked"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		}
	}
}

func ListKeys(keyStore *store.KeyStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		keys := keyStore.List()
		resp := make([]KeyResponse, 0, len(keys))
		for _, k := range keys {
			resp = append(resp, KeyResponse{
				ID:        k.ID,
				RateLimit: k.RateLimit,
				Status:    string(k.Status),
				CreatedAt: k.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			})
		}
		c.JSON(http.StatusOK, resp)
	}
}
