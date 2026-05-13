package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"router/internal/model"
	"router/internal/router"
	"router/internal/store"
)

type Handler struct {
	store   *store.Store
	matcher *router.Matcher
}

func New(store *store.Store, matcher *router.Matcher) *Handler {
	return &Handler{store: store, matcher: matcher}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/admin/routes", h.handleRoutes)
	mux.HandleFunc("/admin/routes/", h.handleRoute)
	mux.HandleFunc("/admin/reload", h.handleReload)
}

func (h *Handler) handleRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listRoutes(w, r)
	case http.MethodPost:
		h.addRoute(w, r)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) handleRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/admin/routes/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil {
		http.Error(w, "Invalid route ID", http.StatusBadRequest)
		return
	}
	h.deleteRoute(w, r, id)
}

func (h *Handler) listRoutes(w http.ResponseWriter, r *http.Request) {
	routes, err := h.store.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(routes)
}

func (h *Handler) addRoute(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path   string `json:"path"`
		Method string `json:"method"`
		Target string `json:"target"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Path == "" || req.Method == "" || req.Target == "" {
		http.Error(w, "Missing required fields: path, method, target", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(req.Path, "/") {
		http.Error(w, "Path must start with /", http.StatusBadRequest)
		return
	}
	req.Method = strings.ToUpper(req.Method)

	route, err := h.store.Add(r.Context(), req.Path, req.Method, req.Target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.reloadMatcher(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(route)
}

func (h *Handler) deleteRoute(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.store.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = h.reloadMatcher(r.Context())
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := h.reloadMatcher(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) reloadMatcher(ctx context.Context) error {
	routes, err := h.store.GetAll(ctx)
	if err != nil {
		return err
	}
	h.matcher.Update(routes)
	return nil
}

func (h *Handler) ReloadMatcher(ctx context.Context) error {
	return h.reloadMatcher(ctx)
}

func (h *Handler) ListRoutes(ctx context.Context) ([]*model.Route, error) {
	return h.store.GetAll(ctx)
}

func (h *Handler) AddRoute(ctx context.Context, path, method, target string) (*model.Route, error) {
	route, err := h.store.Add(ctx, path, strings.ToUpper(method), target)
	if err != nil {
		return nil, err
	}
	if err := h.reloadMatcher(ctx); err != nil {
		return route, err
	}
	return route, nil
}

func (h *Handler) DeleteRoute(ctx context.Context, id int64) error {
	if err := h.store.Delete(ctx, id); err != nil {
		return err
	}
	return h.reloadMatcher(ctx)
}
