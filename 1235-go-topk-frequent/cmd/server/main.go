package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	"topk-frequent/internal/topk"
	"topk-frequent/pkg/api"
)

var (
	analyzer *topk.Analyzer
	mu       sync.RWMutex
)

func main() {
	port := getPort()

	http.HandleFunc("/create", handleCreate)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/topk", handleTopK)

	fmt.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
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

	return "8216"
}

func sendJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func sendError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, api.ErrorResponse{
		Success: false,
		Message: message,
	})
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	mu.Lock()
	analyzer = topk.New(req.Width, req.Depth, req.K)
	mu.Unlock()

	sendJSON(w, http.StatusOK, api.CreateResponse{
		Success: true,
		Message: "Analyzer created successfully",
	})
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	mu.RLock()
	exists := analyzer != nil
	mu.RUnlock()

	if !exists {
		sendError(w, http.StatusBadRequest, "Analyzer not created. Call /create first")
		return
	}

	var req api.AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	mu.Lock()
	if len(req.Items) == 1 {
		analyzer.Add(req.Items[0])
	} else {
		analyzer.AddBatch(req.Items)
	}
	mu.Unlock()

	sendJSON(w, http.StatusOK, api.AddResponse{
		Success: true,
		Message: fmt.Sprintf("Added %d items", len(req.Items)),
	})
}

func handleTopK(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	mu.RLock()
	exists := analyzer != nil
	mu.RUnlock()

	if !exists {
		sendError(w, http.StatusBadRequest, "Analyzer not created. Call /create first")
		return
	}

	k := 0
	if kStr := r.URL.Query().Get("k"); kStr != "" {
		if parsed, err := strconv.Atoi(kStr); err == nil && parsed > 0 {
			k = parsed
		}
	}

	mu.RLock()
	var items []topk.TopKItem
	if k > 0 {
		items = analyzer.TopKWithK(k)
	} else {
		items = analyzer.TopK()
	}
	mu.RUnlock()

	result := make([]api.TopKItem, len(items))
	for i, item := range items {
		result[i] = api.TopKItem{
			Element:   item.Element,
			Freq:      item.Freq,
			Uncertain: item.Uncertain,
		}
	}

	sendJSON(w, http.StatusOK, api.TopKResponse{
		Success: true,
		Items:   result,
	})
}
