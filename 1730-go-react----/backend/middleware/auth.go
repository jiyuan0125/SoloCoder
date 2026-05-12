package middleware

import (
	"net/http"
	"strings"

	"diploma-auth-system/models"
	"diploma-auth-system/store"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	store *store.Store
}

func NewAuthMiddleware(s *store.Store) *AuthMiddleware {
	return &AuthMiddleware{store: s}
}

func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "未授权访问",
			})
			return
		}
		
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "授权格式错误",
			})
			return
		}
		
		userID := parts[1]
		user, exists := m.store.GetUser(userID)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "用户不存在",
			})
			return
		}
		
		c.Set("user", user)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "未授权访问",
			})
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
			c.AbortWithStatusJSON(http.StatusForbidden, models.ErrorResponse{
				Error:   "forbidden",
				Message: "权限不足",
			})
			return
		}
		
		c.Next()
	}
}

func GetCurrentUser(c *gin.Context) *models.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}
	return user.(*models.User)
}
