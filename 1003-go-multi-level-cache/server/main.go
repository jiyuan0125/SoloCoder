package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"multilevelcache/cache"
	"multilevelcache/common"
)

var tc *cache.TwoLevelCache

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	memoryCap := flag.Int("memory-cap", 1000, "Memory cache capacity")
	fileDir := flag.String("file-dir", "./file_cache", "File cache directory")
	cleanInterval := flag.Duration("clean-interval", 60*time.Second, "File cache clean interval")
	flag.Parse()

	var err error
	tc, err = cache.NewTwoLevelCache(cache.TwoLevelCacheConfig{
		MemoryCapacity: *memoryCap,
		FileCacheDir:   *fileDir,
		CleanInterval:  *cleanInterval,
	})
	if err != nil {
		log.Fatalf("Failed to create cache: %v", err)
	}
	defer tc.Stop()

	log.Printf("Warming up cache from disk...")
	tc.WarmUp(*memoryCap)
	log.Printf("Warm-up complete")

	http.HandleFunc("/get", handleGet)
	http.HandleFunc("/set", handleSet)
	http.HandleFunc("/delete", handleDelete)
	http.HandleFunc("/clear", handleClear)
	http.HandleFunc("/stats", handleStats)

	log.Printf("Server listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.GetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	value, expire, found := tc.Get(req.Key)
	resp := common.GetResponse{
		Key:    req.Key,
		Value:  value,
		Found:  found,
		Expire: expire,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := tc.Set(req.Key, req.Value, req.MemoryTTL, req.FileTTL)
	resp := common.SetResponse{
		Success: err == nil,
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var deleted int
	if req.Prefix != "" {
		deleted = tc.DeletePrefix(req.Prefix)
	} else {
		if tc.Delete(req.Key) {
			deleted = 1
		}
	}

	resp := common.DeleteResponse{
		Deleted: deleted,
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tc.Clear()
	resp := common.ClearResponse{Success: true}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := tc.Stats()
	memHits := float64(stats.MemoryHits)
	memMisses := float64(stats.MemoryMisses)
	fileHits := float64(stats.FileHits)
	fileMisses := float64(stats.FileMisses)

	totalHits := memHits + fileHits
	total := totalHits + memMisses + fileMisses

	hitRate := 0.0
	if total > 0 {
		hitRate = totalHits / total
	}

	resp := common.StatsResponse{
		MemoryHits:     stats.MemoryHits,
		MemoryMisses:   stats.MemoryMisses,
		FileHits:       stats.FileHits,
		FileMisses:     stats.FileMisses,
		HitRate:        hitRate,
		MemoryCount:    tc.MemoryCount(),
		FileCount:      tc.FileCount(),
		MemoryCapacity: tc.MemoryCapacity(),
		MemoryUsage:    tc.MemoryUsage(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
