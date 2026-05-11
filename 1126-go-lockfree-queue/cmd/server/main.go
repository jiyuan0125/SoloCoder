package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-lockfree-queue/api"
	"go-lockfree-queue/queue"
	"log"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	q           *queue.LockFreeQueue
	processed   atomic.Int64
	port        int
	workerStop  chan struct{}
	workerWg    sync.WaitGroup
}

func NewServer(port int) *Server {
	return &Server{
		q:          queue.NewLockFree(),
		port:       port,
		workerStop: make(chan struct{}),
	}
}

func (s *Server) StartWorker() {
	s.workerWg.Add(1)
	go func() {
		defer s.workerWg.Done()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-s.workerStop:
				return
			case <-ticker.C:
				for {
					v, ok := s.q.Dequeue()
					if !ok {
						break
					}
					s.processed.Add(1)
					_ = v
				}
			}
		}
	}()
}

func (s *Server) Stop() {
	close(s.workerStop)
	s.workerWg.Wait()
}

func (s *Server) handleEnqueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.QueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.q.Enqueue(&req)

	resp := api.QueueResponse{Success: true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := api.StatsResponse{
		Depth:          s.q.Depth(),
		CASSuccess:     s.q.CASSuccess(),
		CASFailure:     s.q.CASFailure(),
		AllocsTotal:    s.q.AllocsTotal(),
		PoolHits:       s.q.PoolHits(),
		PoolMisses:     s.q.PoolMisses(),
		ProcessedTotal: s.processed.Load(),
		EnqueueTotal:   s.q.EnqueueTotal(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var q queue.Queue
	if req.UseLock {
		q = queue.NewLocked()
	} else {
		q = queue.NewLockFree()
	}

	var wg sync.WaitGroup
	for i, msg := range req.Messages {
		wg.Add(1)
		go func(idx int, m string) {
			defer wg.Done()
			q.Enqueue(m)
		}(i, msg)
	}
	wg.Wait()

	results := make([]string, 0, len(req.Messages))
	for i := 0; i < len(req.Messages); i++ {
		v, ok := q.Dequeue()
		if !ok {
			break
		}
		results = append(results, v.(string))
	}

	sort.Strings(results)
	expected := make([]string, len(req.Messages))
	copy(expected, req.Messages)
	sort.Strings(expected)

	success := len(results) == len(expected)
	if success {
		for i := range results {
			if results[i] != expected[i] {
				success = false
				break
			}
		}
	}

	message := ""
	if !success {
		message = fmt.Sprintf("mismatch: expected %d items, got %d", len(expected), len(results))
	}

	resp := api.VerifyResponse{
		Success:  success,
		Expected: expected,
		Actual:   results,
		Message:  message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleBench(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.BenchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Producers <= 0 {
		req.Producers = 4
	}
	if req.MessagesPerProducer <= 0 {
		req.MessagesPerProducer = 1000
	}

	benchQ := queue.NewLockFree()
	latencies := make([]float64, 0, req.Producers*req.MessagesPerProducer)
	var latMu sync.Mutex

	start := time.Now()

	var wg sync.WaitGroup
	for p := 0; p < req.Producers; p++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			for m := 0; m < req.MessagesPerProducer; m++ {
				msgStart := time.Now()
				msg := fmt.Sprintf("msg-%d-%d", producerID, m)
				benchQ.Enqueue(msg)
				elapsed := float64(time.Since(msgStart).Nanoseconds()) / 1e6
				latMu.Lock()
				latencies = append(latencies, elapsed)
				latMu.Unlock()
			}
		}(p)
	}
	wg.Wait()

	total := req.Producers * req.MessagesPerProducer
	received := 0
	for received < total {
		_, ok := benchQ.Dequeue()
		if !ok {
			time.Sleep(1 * time.Millisecond)
			continue
		}
		received++
	}

	duration := time.Since(start)
	durationMs := duration.Milliseconds()
	if durationMs == 0 {
		durationMs = 1
	}

	throughput := float64(total) / (float64(durationMs) / 1000.0)

	sort.Float64s(latencies)

	minLat := latencies[0]
	maxLat := latencies[len(latencies)-1]

	sum := 0.0
	for _, l := range latencies {
		sum += l
	}
	avgLat := sum / float64(len(latencies))

	p50 := latencies[len(latencies)*50/100]
	p95 := latencies[len(latencies)*95/100]
	p99 := latencies[len(latencies)*99/100]

	result := &api.BenchResult{
		TotalMessages: int64(total),
		Producers:     req.Producers,
		DurationMs:    durationMs,
		Throughput:    throughput,
		LatencyMin:    minLat,
		LatencyMax:    maxLat,
		LatencyAvg:    avgLat,
		LatencyP50:    p50,
		LatencyP95:    p95,
		LatencyP99:    p99,
	}

	resp := api.BenchResponse{
		Success: true,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/enqueue", s.handleEnqueue)
	mux.HandleFunc("/stats", s.handleStats)
	mux.HandleFunc("/verify", s.handleVerify)
	mux.HandleFunc("/bench", s.handleBench)

	s.StartWorker()

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func main() {
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	server := NewServer(*port)
	defer server.Stop()
	server.Run()
}
