package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"obd-platform/pkg/api"
	"obd-platform/pkg/core"
)

type Server struct {
	service *core.Service
}

func NewServer() *Server {
	return &Server{
		service: core.NewService(nil),
	}
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.UploadDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := s.service.UploadData(&req)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, resp)
}

func (s *Server) handleVehicleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		sendError(w, http.StatusBadRequest, "device_id is required")
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			sendError(w, http.StatusBadRequest, "invalid start date, format: YYYY-MM-DD")
			return
		}
	} else {
		start = time.Now().AddDate(0, 0, -7)
	}

	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			sendError(w, http.StatusBadRequest, "invalid end date, format: YYYY-MM-DD")
			return
		}
	} else {
		end = time.Now()
	}

	resp, err := s.service.GetVehicleDailyStats(deviceID, start, end)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, resp)
}

func (s *Server) handleFleetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	var err error

	if startStr != "" {
		start, err = time.Parse("2006-01-02", startStr)
		if err != nil {
			sendError(w, http.StatusBadRequest, "invalid start date, format: YYYY-MM-DD")
			return
		}
	} else {
		start = time.Now().AddDate(0, 0, -7)
	}

	if endStr != "" {
		end, err = time.Parse("2006-01-02", endStr)
		if err != nil {
			sendError(w, http.StatusBadRequest, "invalid end date, format: YYYY-MM-DD")
			return
		}
	} else {
		end = time.Now()
	}

	resp, err := s.service.GetFleetDailyStats(start, end)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, resp)
}

func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	yearMonth := r.URL.Query().Get("year_month")
	if yearMonth == "" {
		sendError(w, http.StatusBadRequest, "year_month is required, format: YYYY-MM")
		return
	}

	data, err := s.service.ExportCSV(yearMonth)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"fuel_report_"+yearMonth+".csv\"")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *Server) handleGenerateAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	alerts, err := s.service.GenerateAlerts()
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    alerts,
		"count":   len(alerts),
	})
}

func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		deviceID := r.URL.Query().Get("device_id")
		status := r.URL.Query().Get("status")

		alerts, err := s.service.GetAlerts(deviceID, status)
		if err != nil {
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    alerts,
		})
		return
	}

	if r.Method == http.MethodPost {
		var req api.AlertResolveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if err := s.service.ResolveAlerts(req.AlertIDs); err != nil {
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		sendJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "alerts resolved",
		})
		return
	}

	sendError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	sendJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "obd-platform-server",
	})
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, api.ErrorResponse{
		Success: false,
		Error:   message,
	})
}

func getPort() int {
	portStr := os.Getenv("OBD_SERVER_PORT")
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err == nil && port > 0 && port < 65536 {
			return port
		}
	}

	port := 8080
	flag.IntVar(&port, "port", 8080, "server port (also from OBD_SERVER_PORT env)")
	flag.Parse()

	return port
}

func main() {
	port := getPort()
	server := NewServer()

	http.HandleFunc("/health", server.handleHealth)
	http.HandleFunc("/upload", server.handleUpload)
	http.HandleFunc("/vehicle/stats", server.handleVehicleStats)
	http.HandleFunc("/fleet/stats", server.handleFleetStats)
	http.HandleFunc("/export/csv", server.handleExportCSV)
	http.HandleFunc("/alerts/generate", server.handleGenerateAlerts)
	http.HandleFunc("/alerts", server.handleAlerts)

	addr := ":" + strconv.Itoa(port)
	println("Server starting on", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		println("Server error:", err.Error())
		os.Exit(1)
	}
}
