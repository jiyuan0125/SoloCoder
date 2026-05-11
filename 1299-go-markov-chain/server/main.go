package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"markov-chain/common"
	"markov-chain/markov"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Order < 1 {
		writeError(w, http.StatusBadRequest, "order must be at least 1")
		return
	}

	if req.Text == "" {
		writeError(w, http.StatusBadRequest, "text is empty")
		return
	}

	if req.Length < 0 {
		writeError(w, http.StatusBadRequest, "length must be non-negative")
		return
	}

	if req.Length == 0 {
		resp := common.GenerateResponse{
			Success: true,
			Text:    "",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	var chain *markov.Chain
	var err error

	switch req.Granularity {
	case common.GranularityChar:
		chain, err = markov.NewCharChain(req.Order, req.Seed)
	case common.GranularityWord:
		chain, err = markov.NewWordChain(req.Order, req.Seed)
	default:
		writeError(w, http.StatusBadRequest, "invalid granularity: use 'char' or 'word'")
		return
	}

	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := chain.Train(req.Text); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	generated, err := chain.Generate(req.Length)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := common.GenerateResponse{
		Success: true,
		Text:    generated,
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	resp := common.ErrorResponse{
		Success: false,
		Error:   message,
	}
	json.NewEncoder(w).Encode(resp)
}

func getPort() string {
	port := "8080"

	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	flagPort := flag.String("port", "", "port to listen on")
	flag.Parse()

	if *flagPort != "" {
		port = *flagPort
	}

	if _, err := strconv.Atoi(port); err != nil {
		log.Printf("invalid port '%s', using default 8080", port)
		port = "8080"
	}

	return port
}

func main() {
	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/generate", server.handleGenerate)

	port := getPort()
	addr := fmt.Sprintf(":%s", port)

	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
