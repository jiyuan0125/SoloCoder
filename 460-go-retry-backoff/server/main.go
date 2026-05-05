// Package main 实现了一个演示用的HTTP服务端
// 用于演示重试库的功能
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/example/retrybackoff/common"
)

// Server 定义服务端结构
type Server struct {
	// 配置
	port            int
	failureRate     float64
	delay           time.Duration
	enableRandom5xx bool

	// 统计
	requestCount    int64
	successCount    int64
	failureCount    int64
}

// NewServer 创建一个新的服务端
func NewServer(port int, failureRate float64, delay time.Duration, enableRandom5xx bool) *Server {
	return &Server{
		port:            port,
		failureRate:     failureRate,
		delay:           delay,
		enableRandom5xx: enableRandom5xx,
	}
}

// Start 启动服务端
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// 注册路由
	mux.HandleFunc("/api/health", s.healthHandler)
	mux.HandleFunc("/api/echo", s.echoHandler)
	mux.HandleFunc("/api/flaky", s.flakyHandler)
	mux.HandleFunc("/api/stats", s.statsHandler)
	mux.HandleFunc("/api/reset", s.resetHandler)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Server starting on %s...", addr)
	log.Printf("Configuration:")
	log.Printf("  - Port: %d", s.port)
	log.Printf("  - Failure rate: %.2f%%", s.failureRate*100)
	log.Printf("  - Response delay: %v", s.delay)
	log.Printf("  - Random 5xx errors: %v", s.enableRandom5xx)

	return http.ListenAndServe(addr, mux)
}

// healthHandler 健康检查接口
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&s.requestCount, 1)

	response := common.Response{
		ID:      "health-" + time.Now().Format("20060102150405"),
		Success: true,
		Message: "Service is healthy",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	atomic.AddInt64(&s.successCount, 1)
}

// echoHandler 回显接口
func (s *Server) echoHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&s.requestCount, 1)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	var request common.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 添加延迟
	if s.delay > 0 {
		time.Sleep(s.delay)
	}

	response := common.Response{
		ID:      request.ID,
		Success: true,
		Message: "Echo: " + request.Message,
		Data: map[string]interface{}{
			"original_message": request.Message,
			"timestamp":        time.Now().Unix(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	atomic.AddInt64(&s.successCount, 1)
}

// flakyHandler 不稳定接口，用于测试重试
// 会根据配置随机返回错误或延迟
func (s *Server) flakyHandler(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&s.requestCount, 1)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 读取请求体
	var request common.Request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 检查是否应该失败
	if rand.Float64() < s.failureRate {
		atomic.AddInt64(&s.failureCount, 1)

		// 决定返回什么类型的错误
		if s.enableRandom5xx {
			// 随机选择5xx错误码
			errorCodes := []int{
				http.StatusInternalServerError,
				http.StatusBadGateway,
				http.StatusServiceUnavailable,
				http.StatusGatewayTimeout,
				http.StatusTooManyRequests, // 429
			}
			statusCode := errorCodes[rand.Intn(len(errorCodes))]

			response := common.Response{
				ID:      request.ID,
				Success: false,
				Message: fmt.Sprintf("Server error (simulated)"),
				Error:   fmt.Sprintf("HTTP %d", statusCode),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(response)
			return
		}

		// 返回500错误
		response := common.Response{
			ID:      request.ID,
			Success: false,
			Message: "Internal server error (simulated)",
			Error:   "HTTP 500",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 添加延迟
	if s.delay > 0 {
		time.Sleep(s.delay)
	}

	// 成功响应
	response := common.Response{
		ID:      request.ID,
		Success: true,
		Message: "Request processed successfully",
		Data: map[string]interface{}{
			"original_message": request.Message,
			"timestamp":        time.Now().Unix(),
			"processed":        true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	atomic.AddInt64(&s.successCount, 1)
}

// statsHandler 统计接口
func (s *Server) statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"total_requests":   atomic.LoadInt64(&s.requestCount),
		"successful_count": atomic.LoadInt64(&s.successCount),
		"failed_count":     atomic.LoadInt64(&s.failureCount),
		"failure_rate":     s.failureRate,
		"response_delay":   s.delay.String(),
		"random_5xx":       s.enableRandom5xx,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// resetHandler 重置统计
func (s *Server) resetHandler(w http.ResponseWriter, r *http.Request) {
	atomic.StoreInt64(&s.requestCount, 0)
	atomic.StoreInt64(&s.successCount, 0)
	atomic.StoreInt64(&s.failureCount, 0)

	response := common.Response{
		ID:      "reset-" + time.Now().Format("20060102150405"),
		Success: true,
		Message: "Stats reset successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// 解析命令行参数
	port := flag.Int("port", 8080, "Server port")
	failureRate := flag.Float64("failure-rate", 0.3, "Failure rate (0.0-1.0)")
	delay := flag.Duration("delay", 0, "Response delay")
	enableRandom5xx := flag.Bool("random-5xx", true, "Enable random 5xx error codes")

	flag.Parse()

	// 验证参数
	if *failureRate < 0 || *failureRate > 1 {
		log.Fatal("Failure rate must be between 0 and 1")
	}

	// 创建并启动服务端
	server := NewServer(*port, *failureRate, *delay, *enableRandom5xx)
	if err := server.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
