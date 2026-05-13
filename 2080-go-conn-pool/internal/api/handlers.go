package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"connpool/internal/models"
	"connpool/internal/pool"
	"connpool/internal/stats"
	"connpool/internal/storage"
)

type Handler struct {
	storage    *storage.SQLiteStorage
	poolMgr    *pool.Manager
	statsMgr   *stats.Manager
}

func NewHandler(storage *storage.SQLiteStorage, poolMgr *pool.Manager, statsMgr *stats.Manager) *Handler {
	return &Handler{
		storage:  storage,
		poolMgr:  poolMgr,
		statsMgr: statsMgr,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/pools", h.handlePools)
	mux.HandleFunc("/pools/", h.handlePoolByID)
	mux.HandleFunc("/pools/{id}/status", h.handlePoolStatus)
	mux.HandleFunc("/pools/{id}/acquire", h.handleAcquire)
	mux.HandleFunc("/pools/{id}/release", h.handleRelease)
	mux.HandleFunc("/quotas/total", h.handleQuotasTotal)
	mux.HandleFunc("/statistics", h.handleStatistics)
	mux.HandleFunc("/health", h.handleHealth)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]interface{}{
		"error":   http.StatusText(status),
		"message": message,
	})
}

func (h *Handler) handlePools(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listPools(w, r)
	case http.MethodPost:
		h.createPool(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) listPools(w http.ResponseWriter, r *http.Request) {
	pools, err := h.storage.ListPools(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, pools)
}

type CreatePoolRequest struct {
	Name            string        `json:"name"`
	MaxConnections  int           `json:"max_connections"`
	IdleTimeout     time.Duration `json:"idle_timeout"`
	MaxLifetime     time.Duration `json:"max_lifetime"`
	BackendAddress  string        `json:"backend_address"`
	WaitTimeout     time.Duration `json:"wait_timeout"`
	HealthCheckURL  string        `json:"health_check_url"`
}

func (h *Handler) createPool(w http.ResponseWriter, r *http.Request) {
	var req CreatePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.MaxConnections <= 0 {
		writeError(w, http.StatusBadRequest, "max_connections must be positive")
		return
	}

	if req.BackendAddress == "" {
		writeError(w, http.StatusBadRequest, "backend_address is required")
		return
	}

	config := &models.PoolConfig{
		ID:             generatePoolID(req.Name),
		Name:           req.Name,
		MaxConnections: req.MaxConnections,
		IdleTimeout:    req.IdleTimeout,
		MaxLifetime:    req.MaxLifetime,
		BackendAddress: req.BackendAddress,
		WaitTimeout:    req.WaitTimeout,
		HealthCheckURL: req.HealthCheckURL,
		Quota:          req.MaxConnections,
	}

	if err := h.storage.CreatePool(r.Context(), config); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.poolMgr.CreatePool(config); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.statsMgr.UpdateAndSaveStats(r.Context())
	writeJSON(w, http.StatusCreated, config)
}

func (h *Handler) handlePoolByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/pools/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusBadRequest, "pool id required")
		return
	}

	poolID := parts[0]

	switch r.Method {
	case http.MethodGet:
		h.getPool(w, r, poolID)
	case http.MethodDelete:
		h.deletePool(w, r, poolID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) getPool(w http.ResponseWriter, r *http.Request, poolID string) {
	poolConfig, err := h.storage.GetPool(r.Context(), poolID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if poolConfig == nil {
		writeError(w, http.StatusNotFound, "pool not found")
		return
	}
	writeJSON(w, http.StatusOK, poolConfig)
}

func (h *Handler) deletePool(w http.ResponseWriter, r *http.Request, poolID string) {
	if err := h.poolMgr.DeletePool(poolID); err != nil {
		if errors.Is(err, pool.ErrPoolNotFound) {
			writeError(w, http.StatusNotFound, "pool not found")
			return
		}
		if errors.Is(err, pool.ErrPoolInUse) {
			p, _ := h.poolMgr.GetPool(poolID)
			status := p.GetStatus()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "pool is in use",
				"message": "pool has active connections or waiting requests",
				"status":  status,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.storage.DeletePool(r.Context(), poolID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.statsMgr.UpdateAndSaveStats(r.Context())
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handlePoolStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/pools/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		writeError(w, http.StatusBadRequest, "pool id required")
		return
	}

	poolID := parts[0]
	p, err := h.poolMgr.GetPool(poolID)
	if err != nil {
		if errors.Is(err, pool.ErrPoolNotFound) {
			writeError(w, http.StatusNotFound, "pool not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	status := p.GetStatus()
	writeJSON(w, http.StatusOK, status)
}

func (h *Handler) handleAcquire(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/pools/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		writeError(w, http.StatusBadRequest, "pool id required")
		return
	}

	poolID := parts[0]
	p, err := h.poolMgr.GetPool(poolID)
	if err != nil {
		if errors.Is(err, pool.ErrPoolNotFound) {
			writeError(w, http.StatusNotFound, "pool not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	connID, err := p.Acquire(ctx)
	if err != nil {
		status := p.GetStatus()
		if errors.Is(err, pool.ErrWaitTimeout) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "service unavailable",
				"message": "wait timeout for connection",
				"status":  status,
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "service unavailable",
			"message": err.Error(),
			"status":  status,
		})
		return
	}

	h.statsMgr.UpdateAndSaveStats(r.Context())
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"connection_id": connID,
		"pool_id":       poolID,
	})
}

type ReleaseRequest struct {
	ConnectionID string `json:"connection_id"`
	Healthy      bool   `json:"healthy"`
}

func (h *Handler) handleRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/pools/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		writeError(w, http.StatusBadRequest, "pool id required")
		return
	}

	poolID := parts[0]
	p, err := h.poolMgr.GetPool(poolID)
	if err != nil {
		if errors.Is(err, pool.ErrPoolNotFound) {
			writeError(w, http.StatusNotFound, "pool not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req ReleaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.ConnectionID == "" {
		writeError(w, http.StatusBadRequest, "connection_id is required")
		return
	}

	if err := p.Release(req.ConnectionID, req.Healthy); err != nil {
		if errors.Is(err, pool.ErrConnNotFound) {
			writeError(w, http.StatusNotFound, "connection not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.statsMgr.UpdateAndSaveStats(r.Context())
	w.WriteHeader(http.StatusNoContent)
}

type AdjustQuotasRequest struct {
	NewTotal int `json:"new_total"`
}

func (h *Handler) handleQuotasTotal(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		var req AdjustQuotasRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.NewTotal < 0 {
			writeError(w, http.StatusBadRequest, "new_total must be non-negative")
			return
		}

		if err := h.statsMgr.AdjustQuotas(r.Context(), req.NewTotal); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		stats, err := h.statsMgr.GetLatestStats(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, stats)
	case http.MethodGet:
		stats, err := h.statsMgr.GetLatestStats(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, stats)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stats := h.poolMgr.GetStatistics()
	writeJSON(w, http.StatusOK, stats)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func generatePoolID(name string) string {
	safe := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)
	return strings.ToLower(strings.Trim(safe, "-"))
}
