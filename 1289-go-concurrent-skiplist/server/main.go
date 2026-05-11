package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/example/concurrent-skiplist/common"
	"github.com/example/concurrent-skiplist/skiplist"
)

const DefaultPort = 8516

type Server struct {
	manager *skiplist.SkipListManager
}

func NewServer(strategy skiplist.ConcurrentStrategy) *Server {
	return &Server{
		manager: skiplist.NewSkipListManager(strategy, skiplist.DefaultMaxLevel),
	}
}

func (s *Server) InsertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.InsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	success := s.manager.Insert(req.Key, req.Value)

	resp := common.InsertResponse{
		Success: success,
	}
	if !success {
		resp.Message = "Insert failed"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	value, found := s.manager.Get(key)

	resp := common.GetResponse{
		Success: found,
	}
	if found {
		resp.Value = value
	} else {
		resp.Message = "Key not found"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	success := s.manager.Delete(key)

	resp := common.DeleteResponse{
		Success: success,
	}
	if !success {
		resp.Message = "Key not found"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) RangeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")

	results := s.manager.Range(start, end)

	data := make([]common.KeyValue, len(results))
	for i, kv := range results {
		data[i] = common.KeyValue{
			Key:   kv.Key,
			Value: kv.Value,
		}
	}

	resp := common.RangeQueryResponse{
		Success: true,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) SwitchStrategyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SwitchStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var strategy skiplist.ConcurrentStrategy
	switch strings.ToLower(req.Strategy) {
	case "global_lock", "global":
		strategy = skiplist.StrategyGlobalLock
	case "node_level_lock", "node":
		strategy = skiplist.StrategyNodeLevelLock
	default:
		resp := common.SwitchStrategyResponse{
			Success: false,
			Message: "Invalid strategy. Use 'global_lock' or 'node_level_lock'",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	s.manager.SetStrategy(strategy)

	resp := common.SwitchStrategyResponse{
		Success: true,
		Message: fmt.Sprintf("Switched to %s strategy", strategy.String()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) GetStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.manager.GetStats()
	strategy := s.manager.GetStrategy()

	resp := common.GetStatsResponse{
		Success: true,
		Stats: common.PerformanceStats{
			TotalOperations:   stats.TotalOperations,
			SuccessOperations: stats.SuccessOperations,
			FailedOperations:  stats.FailedOperations,
			ConflictCount:     stats.ConflictCount,
			TotalWaitTime:     stats.TotalWaitTime,
			AvgWaitTime:       stats.AvgWaitTime,
			InsertCount:       stats.InsertCount,
			DeleteCount:       stats.DeleteCount,
			GetCount:          stats.GetCount,
			RangeCount:        stats.RangeCount,
			CurrentStrategy:   strategy.String(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) StressTestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.StressTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Operations <= 0 {
		req.Operations = 10000
	}
	if req.Concurrency <= 0 {
		req.Concurrency = 10
	}

	startTime := time.Now()

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, req.Concurrency)
	localRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < req.Operations; i++ {
		semaphore <- struct{}{}
		wg.Add(1)
		
		go func(opId int) {
			defer wg.Done()
			defer func() { <-semaphore }()
			
			opType := localRand.Intn(3)
			key := fmt.Sprintf("key-%d", localRand.Intn(req.Operations))
			
			switch opType {
			case 0:
				s.manager.Insert(key, fmt.Sprintf("value-%d", opId))
			case 1:
				s.manager.Get(key)
			case 2:
				s.manager.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	duration := time.Since(startTime)
	durationMs := duration.Milliseconds()
	throughput := float64(req.Operations) / duration.Seconds()

	resp := common.StressTestResponse{
		Success:    true,
		DurationMs: durationMs,
		Operations: req.Operations,
		Throughput: throughput,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getPort() int {
	portFlag := flag.Int("port", 0, "Server port")
	flag.Parse()

	if *portFlag > 0 {
		return *portFlag
	}

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			return p
		}
	}

	return DefaultPort
}

func main() {
	port := getPort()
	server := NewServer(skiplist.StrategyGlobalLock)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/insert", server.InsertHandler)
	mux.HandleFunc("/api/get", server.GetHandler)
	mux.HandleFunc("/api/delete", server.DeleteHandler)
	mux.HandleFunc("/api/range", server.RangeHandler)
	mux.HandleFunc("/api/strategy", server.SwitchStrategyHandler)
	mux.HandleFunc("/api/stats", server.GetStatsHandler)
	mux.HandleFunc("/api/stress", server.StressTestHandler)

	log.Printf("Server starting on port %d...", port)
	log.Printf("Available endpoints:")
	log.Printf("  POST /api/insert - Insert key-value pair")
	log.Printf("  GET  /api/get    - Get value by key")
	log.Printf("  DELETE /api/delete - Delete key")
	log.Printf("  GET  /api/range  - Range query [start, end]")
	log.Printf("  POST /api/strategy - Switch concurrency strategy")
	log.Printf("  GET  /api/stats  - Get performance statistics")
	log.Printf("  POST /api/stress - Run stress test")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
