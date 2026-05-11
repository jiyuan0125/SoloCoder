package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"genetic-algo/cmd/server/handlers"
	"genetic-algo/cmd/server/task"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "Server port (default: from GA_PORT env or 8080)")
	flag.Parse()

	if port == 0 {
		if envPort := os.Getenv("GA_PORT"); envPort != "" {
			if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
				port = p
			}
		}
		if port == 0 {
			port = 8102
		}
	}

	taskManager := task.NewManager()

	mux := http.NewServeMux()

	handlers.Register(mux, taskManager)

	addr := ":" + strconv.Itoa(port)
	log.Printf("Genetic Algorithm Server starting on %s", addr)
	log.Printf("Available endpoints:")
	log.Printf("  GET  /api/v1/functions      - List available test functions")
	log.Printf("  POST /api/v1/optimize       - Submit optimization task")
	log.Printf("  GET  /api/v1/tasks          - List all tasks")
	log.Printf("  GET  /api/v1/tasks/{id}     - Get task status and result")
	log.Printf("  GET  /api/v1/tasks/{id}/statistics - Get task convergence statistics")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
