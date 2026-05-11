package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"1127-go-fan-inout/pkg/api"
	"1127-go-fan-inout/pkg/fanoutfanin"
)

type Simulator struct {
	Name     string
	Interval time.Duration
	Payload  string
	StopChan chan struct{}
	Channel  chan<- fanoutfanin.DataItem
	Done     sync.WaitGroup
}

type ResultBuffer struct {
	mu    sync.Mutex
	items []fanoutfanin.DataItem
	max   int
}

func (rb *ResultBuffer) Add(item fanoutfanin.DataItem) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.items = append(rb.items, item)
	if len(rb.items) > rb.max {
		rb.items = rb.items[len(rb.items)-rb.max:]
	}
}

func (rb *ResultBuffer) GetAll() []fanoutfanin.DataItem {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	result := make([]fanoutfanin.DataItem, len(rb.items))
	copy(result, rb.items)
	return result
}

var (
	scheduler    *fanoutfanin.Scheduler
	simulators   = make(map[string]*Simulator)
	simulatorsMu sync.RWMutex
	resultBuffer = &ResultBuffer{max: 100}
	server       *http.Server
)

func main() {
	scheduler = fanoutfanin.NewScheduler(fanoutfanin.DefaultConfig())

	for _, w := range scheduler.Workers() {
		go func(worker *fanoutfanin.Worker) {
			for item := range worker.Output() {
				resultBuffer.Add(item)
			}
		}(w)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/source/add", handleAddSource)
	mux.HandleFunc("/api/source/remove", handleRemoveSource)
	mux.HandleFunc("/api/policy", handlePolicy)
	mux.HandleFunc("/api/flow", handleFlow)
	mux.HandleFunc("/api/result", handleResult)

	server = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	simulatorsMu.Lock()
	for name, sim := range simulators {
		close(sim.StopChan)
		delete(simulators, name)
	}
	simulatorsMu.Unlock()

	scheduler.Close()

	log.Println("Server stopped")
}

func handleAddSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.AddSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.IntervalMs <= 0 {
		req.IntervalMs = 1000
	}

	if req.Payload == "" {
		req.Payload = fmt.Sprintf("data-from-%s", req.Name)
	}

	simulatorsMu.Lock()
	if _, exists := simulators[req.Name]; exists {
		simulatorsMu.Unlock()
		json.NewEncoder(w).Encode(api.AddSourceResponse{
			Success: false,
			Message: "source already exists",
		})
		return
	}

	ch, err := scheduler.RegisterSource(req.Name)
	if err != nil {
		simulatorsMu.Unlock()
		json.NewEncoder(w).Encode(api.AddSourceResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	sim := &Simulator{
		Name:     req.Name,
		Interval: time.Duration(req.IntervalMs) * time.Millisecond,
		Payload:  req.Payload,
		StopChan: make(chan struct{}),
		Channel:  ch,
	}
	simulators[req.Name] = sim
	simulatorsMu.Unlock()

	sim.Done.Add(1)
	go runSimulator(sim)

	json.NewEncoder(w).Encode(api.AddSourceResponse{
		Success: true,
		Message: fmt.Sprintf("source %s added", req.Name),
	})
}

func handleRemoveSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.RemoveSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	simulatorsMu.Lock()
	sim, exists := simulators[req.Name]
	if !exists {
		simulatorsMu.Unlock()
		json.NewEncoder(w).Encode(api.RemoveSourceResponse{
			Success: false,
			Message: "source not found",
		})
		return
	}

	close(sim.StopChan)
	sim.Done.Wait()
	delete(simulators, req.Name)
	simulatorsMu.Unlock()

	if err := scheduler.UnregisterSource(req.Name); err != nil {
		json.NewEncoder(w).Encode(api.RemoveSourceResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(api.RemoveSourceResponse{
		Success: true,
		Message: fmt.Sprintf("source %s removed", req.Name),
	})
}

func handlePolicy(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		policy := scheduler.GetPolicy()
		json.NewEncoder(w).Encode(api.SetPolicyResponse{
			Success: true,
			Policy:  api.BackpressurePolicy(policy),
		})
	case http.MethodPost:
		var req api.SetPolicyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := scheduler.SetPolicy(fanoutfanin.BackpressurePolicy(req.Policy))
		if err != nil {
			json.NewEncoder(w).Encode(api.SetPolicyResponse{
				Success: false,
				Message: err.Error(),
			})
			return
		}

		json.NewEncoder(w).Encode(api.SetPolicyResponse{
			Success: true,
			Policy:  req.Policy,
		})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleFlow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := scheduler.SourceStats()
	response := api.FlowResponse{
		Sources: make([]api.FlowStats, 0, len(stats)),
	}

	for _, s := range stats {
		response.Sources = append(response.Sources, api.FlowStats{
			SourceName:     s.Name,
			TotalSent:    s.TotalSent,
			TotalReceived: s.TotalRecv,
			Dropped:      s.Dropped,
			CurrentBacklog: s.Backlog,
			IsActive:     s.Active,
		})
	}

	json.NewEncoder(w).Encode(response)
}

func handleResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	items := resultBuffer.GetAll()
	response := api.ResultResponse{
		Items: make([]api.ResultItem, 0, len(items)),
	}

	for _, item := range items {
		response.Items = append(response.Items, api.ResultItem{
			Source:    item.Source,
			Payload:   item.Payload,
			Timestamp: item.Timestamp,
		})
	}

	json.NewEncoder(w).Encode(response)
}

func runSimulator(sim *Simulator) {
	defer sim.Done.Done()

	ticker := time.NewTicker(sim.Interval)
	defer ticker.Stop()

	counter := 0

	for {
		select {
		case <-sim.StopChan:
			return
		case <-ticker.C:
			counter++
			item := fanoutfanin.DataItem{
				Source:    sim.Name,
				Payload:   fmt.Sprintf("%s-%d", sim.Payload, counter),
				Timestamp: time.Now().UnixNano(),
			}

			select {
			case <-sim.StopChan:
				return
			case sim.Channel <- item:
			}
		}
	}
}
