package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"idempotent-retry/internal/api"
	"idempotent-retry/internal/executor"
	"idempotent-retry/internal/query"
	"idempotent-retry/internal/scheduler"
	"idempotent-retry/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8300"
	}

	store := store.New()
	executor := executor.New()
	queryService := query.New(store)
	sch := scheduler.New(store, executor)

	sch.Start(5)

	handler := api.NewHandler(store, sch, queryService)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.HealthCheck)
	mux.HandleFunc("/tasks", handler.SubmitTask)
	mux.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/subscribe") {
			handler.SubscribeCallback(w, r)
			return
		}
		handler.GetTask(w, r)
	})
	mux.HandleFunc("/tasks/search", handler.GetTaskByIdempotencyKey)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	sch.Stop()
}
