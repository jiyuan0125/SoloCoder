package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"topk-service/pkg/api"
	"topk-service/pkg/topk"
)

type Server struct {
	service *topk.TopKService
}

func NewServer(k int) *Server {
	return &Server{
		service: topk.NewTopKService(k),
	}
}

func (s *Server) submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req api.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	s.service.Add(req.Values)
	
	resp := api.SubmitResponse{
		Code:    0,
		Message: "Success",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) queryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	data := s.service.QueryWithHeap()
	
	resp := api.QueryResponse{
		Code:    0,
		Message: "Success",
		Data:    data,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) clearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	s.service.Clear()
	
	resp := api.ClearResponse{
		Code:    0,
		Message: "Success",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) adjustKHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req api.AdjustKRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	s.service.SetK(req.K)
	
	resp := api.AdjustKResponse{
		Code:    0,
		Message: "Success",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getPort() string {
	port := flag.String("port", "", "Server port")
	flag.Parse()
	
	if *port != "" {
		return *port
	}
	
	if envPort := os.Getenv("TOPK_PORT"); envPort != "" {
		return envPort
	}
	
	return "8509"
}

func getK() int {
	kStr := os.Getenv("TOPK_K")
	if kStr == "" {
		return 10
	}
	
	k, err := strconv.Atoi(kStr)
	if err != nil || k < 0 {
		return 10
	}
	
	return k
}

func main() {
	port := getPort()
	k := getK()
	
	server := NewServer(k)
	
	http.HandleFunc("/submit", server.submitHandler)
	http.HandleFunc("/query", server.queryHandler)
	http.HandleFunc("/clear", server.clearHandler)
	http.HandleFunc("/adjust-k", server.adjustKHandler)
	
	addr := fmt.Sprintf(":%s", port)
	log.Printf("TopK server starting on port %s with K=%d...", port, k)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
