package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"bloomfilter/api"
	"bloomfilter/bloomfilter"
)

var filter *bloomfilter.Filter

func main() {
	capacity := flag.Uint64("capacity", 10000, "Expected number of elements")
	falsePositive := flag.Float64("fp", 0.01, "False positive rate")
	counterBits := flag.Int("bits", 4, "Number of bits per counter (1-8)")
	port := flag.String("port", "8080", "Server port")
	flag.Parse()

	filter = bloomfilter.New(bloomfilter.Config{
		Capacity:      *capacity,
		FalsePositive: *falsePositive,
		CounterBits:   uint8(*counterBits),
	})

	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/delete", handleDelete)
	http.HandleFunc("/query", handleQuery)
	http.HandleFunc("/info", handleInfo)

	log.Printf("Bloom filter server starting on port %s", *port)
	log.Printf("Config: capacity=%d, fp=%f, bits=%d", *capacity, *falsePositive, *counterBits)

	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Item == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Item cannot be empty"})
		return
	}

	filter.Add(req.Item)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.AddResponse{Success: true})
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Item == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Item cannot be empty"})
		return
	}

	filter.Delete(req.Item)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.DeleteResponse{Success: true})
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Invalid request body"})
		return
	}

	if req.Item == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ErrorResponse{Error: "Item cannot be empty"})
		return
	}

	result := filter.MayContain(req.Item)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.QueryResponse{MayContain: result})
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	info := filter.Info()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.InfoResponse{
		NumHashes:       info.NumHashes,
		CounterBits:     info.CounterBits,
		Capacity:        info.Capacity,
		Inserted:       info.Inserted,
		NonZeroCounters: info.NonZeroCounters,
		OverflowCount:  info.OverflowCount,
	})
}
