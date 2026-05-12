package middleware

import (
	"net/http"
	"school-connect/internal/models"
	"school-connect/internal/storage"
	"strings"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	store *storage.MemoryStore
}

func NewAuthMiddleware(store *storage.MemoryStore) *AuthMiddleware {
	return &AuthMiddleware{store: store}
}

func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证格式错误"})
			c.Abort()
			return
		}

		userID := parts[1]
		user := m.store.GetUserByID(userID)
		if user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("userID", user.ID)
		c.Next()
	}
}

func (m *AuthMiddleware) RoleRequired(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		u := user.(*models.User)
		hasRole := false
		for _, role := range roles {
			if u.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}

func GetUserID(c *gin.Context) string {
	return c.GetString("userID")
}
