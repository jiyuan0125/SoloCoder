package api

import (
	"cache-layer/cache"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	cache cache.Cache
}

func NewHandler(c cache.Cache) *Handler {
	return &Handler{cache: c}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/get/", h.handleGet)
	mux.HandleFunc("/set/", h.handleSet)
	mux.HandleFunc("/delete/", h.handleDelete)
	mux.HandleFunc("/flush", h.handleFlush)
	mux.HandleFunc("/capacity", h.handleCapacity)
	mux.HandleFunc("/stats", h.handleStats)
}

func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/get/")
	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}

	value, ok := h.cache.Get(key)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

func (h *Handler) handleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/set/")
	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}

	value := r.FormValue("value")
	if value == "" {
		http.Error(w, "value required", http.StatusBadRequest)
		return
	}

	ttlStr := r.FormValue("ttl")
	ttl := time.Duration(0)
	if ttlStr != "" {
		ttlSec, err := strconv.ParseInt(ttlStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid ttl", http.StatusBadRequest)
			return
		}
		ttl = time.Duration(ttlSec) * time.Second
	}

	h.cache.Set(key, value, ttl)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/delete/")
	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}

	if !h.cache.Delete(key) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) handleFlush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.cache.Flush()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) handleCapacity(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cap := h.cache.Capacity()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"capacity": cap})

	case http.MethodPut, http.MethodPost:
		capStr := r.FormValue("capacity")
		if capStr == "" {
			http.Error(w, "capacity required", http.StatusBadRequest)
			return
		}
		cap, err := strconv.Atoi(capStr)
		if err != nil || cap < 0 {
			http.Error(w, "invalid capacity", http.StatusBadRequest)
			return
		}
		h.cache.SetCapacity(cap)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := h.cache.Stats()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"hit_rate":    stats.HitRate,
		"item_count":  stats.ItemCount,
		"total_query": stats.TotalHits,
	})
}
