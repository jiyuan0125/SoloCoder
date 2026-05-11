package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/pingtool/pkg/model"
	"github.com/pingtool/pkg/ping"
)

type PingHandler struct {
	manager *TaskManager
}

func NewPingHandler(manager *TaskManager) *PingHandler {
	return &PingHandler{
		manager: manager,
	}
}

func (h *PingHandler) RegisterRoutes(r *mux.Router) {
	api := r.PathPrefix("/api/ping").Subrouter()
	api.HandleFunc("", h.handleStart).Methods("POST")
	api.HandleFunc("/{id}", h.handleGetStatusOrStream).Methods("GET")
	api.HandleFunc("/{id}", h.handleStop).Methods("DELETE")
}

func (h *PingHandler) handleGetStatusOrStream(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	
	accept := r.Header.Get("Accept")
	if accept == "text/event-stream" {
		h.handleStream(w, r, taskID)
		return
	}
	h.handleStatus(w, r)
}

func (h *PingHandler) handleStart(w http.ResponseWriter, r *http.Request) {
	var req model.PingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Target == "" {
		http.Error(w, "target is required", http.StatusBadRequest)
		return
	}

	config := &ping.Config{
		Target:   req.Target,
		Count:    req.Count,
		Interval: req.Interval,
		TTL:      req.TTL,
	}

	task, err := h.manager.CreateTask(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := model.PingResponse{
		TaskID: task.ID,
		Status: "started",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *PingHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	task, exists := h.manager.GetTask(taskID)
	if !exists {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	stats := task.Pinger.GetStats()
	config := task.Pinger.GetConfig()

	task.mu.RLock()
	isRunning := task.IsRunning
	createdAt := task.CreatedAt
	finishedAt := task.FinishedAt
	task.mu.RUnlock()

	status := model.TaskStatus{
		ID:              task.ID,
		Target:          config.Target,
		Count:           config.Count,
		Interval:        config.Interval,
		TTL:             config.TTL,
		PacketsSent:     stats.PacketsSent,
		PacketsReceived: stats.PacketsReceived,
		PacketLoss:      stats.PacketLoss,
		MinRTT:          stats.MinRTT.Nanoseconds(),
		MaxRTT:          stats.MaxRTT.Nanoseconds(),
		AvgRTT:          stats.AvgRTT.Nanoseconds(),
		StdDevRTT:       stats.StdDevRTT.Nanoseconds(),
		IsRunning:       isRunning,
		CreatedAt:       createdAt,
		FinishedAt:      finishedAt,
		TimeExceeded:    make([]*model.TimeExceededInfo, 0, len(stats.TimeExceeded)),
	}

	for _, te := range stats.TimeExceeded {
		status.TimeExceeded = append(status.TimeExceeded, &model.TimeExceededInfo{
			RouterIP: te.RouterIP.String(),
			TargetIP: te.OriginalTarget.String(),
			Sequence: te.Seq,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *PingHandler) handleStop(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	taskID := vars["id"]
	success := h.manager.StopTask(taskID)
	
	if !success {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	resp := model.StopResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *PingHandler) handleStream(w http.ResponseWriter, r *http.Request, taskID string) {
	task, exists := h.manager.GetTask(taskID)
	if !exists {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	_ = r

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case result, ok := <-task.ResultChan:
			if !ok {
				return
			}
			var data map[string]interface{}
			if result.TimeExceeded != nil {
				data = map[string]interface{}{
					"type":       "timeout",
					"seq":        result.Seq,
					"router_ip":  result.TimeExceeded.RouterIP.String(),
					"target_ip":  result.TimeExceeded.OriginalTarget.String(),
				}
			} else {
				data = map[string]interface{}{
					"type":      "reply",
					"seq":       result.Seq,
					"rtt_ns":    result.RTT.Nanoseconds(),
					"target_ip": result.TargetIP.String(),
				}
			}
			jsonData, _ := json.Marshal(data)
			w.Write([]byte("data: " + string(jsonData) + "\n\n"))
			flusher.Flush()
		case <-r.Context().Done():
			return
		case <-time.After(100 * time.Millisecond):
			task.mu.RLock()
			running := task.IsRunning
			task.mu.RUnlock()
			if !running {
				select {
				case _, ok := <-task.ResultChan:
					if !ok {
						return
					}
				default:
				}
			}
		}
	}
}
