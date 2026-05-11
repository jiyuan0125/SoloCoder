package main

import (
	"cleaning-service/internal/core"
	"flag"
	"fmt"
	"net/http"
	"os"
)

func main() {
	var port string

	flag.StringVar(&port, "port", "", "HTTP server port")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "9013"
	}

	store := core.NewStore()
	service := core.NewService(store)
	handler := NewHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/clients", handler.HandleClients)
	mux.HandleFunc("/api/clients/", handler.HandleClientByID)

	mux.HandleFunc("/api/zones", handler.HandleZones)
	mux.HandleFunc("/api/service-areas", handler.HandleServiceAreas)

	mux.HandleFunc("/api/teams", handler.HandleTeams)
	mux.HandleFunc("/api/cleaners", handler.HandleCleaners)

	mux.HandleFunc("/api/schedules", handler.HandleSchedules)
	mux.HandleFunc("/api/leaves", handler.HandleLeaves)
	mux.HandleFunc("/api/leaves/", handler.HandleLeaveByID)

	mux.HandleFunc("/api/tasks", handler.HandleTasks)
	mux.HandleFunc("/api/tasks/generate", handler.HandleGenerateTasks)
	mux.HandleFunc("/api/tasks/complete", handler.HandleCompleteTask)

	mux.HandleFunc("/api/quality-checks", handler.HandleQualityChecks)
	mux.HandleFunc("/api/todos", handler.HandleTodos)
	mux.HandleFunc("/api/todos/", handler.HandleTodoByID)

	addr := ":" + port
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
