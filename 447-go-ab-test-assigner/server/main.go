package main

import (
	"abtest/server/handler"
	"abtest/server/service"
	"abtest/server/store"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := flag.Int("port", 8080, "HTTP server port")
	flag.Parse()

	st := store.NewStore()
	svc := service.NewService(st)
	h := handler.NewHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("AB Test server starting on %s...", addr)
	log.Printf("Endpoints:")
	log.Printf("  GET  /health          - Health check")
	log.Printf("  GET  /                 - API info")
	log.Printf("  POST /experiments/create - Create experiment")
	log.Printf("  GET  /experiments      - List experiments")
	log.Printf("  GET  /experiments/{id} - Get experiment")
	log.Printf("  POST /experiments/start/{id} - Start experiment")
	log.Printf("  POST /experiments/end/{id}   - End experiment")
	log.Printf("  POST /experiments/archive/{id} - Archive experiment")
	log.Printf("  POST /experiments/traffic - Update traffic distribution (gray only)")
	log.Printf("  POST /assign           - Assign user to group")
	log.Printf("  POST /metrics          - Record user metrics")
	log.Printf("  GET  /experiments/stats/{id} - Get experiment statistics")
	log.Printf("  GET  /users/assignments/{user_id} - Get user assignments")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
