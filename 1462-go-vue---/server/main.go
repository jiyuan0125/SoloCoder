package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"energymanagement/common"
	"energymanagement/core"
)

type Server struct {
	service *core.Service
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "server port")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	srv := &Server{
		service: core.NewService(),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/points", srv.handlePoints)
	mux.HandleFunc("/api/data/report", srv.handleReportData)
	mux.HandleFunc("/api/data/point", srv.handlePointData)
	mux.HandleFunc("/api/analysis/point", srv.handlePointConsumption)
	mux.HandleFunc("/api/analysis/area", srv.handleAreaConsumption)
	mux.HandleFunc("/api/analysis/area-comparison", srv.handleAreaComparison)
	mux.HandleFunc("/api/overview", srv.handleOverview)
	mux.HandleFunc("/api/suggestions", srv.handleSuggestions)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server starting on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func (s *Server) handlePoints(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		points := s.service.ListPoints()
		sendSuccess(w, points)
	case http.MethodPost:
		var req common.CreatePointRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		point, err := s.service.CreatePoint(req)
		if err != nil {
			sendError(w, http.StatusBadRequest, err.Error())
			return
		}
		sendSuccess(w, point)
	default:
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleReportData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.ReportDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	data, err := s.service.ReportData(req)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	sendSuccess(w, data)
}

func (s *Server) handlePointData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	pointID := r.URL.Query().Get("point_id")
	if pointID == "" {
		sendError(w, http.StatusBadRequest, "point_id is required")
		return
	}

	start := parseTime(r.URL.Query().Get("start"), time.Now().AddDate(0, 0, -7))
	end := parseTime(r.URL.Query().Get("end"), time.Now())

	data := s.service.GetPointData(pointID, start, end)
	sendSuccess(w, data)
}

func (s *Server) handlePointConsumption(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.PointTimeQuery
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result := s.service.GetPointConsumption(req.PointID, req.Start, req.End, req.Granularity)
	sendSuccess(w, result)
}

func (s *Server) handleAreaConsumption(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.AreaTimeQuery
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result := s.service.GetAreaConsumption(req.Area, req.Start, req.End, req.Granularity)
	sendSuccess(w, result)
}

func (s *Server) handleAreaComparison(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.TimeRangeQuery
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result := s.service.GetAreaComparison(req.Start, req.End)
	sendSuccess(w, result)
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	refTime := parseTime(r.URL.Query().Get("time"), time.Now())
	metrics := s.service.GetOverviewMetrics(refTime)
	sendSuccess(w, metrics)
}

func (s *Server) handleSuggestions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	suggestions := s.service.GetSuggestions()
	sendSuccess(w, suggestions)
}

func parseTime(s string, defaultTime time.Time) time.Time {
	if s == "" {
		return defaultTime
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	return defaultTime
}

func sendSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.Response{
		Success: true,
		Data:    data,
	})
}

func sendError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.Response{
		Success: false,
		Error:   msg,
	})
}
