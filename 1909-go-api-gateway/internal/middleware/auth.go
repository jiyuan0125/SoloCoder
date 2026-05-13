package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"strings"

	"github.com/gin-gonic/gin"

	"gateway/internal/store"
)

const authContextKey = "api_key_id"
const authRateLimitKey = "rate_limit"

type SecurityLogger interface {
	LogSecurity(ip, path, keyID string)
}

type defaultLogger struct{}

func (d *defaultLogger) LogSecurity(ip, path, keyID string) {
	log.Printf("[SECURITY] Auth failed - IP: %s, Path: %s, KeyID: %s", ip, path, keyID)
}

func Auth(keyStore *store.KeyStore, securityLogger SecurityLogger) gin.HandlerFunc {
	if securityLogger == nil {
		securityLogger = &defaultLogger{}
	}

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			securityLogger.LogSecurity(c.ClientIP(), c.FullPath(), "")
			c.AbortWithStatusJSON(401, gin.H{"error": "missing authorization header"})
			return
		}

		if !strings.HasPrefix(authHeader, "HMAC ") {
			securityLogger.LogSecurity(c.ClientIP(), c.FullPath(), "")
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid authorization scheme"})
			return
		}

		parts := strings.SplitN(authHeader[5:], ":", 2)
		if len(parts) != 2 {
			securityLogger.LogSecurity(c.ClientIP(), c.FullPath(), "")
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid authorization format"})
			return
		}

		keyID, clientSig := parts[0], parts[1]

		key, exists := keyStore.Get(keyID)
		if !exists || key.Status != store.StatusActive {
			securityLogger.LogSecurity(c.ClientIP(), c.FullPath(), keyID)
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid or revoked key"})
			return
		}

		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			securityLogger.LogSecurity(c.ClientIP(), c.FullPath(), keyID)
			c.AbortWithStatusJSON(413, gin.H{"error": "request body too large"})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		mac := hmac.New(sha256.New, []byte(key.Secret))
		mac.Write(bodyBytes)
		expectedSig := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(expectedSig), []byte(clientSig)) {
			securityLogger.LogSecurity(c.ClientIP(), c.FullPath(), keyID)
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid signature"})
			return
		}

		c.Set(authContextKey, keyID)
		c.Set(authRateLimitKey, key.RateLimit)
		c.Next()
	}
}

func GetKeyID(c *gin.Context) (string, bool) {
	keyID, ok := c.Get(authContextKey)
	if !ok {
		return "", false
	}
	return keyID.(string), true
}

func GetRateLimit(c *gin.Context) (int, bool) {
	rate, ok := c.Get(authRateLimitKey)
	if !ok {
		return 0, false
	}
	return rate.(int), true
}
