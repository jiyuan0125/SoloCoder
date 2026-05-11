package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	"sliding-median/internal/median"
	"sliding-median/internal/models"
)

type Server struct {
	sm     *median.SlidingMedian
	mu     sync.Mutex
	port   string
}

func NewServer(windowSize int, port string) *Server {
	return &Server{
		sm:   median.NewSlidingMedian(windowSize),
		port: port,
	}
}

func (s *Server) handlePush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.PushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.sm.Push(req.Value)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.PushResponse{
		Success: true,
	})
}

func (s *Server) handleMedian(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	median, ok := s.sm.Median()
	s.mu.Unlock()

	var resp models.MedianResponse
	if ok {
		resp.Success = true
		resp.Median = &median
	} else {
		resp.Success = false
		resp.Message = "No data available"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleSetWindowSize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SetWindowSizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Size <= 0 {
		http.Error(w, "Window size must be positive", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.sm.SetWindowSize(req.Size)
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.SetWindowSizeResponse{
		Success: true,
	})
}

func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	s.sm.Reset()
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ResetResponse{
		Success: true,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	windowSize := s.sm.WindowSize()
	dataCount := s.sm.DataCount()
	medianVal, ok := s.sm.Median()
	s.mu.Unlock()

	resp := models.StatusResponse{
		Success:    true,
		WindowSize: windowSize,
		DataCount:  dataCount,
	}
	if ok {
		resp.Median = &medianVal
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getPort() string {
	port := flag.String("port", "", "Server port (e.g., 8080)")
	flag.Parse()

	if *port != "" {
		return ":" + *port
	}

	envPort := os.Getenv("PORT")
	if envPort != "" {
		return ":" + envPort
	}

	return ":8508"
}

func getWindowSize() int {
	sizeStr := os.Getenv("WINDOW_SIZE")
	if sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			return size
		}
	}
	return 10
}

func main() {
	port := getPort()
	windowSize := getWindowSize()

	server := NewServer(windowSize, port)

	http.HandleFunc("/push", server.handlePush)
	http.HandleFunc("/median", server.handleMedian)
	http.HandleFunc("/window-size", server.handleSetWindowSize)
	http.HandleFunc("/reset", server.handleReset)
	http.HandleFunc("/status", server.handleStatus)

	fmt.Printf("Server starting on port %s (window size: %d)\n", port, windowSize)
	if err := http.ListenAndServe(server.port, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}
