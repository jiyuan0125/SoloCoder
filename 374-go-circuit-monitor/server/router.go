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

	mux.HandleFunc("GET /api/circuits/{name}", handler.GetCircuit)
	mux.HandleFunc("GET /api/circuits/{name}/config", handler.GetConfig)

	mux.HandleFunc("POST /api/circuits/{name}/reset", handler.ResetCircuit)
	mux.HandleFunc("POST /api/circuits/{name}/force-state", handler.ForceState)
	mux.HandleFunc("POST /api/circuits/{name}/save", handler.SaveState)
	mux.HandleFunc("POST /api/circuits/{name}/load", handler.LoadState)

	return mux
}
