package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"api-gateway/backend"
	"api-gateway/logger"
	"api-gateway/router"
)

type AdminServer struct {
	router       *router.Router
	backendMgr   *backend.BackendManager
	accessLogger *logger.AccessLogger
	mux          *http.ServeMux
}

func NewAdminServer(r *router.Router, bm *backend.BackendManager, al *logger.AccessLogger) *AdminServer {
	as := &AdminServer{
		router:       r,
		backendMgr:   bm,
		accessLogger: al,
		mux:          http.NewServeMux(),
	}

	as.mux.HandleFunc("/admin/routes", as.handleRoutes)
	as.mux.HandleFunc("/admin/routes/add", as.handleAddRoute)
	as.mux.HandleFunc("/admin/routes/remove", as.handleRemoveRoute)
	as.mux.HandleFunc("/admin/backends", as.handleBackends)
	as.mux.HandleFunc("/admin/backends/add", as.handleAddBackend)
	as.mux.HandleFunc("/admin/backends/remove", as.handleRemoveBackend)
	as.mux.HandleFunc("/admin/logs", as.handleLogs)

	return as
}

func (as *AdminServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	as.mux.ServeHTTP(w, r)
}

func (as *AdminServer) handleRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	routes := as.router.ListRoutes()
	type RouteResponse struct {
		Path        string              `json:"path"`
		MatchType   router.RouteMatchType `json:"match_type"`
		BackendHost string              `json:"backend_host"`
		BackendPort int                 `json:"backend_port"`
		Timeout     int                 `json:"timeout"`
	}

	response := make([]RouteResponse, len(routes))
	for i, rt := range routes {
		response[i] = RouteResponse{
			Path:        rt.Path,
			MatchType:   rt.MatchType,
			BackendHost: rt.Backend.Host,
			BackendPort: rt.Backend.Port,
			Timeout:     rt.Timeout,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (as *AdminServer) handleAddRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path        string `json:"path"`
		BackendHost string `json:"backend_host"`
		BackendPort int    `json:"backend_port"`
		Timeout     int    `json:"timeout"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Path == "" || req.BackendHost == "" || req.BackendPort <= 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	as.backendMgr.AddBackend(req.BackendHost, req.BackendPort)
	as.router.AddRoute(req.Path, router.BackendRef{
		Host: req.BackendHost,
		Port: req.BackendPort,
	}, req.Timeout)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Route added: %s", req.Path)
}

func (as *AdminServer) handleRemoveRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Path string `json:"path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Path == "" {
		http.Error(w, "Missing path", http.StatusBadRequest)
		return
	}

	as.router.RemoveRoute(req.Path)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Route removed: %s", req.Path)
}

func (as *AdminServer) handleBackends(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	backends := as.backendMgr.ListBackends()
	type BackendResponse struct {
		Host   string             `json:"host"`
		Port   int                `json:"port"`
		Status backend.BackendStatus `json:"status"`
	}

	response := make([]BackendResponse, len(backends))
	for i, b := range backends {
		response[i] = BackendResponse{
			Host:   b.Host,
			Port:   b.Port,
			Status: b.GetStatus(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (as *AdminServer) handleAddBackend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Host == "" || req.Port <= 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	as.backendMgr.AddBackend(req.Host, req.Port)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Backend added: %s:%d", req.Host, req.Port)
}

func (as *AdminServer) handleRemoveBackend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Host == "" || req.Port <= 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	as.backendMgr.RemoveBackend(req.Host, req.Port)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Backend removed: %s:%d", req.Host, req.Port)
}

func (as *AdminServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count := 100
	if countStr := r.URL.Query().Get("count"); countStr != "" {
		if c, err := strconv.Atoi(countStr); err == nil && c > 0 {
			count = c
		}
	}

	logs := as.accessLogger.Recent(count)
	type LogResponse struct {
		ID          int    `json:"id"`
		Time        string `json:"time"`
		ClientIP    string `json:"client_ip"`
		RequestPath string `json:"request_path"`
		BackendHost string `json:"backend_host"`
		BackendPort int    `json:"backend_port"`
		StatusCode  int    `json:"status_code"`
		Duration    string `json:"duration"`
	}

	response := make([]LogResponse, len(logs))
	for i, log := range logs {
		response[i] = LogResponse{
			ID:          log.ID,
			Time:        log.Time.Format("2006-01-02T15:04:05Z07:00"),
			ClientIP:    log.ClientIP,
			RequestPath: log.RequestPath,
			BackendHost: log.BackendHost,
			BackendPort: log.BackendPort,
			StatusCode:  log.StatusCode,
			Duration:    log.Duration.String(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
