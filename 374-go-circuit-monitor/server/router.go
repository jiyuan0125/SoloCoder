package main

import (
	"net/http"

	"circuit-monitor/circuitbreaker"
)

func NewRouter(registry *circuitbreaker.Registry) http.Handler {
	handler := NewHandler(registry)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.HealthCheck)

	mux.HandleFunc("GET /api/circuits", handler.ListCircuits)
	mux.HandleFunc("POST /api/circuits", handler.CreateCircuit)

	mux.HandleFunc("GET /api/circuits/", handler.GetCircuit)
	mux.HandleFunc("POST /api/circuits/", handler.HandleCircuitPost)

	return mux
}

func (h *Handler) HandleCircuitPost(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	parts := splitPath(path)

	if len(parts) < 4 {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	if len(parts) == 4 {
		name := parts[3]
		if name == "" {
			writeError(w, http.StatusBadRequest, "circuit name is required")
			return
		}

		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	action := parts[4]
	switch action {
	case "reset":
		h.ResetCircuit(w, r)
	case "force-state":
		h.ForceState(w, r)
	case "config":
		if r.Method == http.MethodGet {
			h.GetConfig(w, r)
		} else {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	case "save":
		h.SaveState(w, r)
	case "load":
		h.LoadState(w, r)
	default:
		writeError(w, http.StatusNotFound, "unknown action")
	}
}

func splitPath(path string) []string {
	var parts []string
	current := ""
	for _, c := range path {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
