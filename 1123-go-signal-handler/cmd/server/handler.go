package main

import (
	"encoding/json"
	"net/http"

	"github.com/example/graceful-shutdown/pkg/api"
)

func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := s.graceful.Status()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *server) handleGraceful(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	state := s.graceful.GetState()
	if state != api.StateIdle {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.ShutdownResponse{
			Success: false,
			Message: "shutdown already in progress",
		})
		return
	}

	go s.graceful.Shutdown()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.ShutdownResponse{
		Success: true,
		Message: "graceful shutdown initiated",
	})
}

func (s *server) handleForce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.ShutdownResponse{
		Success: true,
		Message: "force exit initiated",
	})

	s.graceful.ForceExit()
}

func (s *server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	err := s.graceful.RegisterDynamic(req.Name, req.Order)

	resp := api.RegisterResponse{}
	if err != nil {
		resp.Success = false
		resp.Message = err.Error()
	} else {
		resp.Success = true
		resp.Message = "callback registered successfully"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "graceful-shutdown-server",
		"endpoints": []string{
			"GET  /status    - 查看状态和回调列表",
			"POST /graceful  - 发起优雅退出",
			"POST /force     - 发起强制退出",
			"POST /register  - 动态注册回调",
		},
	})
}
