package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"circuit-breaker/circuitbreaker"
)

type App struct {
	manager    *circuitbreaker.Manager
	callbacks  map[string]circuitbreaker.StateChangeCallback
	callbackMu sync.RWMutex
}

type ConfigRequest struct {
	ServiceName      string `json:"service_name"`
	FailureThreshold int    `json:"failure_threshold"`
	TimeoutSeconds   int    `json:"timeout_seconds"`
}

type StatusResponse struct {
	ServiceName string                 `json:"service_name"`
	State       string                 `json:"state"`
	Records     []CallRecordResponse   `json:"records"`
}

type CallRecordResponse struct {
	Timestamp string `json:"timestamp"`
	Success   bool   `json:"success"`
	DurationMs int64 `json:"duration_ms"`
}

type CallRequest struct {
	ServiceName string `json:"service_name"`
	SimulateFail bool  `json:"simulate_fail"`
}

func NewApp() *App {
	app := &App{
		manager:   circuitbreaker.NewManager(5, 30*time.Second),
		callbacks: make(map[string]circuitbreaker.StateChangeCallback),
	}

	return app
}

func (app *App) registerCallback(serviceName string, callback circuitbreaker.StateChangeCallback) {
	app.callbackMu.Lock()
	defer app.callbackMu.Unlock()

	app.callbacks[serviceName] = callback
	app.manager.SetCallback(serviceName, func(sn string, from, to circuitbreaker.State) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[WARN] 服务 %s 状态变更回调执行异常: %v", serviceName, r)
			}
		}()
		callback(serviceName, from, to)
	})
}

func (app *App) handleSetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ServiceName == "" {
		http.Error(w, "service_name is required", http.StatusBadRequest)
		return
	}

	if req.FailureThreshold <= 0 {
		http.Error(w, "failure_threshold must be positive", http.StatusBadRequest)
		return
	}

	if req.TimeoutSeconds <= 0 {
		http.Error(w, "timeout_seconds must be positive", http.StatusBadRequest)
		return
	}

	cfg := circuitbreaker.Config{
		FailureThreshold: req.FailureThreshold,
		Timeout:          time.Duration(req.TimeoutSeconds) * time.Second,
	}

	app.manager.SetConfig(req.ServiceName, cfg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "配置已更新",
	})
}

func (app *App) handleGetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	serviceName := r.URL.Query().Get("service_name")
	if serviceName == "" {
		http.Error(w, "service_name is required", http.StatusBadRequest)
		return
	}

	state, records := app.manager.GetStatus(serviceName)

	recordResponses := make([]CallRecordResponse, 0, len(records))
	for _, rec := range records {
		recordResponses = append(recordResponses, CallRecordResponse{
			Timestamp:  rec.Timestamp.Format(time.RFC3339),
			Success:    rec.Success,
			DurationMs: rec.Duration.Milliseconds(),
		})
	}

	response := StatusResponse{
		ServiceName: serviceName,
		State:       string(state),
		Records:     recordResponses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) handleListServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	services := app.manager.ListServices()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"services": services,
	})
}

func (app *App) handleCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ServiceName == "" {
		http.Error(w, "service_name is required", http.StatusBadRequest)
		return
	}

	breaker := app.manager.GetOrCreate(req.ServiceName)

	if !breaker.Allow() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "服务熔断中",
		})
		return
	}

	start := time.Now()
	success := !req.SimulateFail
	duration := time.Since(start)

	breaker.Record(success, duration)

	if success {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "调用成功",
		})
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "调用失败",
		})
	}
}

func (app *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
	})
}

func main() {
	app := NewApp()

	app.registerCallback("payment-service", func(serviceName string, from, to circuitbreaker.State) {
		if to == circuitbreaker.StateOpen {
			log.Printf("[ALERT] 服务 %s 已熔断，状态从 %s 变为 %s", serviceName, from, to)
		} else if to == circuitbreaker.StateHalfOpen {
			log.Printf("[INFO] 服务 %s 进入半开状态，开始试探", serviceName)
		} else if to == circuitbreaker.StateClosed {
			log.Printf("[INFO] 服务 %s 已恢复正常", serviceName)
		}
	})

	app.registerCallback("order-service", func(serviceName string, from, to circuitbreaker.State) {
		if to == circuitbreaker.StateOpen {
			log.Printf("[ALERT] 服务 %s 已熔断，状态从 %s 变为 %s", serviceName, from, to)
		} else if to == circuitbreaker.StateHalfOpen {
			log.Printf("[INFO] 服务 %s 进入半开状态，开始试探", serviceName)
		} else if to == circuitbreaker.StateClosed {
			log.Printf("[INFO] 服务 %s 已恢复正常", serviceName)
		}
	})

	http.HandleFunc("/api/config", app.handleSetConfig)
	http.HandleFunc("/api/status", app.handleGetStatus)
	http.HandleFunc("/api/services", app.handleListServices)
	http.HandleFunc("/api/call", app.handleCall)
	http.HandleFunc("/health", app.handleHealth)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("Invalid PORT: %s", port)
	}

	log.Printf("熔断器服务启动，监听端口: %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}
