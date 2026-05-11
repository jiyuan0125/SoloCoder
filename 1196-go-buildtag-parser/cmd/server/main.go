package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"buildtag/api"
	"buildtag/buildtag"
)

type Server struct {
	analyzer *buildtag.Analyzer
	port     string
}

func NewServer(port string) *Server {
	return &Server{
		analyzer: buildtag.NewAnalyzer(),
		port:     port,
	}
}

func (s *Server) analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req api.AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	
	response := s.analyzer.Analyze(req)
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/analyze", s.analyzeHandler)
	mux.HandleFunc("/health", s.healthHandler)
	
	addr := ":" + s.port
	fmt.Printf("Server starting on port %s...\n", s.port)
	return http.ListenAndServe(addr, mux)
}

func getPort() string {
	port := flag.String("port", "", "Port to listen on")
	flag.Parse()
	
	if *port != "" {
		return *port
	}
	
	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}
	
	return "8400"
}

func main() {
	port := getPort()
	server := NewServer(port)
	
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
