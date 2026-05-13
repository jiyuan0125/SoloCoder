package server

import (
	"config-center/internal/store"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	s *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{s: s}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/configs", h.handleConfigs)
	mux.HandleFunc("/configs/", h.handleConfigByKey)

	mux.HandleFunc("/versions", h.handleVersions)
	mux.HandleFunc("/versions/", h.handleVersionByID)

	mux.HandleFunc("/diff", h.handleDiff)
	mux.HandleFunc("/rollback", h.handleRollback)

	mux.HandleFunc("/watches", h.handleWatches)
	mux.HandleFunc("/watches/", h.handleWatchByID)
}

func (h *Handler) handleConfigs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		env := r.URL.Query().Get("env")
		project := r.URL.Query().Get("project")
		configs := h.s.List(env, project)
		writeJSON(w, http.StatusOK, configs)
	case http.MethodPost, http.MethodPut:
		var req struct {
			Env     string `json:"env"`
			Project string `json:"project"`
			Key     string `json:"key"`
			Value   string `json:"value"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Env == "" || req.Project == "" || req.Key == "" {
			writeError(w, http.StatusBadRequest, "env, project, key are required")
			return
		}
		if !store.ValidateKey(req.Key) {
			writeError(w, http.StatusBadRequest, "invalid key: only alphanumeric, underscore, and dot allowed")
			return
		}
		if !store.ValidateValue(req.Value) {
			writeError(w, http.StatusUnprocessableEntity, "value exceeds 64KB limit")
			return
		}
		cfg, _, err := h.s.Set(req.Env, req.Project, req.Key, req.Value)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleConfigByKey(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/configs/"), "/")
	if len(parts) < 3 {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	env, project, key := parts[0], parts[1], parts[2]

	switch r.Method {
	case http.MethodGet:
		cfg, ok := h.s.Get(env, project, key)
		if !ok {
			writeError(w, http.StatusNotFound, "config not found")
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	case http.MethodDelete:
		_, ok := h.s.Delete(env, project, key)
		if !ok {
			writeError(w, http.StatusNotFound, "config not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	env := r.URL.Query().Get("env")
	project := r.URL.Query().Get("project")
	key := r.URL.Query().Get("key")
	versions := h.s.ListVersions(env, project, key)
	writeJSON(w, http.StatusOK, versions)
}

func (h *Handler) handleVersionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/versions/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid version id")
		return
	}
	v, ok := h.s.GetVersion(id)
	if !ok {
		writeError(w, http.StatusNotFound, "version not found")
		return
	}
	writeJSON(w, http.StatusOK, v)
}

func (h *Handler) handleDiff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	v1Str := r.URL.Query().Get("v1")
	v2Str := r.URL.Query().Get("v2")
	if v1Str == "" || v2Str == "" {
		writeError(w, http.StatusBadRequest, "v1 and v2 are required")
		return
	}
	v1, err1 := strconv.Atoi(v1Str)
	v2, err2 := strconv.Atoi(v2Str)
	if err1 != nil || err2 != nil {
		writeError(w, http.StatusBadRequest, "invalid version numbers")
		return
	}
	diffs := h.s.Diff(v1, v2)
	writeJSON(w, http.StatusOK, diffs)
}

func (h *Handler) handleRollback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req struct {
		Version int `json:"version"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Version <= 0 {
		writeError(w, http.StatusBadRequest, "valid version is required")
		return
	}
	cfg, _, err := h.s.Rollback(req.Version)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func (h *Handler) handleWatches(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		watches := h.s.ListWatches()
		writeJSON(w, http.StatusOK, watches)
	case http.MethodPost:
		var req struct {
			Env         string `json:"env"`
			Project     string `json:"project"`
			Key         string `json:"key"`
			CallbackURL string `json:"callback_url"`
			WatchType   string `json:"watch_type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if req.Env == "" || req.Project == "" || req.CallbackURL == "" {
			writeError(w, http.StatusBadRequest, "env, project, callback_url are required")
			return
		}
		isKeyWatch := req.WatchType == "key"
		if isKeyWatch {
			if req.Key == "" {
				writeError(w, http.StatusBadRequest, "key is required for key watch")
				return
			}
			if !store.ValidateKey(req.Key) {
				writeError(w, http.StatusBadRequest, "invalid key")
				return
			}
		}
		watch := h.s.RegisterWatch(req.Env, req.Project, req.Key, req.CallbackURL, isKeyWatch)
		writeJSON(w, http.StatusOK, watch)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) handleWatchByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/watches/")
	switch r.Method {
	case http.MethodGet:
		watch, ok := h.s.GetWatch(id)
		if !ok {
			writeError(w, http.StatusNotFound, "watch not found")
			return
		}
		writeJSON(w, http.StatusOK, watch)
	case http.MethodDelete:
		if !h.s.UnregisterWatch(id) {
			writeError(w, http.StatusNotFound, "watch not found")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "unregistered"})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
