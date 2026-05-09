package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/solocoder/reed-solomon/pkg/api"
	"github.com/solocoder/reed-solomon/pkg/rs"
)

type storedShard struct {
	ID     string
	Index  int
	Data   []byte
	IsLost bool
}

type storedBlock struct {
	BlockID     string
	N           int
	M           int
	Shards      []storedShard
	OriginalLen int
}

var (
	store sync.Map
)

func generateID() string {
	return uuid.New().String()
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func encodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.EncodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	data := []byte(req.Data)
	shards, originalLen, err := rs.Encode(data, req.N, req.M)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	blockID := generateID()
	storedShards := make([]storedShard, len(shards))
	shardInfos := make([]api.ShardInfo, len(shards))

	for i, s := range shards {
		shardID := generateID()
		storedShards[i] = storedShard{
			ID:    shardID,
			Index:   s.Index,
			Data:    s.Data,
			IsLost:  false,
		}
		shardInfos[i] = api.ShardInfo{
			ID:    shardID,
			Index:   s.Index,
			IsLost:  false,
		}
	}

	block := storedBlock{
		BlockID:     blockID,
		N:           req.N,
		M:           req.M,
		Shards:      storedShards,
		OriginalLen: originalLen,
	}

	store.Store(blockID, block)

	resp := api.EncodeResponse{
		BlockID:     blockID,
		N:           req.N,
		M:           req.M,
		Shards:      shardInfos,
		OriginalLen: originalLen,
	}

	writeJSON(w, http.StatusOK, resp)
}

func decodeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.DecodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	blockVal, ok := store.Load(req.BlockID)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Errorf("block not found"))
		return
	}
	block := blockVal.(storedBlock)

	shardMap := make(map[string]storedShard)
	for _, s := range block.Shards {
		shardMap[s.ID] = s
	}

	var availableShards []rs.Shard
	for _, sid := range req.ShardIDs {
		shard, exists := shardMap[sid]
		if !exists {
			continue
		}
		if shard.IsLost {
			continue
		}
		availableShards = append(availableShards, rs.Shard{
			Index: shard.Index,
			Data:  shard.Data,
		})
	}

	if len(availableShards) < block.N {
		resp := api.DecodeResponse{
			Success: false,
			Error:   fmt.Sprintf("insufficient shards: need %d, have %d", block.N, len(availableShards)),
		}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	data, err := rs.Decode(availableShards, block.N, block.M, block.OriginalLen)
	if err != nil {
		resp := api.DecodeResponse{
			Success: false,
			Error:   err.Error(),
		}
		writeJSON(w, http.StatusOK, resp)
		return
	}

	resp := api.DecodeResponse{
		Success: true,
		Data:    string(data),
	}
	writeJSON(w, http.StatusOK, resp)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var blocks []api.BlockStatus

	store.Range(func(key, value interface{}) bool {
		block := value.(storedBlock)
		shardStatuses := make([]api.ShardStatus, len(block.Shards))
		availableCount := 0

		for i, s := range block.Shards {
			isAvailable := !s.IsLost
			if isAvailable {
				availableCount++
			}
			shardStatuses[i] = api.ShardStatus{
				ShardID:     s.ID,
				Index:       s.Index,
				IsAvailable: isAvailable,
			}
		}

		blocks = append(blocks, api.BlockStatus{
			BlockID:       block.BlockID,
			N:             block.N,
			M:             block.M,
			TotalShards:   len(block.Shards),
			AvailableShards: availableCount,
			Shards:        shardStatuses,
		})
		return true
	})

	resp := api.StatusResponse{
		Blocks: blocks,
	}
	writeJSON(w, http.StatusOK, resp)
}

func markLostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.MarkLostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid request body"))
		return
	}

	blockVal, ok := store.Load(req.BlockID)
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Errorf("block not found"))
		return
	}
	block := blockVal.(storedBlock)

	found := false
	for i := range block.Shards {
		if block.Shards[i].ID == req.ShardID {
			block.Shards[i].IsLost = true
			found = true
			break
		}
	}

	if !found {
		writeError(w, http.StatusNotFound, fmt.Errorf("shard not found"))
		return
	}

	store.Store(req.BlockID, block)

	resp := api.MarkLostResponse{
		Success: true,
		Message: fmt.Sprintf("shard %s marked as lost", req.ShardID),
	}
	writeJSON(w, http.StatusOK, resp)
}

func main() {
	http.HandleFunc("/api/encode", encodeHandler)
	http.HandleFunc("/api/decode", decodeHandler)
	http.HandleFunc("/api/status", statusHandler)
	http.HandleFunc("/api/mark-lost", markLostHandler)

	port := ":8080"
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
