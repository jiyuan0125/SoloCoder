package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type APIKey struct {
	Key      string   `json:"key"`
	Prefixes []string `json:"prefixes"`
	ID       string   `json:"id"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	Token string `json:"token"`
}

type CreateAPIKeyRequest struct {
	Prefixes []string `json:"prefixes"`
}

type Pagination struct {
	Page      int `json:"page"`
	PageSize  int `json:"page_size"`
	Total     int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type APIKeyListResponse struct {
	Data       []APIKey   `json:"data"`
	Pagination Pagination `json:"pagination"`
}

var (
	jwtSecret        []byte
	blacklist        = make(map[string]time.Time)
	refreshRateLimit = make(map[string]time.Time)
	apiKeys          = make(map[string]APIKey)
	apiKeyIDCounter  = 0
	mu               sync.RWMutex
	validUsers       = map[string]string{
		"admin": "password123",
	}
)

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		fmt.Fprintln(os.Stderr, "Error: JWT_SECRET environment variable is required")
		os.Exit(1)
	}
	jwtSecret = []byte(secret)
}

func generateToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(30 * time.Minute).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func validateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, fmt.Errorf("Token 已过期")
			}
			if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				return nil, fmt.Errorf("签名无效")
			}
		}
		return nil, fmt.Errorf("签名无效")
	}

	mu.RLock()
	_, isBlacklisted := blacklist[tokenString]
	mu.RUnlock()
	if isBlacklisted {
		return nil, fmt.Errorf("Token 已过期")
	}

	return token, nil
}

func isTokenRefreshable(tokenString string) error {
	mu.RLock()
	defer mu.RUnlock()

	if _, exists := blacklist[tokenString]; exists {
		return fmt.Errorf("Token 已过期")
	}

	if lastRefresh, exists := refreshRateLimit[tokenString]; exists {
		if time.Since(lastRefresh) < 1*time.Minute {
			return fmt.Errorf("刷新请求过于频繁")
		}
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "未提供认证凭证")
			return
		}

		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			_, err := validateToken(tokenString)
			if err != nil {
				writeError(w, http.StatusUnauthorized, err.Error())
				return
			}
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(authHeader, "ApiKey ") {
			apiKeyString := strings.TrimPrefix(authHeader, "ApiKey ")
			mu.RLock()
			apiKey, exists := apiKeys[apiKeyString]
			mu.RUnlock()

			if !exists {
				writeError(w, http.StatusUnauthorized, "API Key 无效")
				return
			}

			hasPermission := false
			for _, prefix := range apiKey.Prefixes {
				if strings.HasPrefix(r.URL.Path, prefix) {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				writeError(w, http.StatusForbidden, "权限不足")
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		writeError(w, http.StatusUnauthorized, "认证格式无效")
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	mu.RLock()
	validPassword, exists := validUsers[req.Username]
	mu.RUnlock()

	if !exists || validPassword != req.Password {
		writeError(w, http.StatusUnauthorized, "认证失败")
		return
	}

	token, err := generateToken(req.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "生成 Token 失败")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

func refreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	authHeader := r.Header.Get("Authorization")
	var oldToken string

	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		oldToken = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		var req RefreshRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "请求体格式错误")
			return
		}
		oldToken = req.Token
	}

	if oldToken == "" {
		writeError(w, http.StatusBadRequest, "未提供 Token")
		return
	}

	token, err := jwt.Parse(oldToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				writeError(w, http.StatusUnauthorized, "签名无效")
				return
			}
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					exp := int64(claims["exp"].(float64))
					expTime := time.Unix(exp, 0)
					if time.Since(expTime) > 5*time.Minute {
						writeError(w, http.StatusUnauthorized, "Token 已过期")
						return
					}
				}
			}
		} else {
			writeError(w, http.StatusUnauthorized, "签名无效")
			return
		}
	}

	if err := isTokenRefreshable(oldToken); err != nil {
		writeError(w, http.StatusTooManyRequests, err.Error())
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		writeError(w, http.StatusBadRequest, "无效的 Token")
		return
	}

	username, _ := claims["username"].(string)

	mu.Lock()
	blacklist[oldToken] = time.Now()
	refreshRateLimit[oldToken] = time.Now()
	mu.Unlock()

	newToken, err := generateToken(username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "生成 Token 失败")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token": newToken,
	})
}

func listAPIKeysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if val, err := strconv.Atoi(ps); err == nil && val > 0 {
			pageSize = val
		}
	}

	mu.RLock()
	keys := make([]APIKey, 0, len(apiKeys))
	for _, key := range apiKeys {
		keys = append(keys, key)
	}
	mu.RUnlock()

	total := len(keys)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	totalPages := total / pageSize
	if total%pageSize != 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	response := APIKeyListResponse{
		Data: keys[start:end],
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	writeJSON(w, http.StatusOK, response)
}

func createAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求体格式错误")
		return
	}

	if len(req.Prefixes) == 0 {
		writeError(w, http.StatusBadRequest, "至少需要一个路径前缀")
		return
	}

	newKey := fmt.Sprintf("ak-%d-%s", time.Now().UnixNano(), randomString(8))
	apiKeyIDCounter++

	apiKey := APIKey{
		Key:      newKey,
		Prefixes: req.Prefixes,
		ID:       fmt.Sprintf("ak_%d", apiKeyIDCounter),
	}

	mu.Lock()
	apiKeys[newKey] = apiKey
	mu.Unlock()

	writeJSON(w, http.StatusCreated, apiKey)
}

func deleteAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api-keys/")
	if path == "" {
		writeError(w, http.StatusNotFound, "未找到")
		return
	}

	mu.Lock()
	defer mu.Unlock()

	for key, apiKey := range apiKeys {
		if apiKey.ID == path || key == path {
			delete(apiKeys, key)
			writeJSON(w, http.StatusOK, map[string]string{"message": "删除成功"})
			return
		}
	}

	writeError(w, http.StatusNotFound, "API Key 不存在")
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(result)
}

func apiKeysRouter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api-keys" || r.URL.Path == "/api-keys/" {
		if r.Method == http.MethodGet {
			listAPIKeysHandler(w, r)
		} else if r.Method == http.MethodPost {
			createAPIKeyHandler(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "方法不允许")
		}
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api-keys/") {
		deleteAPIKeyHandler(w, r)
		return
	}

	writeError(w, http.StatusNotFound, "未找到")
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "访问成功",
		"path":    r.URL.Path,
	})
}

func main() {
	http.HandleFunc("/auth/login", loginHandler)
	http.HandleFunc("/auth/refresh", refreshHandler)
	http.HandleFunc("/api-keys", authMiddleware(apiKeysRouter))
	http.HandleFunc("/api-keys/", authMiddleware(apiKeysRouter))

	http.HandleFunc("/", authMiddleware(protectedHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("认证网关启动中，监听端口: %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "服务器启动失败: %v\n", err)
		os.Exit(1)
	}
}
