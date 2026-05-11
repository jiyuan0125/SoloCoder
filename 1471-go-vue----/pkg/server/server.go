package server

import (
	"bus-station/internal/core"
	"encoding/json"
	"log"
	"net/http"
)

type Server struct {
	store     *core.Store
	scheduler *core.Scheduler
}

func NewServer(store *core.Store, scheduler *core.Scheduler) *Server {
	return &Server{
		store:     store,
		scheduler: scheduler,
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("JSON encoding error: %v", err)
	}
}

func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
