package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"sharded-map/common"
	"sharded-map/shardedmap"
)

const DefaultPort = 8217

func main() {
	port := getPort()

	sm := shardedmap.NewShardedMap(shardedmap.DefaultShardCount)

	http.HandleFunc("/put", makePutHandler(sm))
	http.HandleFunc("/get", makeGetHandler(sm))
	http.HandleFunc("/delete", makeDeleteHandler(sm))
	http.HandleFunc("/keys", makeKeysHandler(sm))
	http.HandleFunc("/stats", makeStatsHandler(sm))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
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

func makePutHandler(sm *shardedmap.ShardedMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req common.PutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		sm.Put(req.Key, req.Value)
		w.WriteHeader(http.StatusOK)
	}
}

func makeGetHandler(sm *shardedmap.ShardedMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		key := r.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "Missing 'key' query parameter", http.StatusBadRequest)
			return
		}

		value, found := sm.Get(key)
		resp := common.GetResponse{
			Found: found,
			Value: value,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func makeDeleteHandler(sm *shardedmap.ShardedMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req common.DeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		sm.Delete(req.Key)
		w.WriteHeader(http.StatusOK)
	}
}

func makeKeysHandler(sm *shardedmap.ShardedMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		keys := sm.Keys()
		resp := common.KeysResponse{
			Keys: keys,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

func makeStatsHandler(sm *shardedmap.ShardedMap) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		stats := sm.Stats()
		shardStats := make([]common.ShardStats, len(stats))
		total := 0
		for i, s := range stats {
			shardStats[i] = common.ShardStats{
				ShardID:  s.ShardID,
				Count:    s.Count,
				IsUneven: s.IsUneven,
			}
			total += s.Count
		}

		resp := common.StatsResponse{
			TotalCount: total,
			ShardCount: len(stats),
			Shards:     shardStats,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
