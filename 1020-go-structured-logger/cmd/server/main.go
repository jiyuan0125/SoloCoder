package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"structured-logger/pkg/api"
	"structured-logger/pkg/logger"
)

type Server struct {
	log    *logger.Logger
	addr   string
}

func NewServer(log *logger.Logger, addr string) *Server {
	return &Server{
		log:  log,
		addr: addr,
	}
}

func (s *Server) writeLogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req api.WriteLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.WriteLogResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}
	level := req.Level
	if level == "" {
		level = logger.LevelInfo
	}
	s.log.Log(level, req.Message, req.Fields)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.WriteLogResponse{Success: true})
}

func (s *Server) levelHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		s.updateLevelHandler(w, r)
	case http.MethodGet:
		s.getLevelHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) updateLevelHandler(w http.ResponseWriter, r *http.Request) {
	var req api.UpdateLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.UpdateLevelResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}
	level := strings.ToUpper(string(req.Level))
	switch level {
	case "DEBUG", "INFO", "WARN", "ERROR":
		s.log.SetLevel(logger.LogLevel(level))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.UpdateLevelResponse{Success: true})
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.UpdateLevelResponse{
			Success: false,
			Error:   "Invalid level: " + level,
		})
	}
}

func (s *Server) getLevelHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"level": string(s.log.GetLevel())})
}

func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	config := s.log.GetConfig()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

func (s *Server) recentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	nStr := r.URL.Query().Get("n")
	n := 100
	if nStr != "" {
		if parsed, err := strconv.Atoi(nStr); err == nil && parsed > 0 {
			n = parsed
		}
	}
	logs := s.log.GetRecentLogs(n)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.GetRecentLogsResponse{
		Logs:    logs,
		Success: true,
	})
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/log/write", s.writeLogHandler)
	mux.HandleFunc("/log/level", s.levelHandler)
	mux.HandleFunc("/log/config", s.configHandler)
	mux.HandleFunc("/log/recent", s.recentHandler)
	fmt.Printf("Server starting on %s\n", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func main() {
	addr := flag.String("addr", ":8080", "Server address")
	level := flag.String("level", "INFO", "Initial log level (DEBUG, INFO, WARN, ERROR)")
	outputFile := flag.String("file", "", "Output log file path (enables file output)")
	maxSize := flag.Int64("max-size", logger.DefaultMaxFileSize, "Max file size in bytes")
	maxBackups := flag.Int("max-backups", logger.DefaultMaxBackups, "Max number of backup files")
	flag.Parse()

	opts := logger.LoggerOptions{
		Level:       logger.ParseLevel(*level),
		OutputToStd: true,
		MaxFileSize: *maxSize,
		MaxBackups:  *maxBackups,
	}
	if *outputFile != "" {
		opts.OutputToFile = true
		opts.FilePath = *outputFile
	}
	log, err := logger.New(opts)
	if err != nil {
		fmt.Printf("Failed to create logger: %v\n", err)
		return
	}
	server := NewServer(log, *addr)
	log.Info("Log server initialized")
	if err := server.Run(); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
