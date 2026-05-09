package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"leaf-segment/pkg/common"
	"leaf-segment/pkg/leaf"
)

var allocator *leaf.Allocator

func main() {
	port := flag.Int("port", 8080, "HTTP port to listen on")
	businessKey := flag.String("business", "default", "Business key for ID generation")
	centerURL := flag.String("center", "", "Center node URL (e.g. http://localhost:9000). If empty, uses mock center.")
	segmentSize := flag.Int64("segment-size", 1000, "Segment size for mock center")
	preloadRatio := flag.Float64("preload-ratio", 0.1, "Preload threshold ratio (0.0-1.0)")
	maxRetries := flag.Int("retries", 3, "Max retries for center node requests")
	retryInterval := flag.Int("retry-interval", 1000, "Retry interval in milliseconds")
	waitTimeout := flag.Int("wait-timeout", 30000, "Wait timeout in milliseconds")

	flag.Parse()

	config := leaf.DefaultConfig(*businessKey, *centerURL)
	config.PreloadThresholdRatio = *preloadRatio
	config.MaxRetries = *maxRetries
	config.RetryInterval = toDuration(*retryInterval)
	config.WaitTimeout = toDuration(*waitTimeout)

	var centerClient leaf.CenterClient
	if *centerURL == "" {
		log.Println("Using mock center client with segment size:", *segmentSize)
		centerClient = leaf.NewMockCenterClient(*segmentSize)
	} else {
		log.Println("Using HTTP center client at:", *centerURL)
		centerClient = leaf.NewHTTPCenterClient(config)
	}

	var err error
	allocator, err = leaf.NewAllocator(config, centerClient)
	if err != nil {
		log.Fatal("Failed to create allocator:", err)
	}

	http.HandleFunc("/id/allocate", handleAllocate)
	http.HandleFunc("/id/batch", handleBatch)
	http.HandleFunc("/id/status", handleStatus)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Server starting on %s with business key: %s", addr, *businessKey)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func toDuration(ms int) time.Duration {
	return time.Duration(ms) * time.Millisecond
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func handleAllocate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{Error: "method not allowed"})
		return
	}
	id, err := allocator.Allocate()
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, common.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, common.AllocateResponse{ID: id})
}

func handleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{Error: "method not allowed"})
		return
	}
	var req common.AllocateBatchRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: err.Error()})
			return
		}
	} else if countStr := r.URL.Query().Get("count"); countStr != "" {
		var count int
		_, err := fmt.Sscanf(countStr, "%d", &count)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "invalid count parameter"})
			return
		}
		req.Count = count
	}
	if req.Count <= 0 {
		writeJSON(w, http.StatusBadRequest, common.ErrorResponse{Error: "count must be positive"})
		return
	}
	ids, err := allocator.AllocateBatch(req.Count)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, common.ErrorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, common.AllocateBatchResponse{IDs: ids})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.ErrorResponse{Error: "method not allowed"})
		return
	}
	status := allocator.Status()
	writeJSON(w, http.StatusOK, status)
}
