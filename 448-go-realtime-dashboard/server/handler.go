package main

import (
	"encoding/json"
	"net/http"
	"time"

	"realtime-dashboard/common"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type APIHandler struct {
	metricManager    *MetricManager
	connectionManager *ConnectionManager
	layoutManager    *LayoutManager
	upgrader         websocket.Upgrader
}

func NewAPIHandler(mm *MetricManager, cm *ConnectionManager, lm *LayoutManager) *APIHandler {
	return &APIHandler{
		metricManager:    mm,
		connectionManager: cm,
		layoutManager:    lm,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (h *APIHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}

	clientID := common.GenerateClientID()
	queuePos, success := h.connectionManager.AddClient(clientID, conn)

	if !success {
		response := common.SubscribeResponse{
			ClientID:      clientID,
			QueuePosition: &queuePos,
			Error:         "Server is busy, please wait in queue",
		}
		json.NewEncoder(w).Encode(response)
		conn.Close()
		return
	}

	defer func() {
		h.connectionManager.RemoveClient(clientID)
	}()

	go h.handleHeartbeats(clientID)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		h.handleClientMessage(clientID, msg)
	}
}

func (h *APIHandler) handleHeartbeats(clientID string) {
	ticker := time.NewTicker(common.HeartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		client, exists := h.connectionManager.GetClient(clientID)
		if !exists {
			return
		}

		if time.Since(client.LastHeartbeat) > 2*common.HeartbeatInterval {
			h.connectionManager.RemoveClient(clientID)
			return
		}

		heartbeatMsg := map[string]string{
			"type": "heartbeat",
			"time": time.Now().Format(time.RFC3339),
		}
		client.SendMessage(common.MsgTypeHeartbeat, heartbeatMsg)
	}
}

func (h *APIHandler) handleClientMessage(clientID string, rawMsg []byte) {
	msg, err := common.DecodeMessage(rawMsg)
	if err != nil {
		h.connectionManager.SendError(clientID, "Invalid message format")
		return
	}

	switch msg.Type {
	case common.MsgTypeSubscribe:
		var req common.SubscribeRequest
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			h.connectionManager.SendError(clientID, "Invalid subscribe request")
			return
		}

		subscribed := h.connectionManager.Subscribe(clientID, req.MetricKeys, req.RefreshRate)
		response := common.SubscribeResponse{
			ClientID:   clientID,
			Subscribed: subscribed,
		}

		client, _ := h.connectionManager.GetClient(clientID)
		client.SendMessage(common.MsgTypeSubscribe, response)

	case common.MsgTypeUnsubscribe:
		var req common.UnsubscribeRequest
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			h.connectionManager.SendError(clientID, "Invalid unsubscribe request")
			return
		}

		h.connectionManager.Unsubscribe(clientID, req.MetricKeys)

	case common.MsgTypeReconnect:
		var req common.ReconnectRequest
		if err := json.Unmarshal(msg.Payload, &req); err != nil {
			h.connectionManager.SendError(clientID, "Invalid reconnect request")
			return
		}

		response := h.connectionManager.HandleReconnect(req.ClientID, req.DisconnectTime)
		client, _ := h.connectionManager.GetClient(clientID)
		client.SendMessage(common.MsgTypeReconnectResponse, response)

	case common.MsgTypeHeartbeat:
		h.connectionManager.UpdateHeartbeat(clientID)
	}
}

func (h *APIHandler) GetServerStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	includeDetails := r.URL.Query().Get("details") == "true"

	status := common.ServerStatus{
		OnlineClients: h.connectionManager.GetOnlineCount(),
		MaxClients:    common.MaxClients,
		QueueSize:     h.connectionManager.GetQueueSize(),
		MetricsCount:  len(h.metricManager.GetAllMetrics()),
	}

	if includeDetails {
		status.LatestValues = h.metricManager.GetLatestValues()
		status.Clients = h.connectionManager.GetOnlineClients()
	}

	json.NewEncoder(w).Encode(status)
}

func (h *APIHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := h.metricManager.GetAllMetrics()
	result := make([]*common.Metric, 0, len(metrics))
	for _, m := range metrics {
		result = append(result, m)
	}

	json.NewEncoder(w).Encode(result)
}

func (h *APIHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	metricKey := vars["key"]

	metric, exists := h.metricManager.GetMetric(metricKey)
	if !exists {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(metric)
}

func (h *APIHandler) CreateMetric(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Key      string                `json:"key"`
		Metadata common.MetricMetadata `json:"metadata"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "Metric key is required", http.StatusBadRequest)
		return
	}

	success := h.metricManager.RegisterMetric(req.Key, req.Metadata)
	if !success {
		http.Error(w, "Metric already exists", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created", "key": req.Key})
}

func (h *APIHandler) UpdateMetricValue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	metricKey := vars["key"]

	var req struct {
		Value     float64    `json:"value"`
		Timestamp *time.Time `json:"timestamp,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	timestamp := time.Now()
	if req.Timestamp != nil {
		timestamp = *req.Timestamp
	}

	success := h.metricManager.UpdateMetric(metricKey, req.Value, timestamp)
	if !success {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	metric, _ := h.metricManager.GetMetric(metricKey)
	go h.connectionManager.PushMetricUpdate(metricKey, metric)

	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *APIHandler) SetMetricThreshold(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	metricKey := vars["key"]

	var threshold common.AlertThreshold
	if err := json.NewDecoder(r.Body).Decode(&threshold); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if threshold.MaxValue == nil && threshold.MinValue == nil {
		http.Error(w, "At least one threshold (max or min) is required", http.StatusBadRequest)
		return
	}

	success := h.metricManager.SetAlertThreshold(metricKey, &threshold)
	if !success {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "threshold_set"})
}

func (h *APIHandler) GetMetricHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	metricKey := vars["key"]

	query := r.URL.Query()
	startTimeStr := query.Get("start")
	endTimeStr := query.Get("end")

	var startTime, endTime time.Time
	var err error

	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			http.Error(w, "Invalid start time format", http.StatusBadRequest)
			return
		}
	} else {
		startTime = time.Now().Add(-24 * time.Hour)
	}

	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			http.Error(w, "Invalid end time format", http.StatusBadRequest)
			return
		}
	} else {
		endTime = time.Now()
	}

	history := h.metricManager.GetHistory(metricKey, startTime, endTime)
	json.NewEncoder(w).Encode(history)
}

func (h *APIHandler) GetMetricTrend(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	metricKey := vars["key"]

	trend := h.metricManager.GetTrend(metricKey)
	if trend == nil {
		http.Error(w, "Metric not found or no data", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(trend)
}

func (h *APIHandler) GetMetricComparison(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	metricKey := vars["key"]

	comparison, exists := h.metricManager.GetComparison(metricKey)
	if !exists {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(comparison)
}

func (h *APIHandler) GetDashboardLayout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	clientID := r.URL.Query().Get("client_id")
	if clientID == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}

	layout, exists := h.layoutManager.GetLayout(clientID)
	if !exists {
		http.Error(w, "Layout not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(layout)
}

func (h *APIHandler) SaveDashboardLayout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var layout common.DashboardLayout
	if err := json.NewDecoder(r.Body).Decode(&layout); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if layout.ClientID == "" {
		http.Error(w, "client_id is required", http.StatusBadRequest)
		return
	}

	success := h.layoutManager.SaveLayout(layout.ClientID, layout.Metrics, layout.Order)
	if !success {
		http.Error(w, "Invalid layout data", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
}

func (h *APIHandler) GetDelayStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	currentMS, avgMS, maxMS, p99MS := h.metricManager.GetLatencyStats()
	warning, critical, _ := h.metricManager.GetDelayStatus()

	status := map[string]interface{}{
		"current_ms":   currentMS,
		"average_ms":   avgMS,
		"max_ms":       maxMS,
		"p99_ms":       p99MS,
		"warning":      warning,
		"critical":     critical,
		"warning_threshold_ms":  common.DefaultDelayWarning,
		"critical_threshold_ms": common.DefaultDelayCritical,
	}

	json.NewEncoder(w).Encode(status)
}

func (h *APIHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	alerts := h.metricManager.GetPendingAlerts()
	json.NewEncoder(w).Encode(alerts)
}

func (h *APIHandler) CleanupOldData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	h.metricManager.CleanupOldData()
	json.NewEncoder(w).Encode(map[string]string{"status": "cleaned"})
}

func (h *APIHandler) RunBackgroundTasks(stopChan <-chan struct{}) {
	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	alertTicker := time.NewTicker(100 * time.Millisecond)
	defer alertTicker.Stop()

	for {
		select {
		case <-cleanupTicker.C:
			h.metricManager.CleanupOldData()

		case <-alertTicker.C:
			alerts := h.metricManager.GetPendingAlerts()
			for _, alert := range alerts {
				h.connectionManager.PushAlert(alert)
			}

		case <-stopChan:
			return
		}
	}
}
