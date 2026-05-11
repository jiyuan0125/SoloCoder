package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"reservoir-sample/common"
	"reservoir-sample/reservoir"
)

const (
	defaultPort = 8510
	defaultK   = 1000
)

type Server struct {
	sampler *reservoir.Sampler
}

func NewServer(k int) *Server {
	return &Server{
		sampler: reservoir.NewSampler(k),
	}
}

func (s *Server) addDataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.AddDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	s.sampler.Add(req.Data)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.AddDataResponse{Success: true})
}

func (s *Server) sampleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SampleRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
	}

	var samples []string
	if req.K != nil {
		samples = s.sampler.Sample(*req.K)
	} else {
		samples = s.sampler.SampleWithDefault()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.SampleResponse{
		Success: true,
		Samples: samples,
	})
}

func (s *Server) setKHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SetKRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	s.sampler.SetK(req.K)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.SetKResponse{
		Success: true,
		K:       req.K,
	})
}

func (s *Server) countHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count := s.sampler.Count()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.CountResponse{
		Success: true,
		Count:   count,
	})
}

func (s *Server) clearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.sampler.Clear()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.ClearResponse{Success: true})
}

func getPort() int {
	port := defaultPort

	if envPort := os.Getenv("RESERVOIR_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	var flagPort int
	flag.IntVar(&flagPort, "port", 0, "listening port")
	flag.Parse()

	if flagPort > 0 {
		port = flagPort
	}

	return port
}

func main() {
	server := NewServer(defaultK)

	http.HandleFunc("/data", server.addDataHandler)
	http.HandleFunc("/sample", server.sampleHandler)
	http.HandleFunc("/k", server.setKHandler)
	http.HandleFunc("/count", server.countHandler)
	http.HandleFunc("/clear", server.clearHandler)

	port := getPort()
	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("server listening on %s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
