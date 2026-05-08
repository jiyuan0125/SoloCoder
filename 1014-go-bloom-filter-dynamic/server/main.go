package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/bloomfilter/bloom"
	"github.com/example/bloomfilter/common"
)

var filter *bloom.Filter
var persistPath string

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	capacity := flag.Uint64("capacity", bloom.DefaultCapacity, "Initial bloom filter capacity")
	fpr := flag.Float64("fpr", bloom.DefaultFPR, "Target false positive rate")
	flag.StringVar(&persistPath, "persist", "", "Path to persist/load bloom filter state")
	flag.Parse()

	var err error
	if persistPath != "" {
		filter, err = loadFromFile(persistPath)
		if err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: failed to load persisted state: %v, creating new filter", err)
		}
	}

	if filter == nil {
		filter, err = bloom.New(*capacity, *fpr)
		if err != nil {
			log.Fatalf("Failed to create bloom filter: %v", err)
		}
	}

	setupSignalHandler()

	http.HandleFunc("/bloom/add", handleAdd)
	http.HandleFunc("/bloom/add-batch", handleAddBatch)
	http.HandleFunc("/bloom/check", handleCheck)
	http.HandleFunc("/bloom/stats", handleStats)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server starting on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func loadFromFile(path string) (*bloom.Filter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return bloom.Deserialize(data)
}

func saveToFile() {
	if persistPath == "" || filter == nil {
		return
	}
	data := filter.Serialize()
	if err := os.WriteFile(persistPath, data, 0644); err != nil {
		log.Printf("Warning: failed to persist state: %v", err)
	}
}

func setupSignalHandler() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down...")
		saveToFile()
		os.Exit(0)
	}()
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req common.AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Item == "" {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "item cannot be empty"})
		return
	}

	filter.Add(req.Item)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleAddBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{Error: "method not allowed"})
		return
	}

	var req common.AddBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "invalid request body"})
		return
	}

	if len(req.Items) == 0 {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "items cannot be empty"})
		return
	}

	filter.AddBatch(req.Items)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{Error: "method not allowed"})
		return
	}

	item := r.URL.Query().Get("item")
	if item == "" {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "item parameter required"})
		return
	}

	exists := filter.Check(item)
	writeJSON(w, http.StatusOK, common.CheckResponse{Exists: exists})
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{Error: "method not allowed"})
		return
	}

	stats := filter.Stats()
	resp := common.StatsResponse{
		Count:      stats.Count,
		Capacity:   stats.Capacity,
		CurrentFPR: stats.FPR,
		BitUsage:   stats.BitUsage,
	}
	writeJSON(w, http.StatusOK, resp)
}
