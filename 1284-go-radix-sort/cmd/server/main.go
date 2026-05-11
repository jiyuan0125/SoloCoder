package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"radixsort/pkg/api"
	"radixsort/pkg/radixsort"
)

const defaultPort = 8511

func main() {
	port := defaultPort

	flagPort := flag.Int("port", 0, "HTTP listen port")
	flag.Parse()

	if *flagPort > 0 {
		port = *flagPort
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			port = p
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/sort", handleSort)
	mux.HandleFunc("/api/sort/struct", handleStructSort)
	mux.HandleFunc("/api/radix", handleRadix)
	mux.HandleFunc("/api/stats", handleStats)
	mux.HandleFunc("/health", handleHealth)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("radixsort server listening on %s (radix=%d)", addr, radixsort.GetRadix())
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleSort(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(api.SortResponse{Success: false, Error: "method not allowed"})
		return
	}
	var req api.SortRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(api.SortResponse{Success: false, Error: "invalid request"})
		return
	}
	result := radixsort.SortInt64(req.Data)
	stats := radixsort.GetStats()
	_ = json.NewEncoder(w).Encode(api.SortResponse{
		Success: true,
		Data:    result,
		Stats: api.Stats{
			Passes:          stats.Passes,
			TotalOperations: stats.TotalOperations,
			ArraySize:       stats.ArraySize,
			Radix:           stats.Radix,
		},
	})
}

func handleStructSort(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(api.StructSortResponse{Success: false, Error: "method not allowed"})
		return
	}
	var req api.StructSortRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(api.StructSortResponse{Success: false, Error: "invalid request"})
		return
	}
	fields := make([]radixsort.Int32Field, len(req.Data))
	for i, f := range req.Data {
		fields[i] = radixsort.Int32Field{Name: f.Name, Value: int32(f.Value)}
	}
	sorter := radixsort.NewSorter(radixsort.GetRadix())
	result := sorter.SortStructByField(fields, req.FieldName)
	stats := sorter.GetStats()
	out := make([]api.StructField, len(result))
	for i, f := range result {
		out[i] = api.StructField{Name: f.Name, Value: int64(f.Value)}
	}
	_ = json.NewEncoder(w).Encode(api.StructSortResponse{
		Success: true,
		Data:    out,
		Stats: api.Stats{
			Passes:          stats.Passes,
			TotalOperations: stats.TotalOperations,
			ArraySize:       stats.ArraySize,
			Radix:           stats.Radix,
		},
	})
}

func handleRadix(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		_ = json.NewEncoder(w).Encode(api.RadixResponse{
			Success: true,
			Radix:   radixsort.GetRadix(),
		})
	case http.MethodPost:
		var req api.RadixRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.RadixResponse{Success: false, Error: "invalid request"})
			return
		}
		if !isPowerOfTwo(req.Radix) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.RadixResponse{Success: false, Error: "radix must be a power of 2 (e.g. 2, 4, 8, 16, 32, 64, 128, 256)"})
			return
		}
		radixsort.SetRadix(req.Radix)
		_ = json.NewEncoder(w).Encode(api.RadixResponse{
			Success: true,
			Radix:   radixsort.GetRadix(),
		})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(api.RadixResponse{Success: false, Error: "method not allowed"})
	}
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(api.StatsResponse{Success: false, Error: "method not allowed"})
		return
	}
	stats := radixsort.GetStats()
	_ = json.NewEncoder(w).Encode(api.StatsResponse{
		Success: true,
		Stats: api.Stats{
			Passes:          stats.Passes,
			TotalOperations: stats.TotalOperations,
			ArraySize:       stats.ArraySize,
			Radix:           stats.Radix,
		},
	})
}

func isPowerOfTwo(n int) bool {
	if n < 2 {
		return false
	}
	return (n & (n - 1)) == 0
}

func parsePort(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	p, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return p
}
