package middleware

import (
	"net/http"
	"strings"
	"time"

	"jwt-gateway/storage"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func Auth(km *storage.KeyManager, auditLog *storage.AuditLog) gin.HandlerFunc {
	return func(c *gin.Context) {
		var clientID string
		result := "fail"

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			auditLog.Add(storage.AuditEntry{
				Timestamp: time.Now(),
				ClientID:  "",
				Path:      c.Request.URL.Path,
				Result:    "fail",
			})
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			kid, ok := token.Header["kid"].(string)
			if !ok {
				auditLog.Add(storage.AuditEntry{
					Timestamp: time.Now(),
					ClientID:  "",
					Path:      c.Request.URL.Path,
					Result:    "fail",
				})
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return nil, jwt.ErrSignatureInvalid
			}

			key := km.GetKey(kid)
			if key == nil || key.Status == storage.StatusDeprecated {
				auditLog.Add(storage.AuditEntry{
					Timestamp: time.Now(),
					ClientID:  "",
					Path:      c.Request.URL.Path,
					Result:    "fail",
				})
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return nil, jwt.ErrSignatureInvalid
			}

			return key.Secret, nil
		})

		if err != nil || !token.Valid {
			auditLog.Add(storage.AuditEntry{
				Timestamp: time.Now(),
				ClientID:  "",
				Path:      c.Request.URL.Path,
				Result:    "fail",
			})
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			auditLog.Add(storage.AuditEntry{
				Timestamp: time.Now(),
				ClientID:  "",
				Path:      c.Request.URL.Path,
				Result:    "fail",
			})
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		clientID = claims.Subject
		result = "pass"

		c.Set("claims", claims)
		c.Set("role", claims.Role)

		auditLog.Add(storage.AuditEntry{
			Timestamp: time.Now(),
			ClientID:  clientID,
			Path:      c.Request.URL.Path,
			Result:    result,
		})

		c.Next()
	}
}

func RequireAdmin(auditLog *storage.AuditLog) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		clientID, _ := c.Get("claims")
		sub := ""
		if claims, ok := clientID.(*Claims); ok {
			sub = claims.Subject
		}

		if !exists || role != "admin" {
			auditLog.Add(storage.AuditEntry{
				Timestamp: time.Now(),
				ClientID:  sub,
				Path:      c.Request.URL.Path,
				Result:    "forbidden",
			})
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.Next()
	}
}
