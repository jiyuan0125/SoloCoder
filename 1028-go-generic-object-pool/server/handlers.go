package main

import (
	"encoding/json"
	"generic-object-pool/common"
	"net/http"
	"time"
)

func createPoolHandler(manager *PoolManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req common.CreatePoolRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			json.NewEncoder(w).Encode(common.CreatePoolResponse{
				Success: false,
				Message: "invalid request body",
			})
			return
		}

		if req.PoolID == "" {
			json.NewEncoder(w).Encode(common.CreatePoolResponse{
				Success: false,
				Message: "pool_id is required",
			})
			return
		}

		var idleTimeout time.Duration
		if req.IdleTimeout != "" {
			parsed, err := time.ParseDuration(req.IdleTimeout)
			if err == nil {
				idleTimeout = parsed
			}
		}

		if manager.Create(req.PoolID, req.MaxSize, idleTimeout) {
			json.NewEncoder(w).Encode(common.CreatePoolResponse{
				Success: true,
				Message: "pool created",
			})
		} else {
			json.NewEncoder(w).Encode(common.CreatePoolResponse{
				Success: false,
				Message: "pool already exists",
			})
		}
	}
}

func getObjectHandler(manager *PoolManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req common.GetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			json.NewEncoder(w).Encode(common.GetResponse{
				Success: false,
				Message: "invalid request body",
			})
			return
		}

		if req.PoolID == "" {
			json.NewEncoder(w).Encode(common.GetResponse{
				Success: false,
				Message: "pool_id is required",
			})
			return
		}

		obj, ok := manager.Get(req.PoolID)
		if !ok {
			json.NewEncoder(w).Encode(common.GetResponse{
				Success: false,
				Message: "pool not found",
			})
			return
		}

		json.NewEncoder(w).Encode(common.GetResponse{
			Success: true,
			Object:  obj,
		})
	}
}

func putObjectHandler(manager *PoolManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req common.PutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			json.NewEncoder(w).Encode(common.PutResponse{
				Success: false,
				Message: "invalid request body",
			})
			return
		}

		if req.PoolID == "" {
			json.NewEncoder(w).Encode(common.PutResponse{
				Success: false,
				Message: "pool_id is required",
			})
			return
		}

		if req.Object == nil {
			json.NewEncoder(w).Encode(common.PutResponse{
				Success: false,
				Message: "object is required",
			})
			return
		}

		if manager.Put(req.PoolID, req.Object) {
			json.NewEncoder(w).Encode(common.PutResponse{
				Success: true,
				Message: "object returned to pool",
			})
		} else {
			json.NewEncoder(w).Encode(common.PutResponse{
				Success: false,
				Message: "pool not found",
			})
		}
	}
}

func statsHandler(manager *PoolManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req common.StatsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			json.NewEncoder(w).Encode(common.StatsResponse{
				Success: false,
				Message: "invalid request body",
			})
			return
		}

		if req.PoolID == "" {
			json.NewEncoder(w).Encode(common.StatsResponse{
				Success: false,
				Message: "pool_id is required",
			})
			return
		}

		stats, ok := manager.Stats(req.PoolID)
		if !ok {
			json.NewEncoder(w).Encode(common.StatsResponse{
				Success: false,
				Message: "pool not found",
			})
			return
		}

		json.NewEncoder(w).Encode(common.StatsResponse{
			Success:   true,
			Free:      stats.Free,
			Created:   stats.Created,
			Discarded: stats.Discarded,
			Hits:      stats.Hits,
			Misses:    stats.Misses,
			HitRate:   stats.HitRate,
		})
	}
}

func closePoolHandler(manager *PoolManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req common.ClosePoolRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			json.NewEncoder(w).Encode(common.ClosePoolResponse{
				Success: false,
				Message: "invalid request body",
			})
			return
		}

		if req.PoolID == "" {
			json.NewEncoder(w).Encode(common.ClosePoolResponse{
				Success: false,
				Message: "pool_id is required",
			})
			return
		}

		if manager.Close(req.PoolID) {
			json.NewEncoder(w).Encode(common.ClosePoolResponse{
				Success: true,
				Message: "pool closed",
			})
		} else {
			json.NewEncoder(w).Encode(common.ClosePoolResponse{
				Success: false,
				Message: "pool not found",
			})
		}
	}
}
