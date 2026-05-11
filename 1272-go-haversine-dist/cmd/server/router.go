package main

import (
	"net/http"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/distance", handleDistance)
	mux.HandleFunc("/polyline/length", handlePolylineLength)
	mux.HandleFunc("/polyline/closest", handlePointToPolyline)
	mux.HandleFunc("/batch", handleBatch)
	mux.HandleFunc("/circle", handleCircle)
	mux.HandleFunc("/health", handleHealth)

	return mux
}
