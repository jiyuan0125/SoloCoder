package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"pprof-analyzer/internal/apimodels"
	"pprof-analyzer/internal/profiled"
)

type historyStore struct {
	mu      sync.RWMutex
	records []apimodels.HistoryRecord
}

var store = &historyStore{}

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = &p
		}
	}

	http.HandleFunc("/analyze", handleAnalyze)
	http.HandleFunc("/history", handleHistory)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("pprof analyzer server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	result, err := profiled.Analyze(file)
	if err != nil {
		http.Error(w, "failed to analyze profile: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := convertToAPIResponse(result)

	record := apimodels.HistoryRecord{
		ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
		FileName:   header.Filename,
		ProfileType: response.ProfileType,
		TotalBytes: response.TotalBytes,
		AnalyzedAt: time.Now().Format(time.RFC3339),
	}

	store.mu.Lock()
	store.records = append([]apimodels.HistoryRecord{record}, store.records...)
	if len(store.records) > 100 {
		store.records = store.records[:100]
	}
	store.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	store.mu.RLock()
	records := make([]apimodels.HistoryRecord, len(store.records))
	copy(records, store.records)
	store.mu.RUnlock()

	response := apimodels.HistoryResponse{
		Records: records,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func convertToAPIResponse(result *profiled.AnalysisResult) apimodels.AnalysisResponse {
	functions := make([]apimodels.FunctionInfo, len(result.Functions))
	for i, f := range result.Functions {
		functions[i] = apimodels.FunctionInfo{
			FullName:    f.FullName,
			PackageName: f.PackageName,
			FuncName:    f.FuncName,
			Bytes:       f.Bytes,
			Objects:     f.Objects,
			Percentage:  f.Percentage,
		}
	}

	return apimodels.AnalysisResponse{
		ProfileType:  apimodels.ProfileType(result.ProfileType),
		TotalBytes:   result.TotalBytes,
		TotalObjects: result.TotalObjects,
		Functions:    functions,
	}
}
