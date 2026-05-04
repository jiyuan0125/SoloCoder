package main

import (
	"net/http"
)

func NewRouter(handler *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/check", handler.Check)
	mux.HandleFunc("/calculate-hours", handler.CalculateHours)
	mux.HandleFunc("/config", handler.Config)

	return mux
}
