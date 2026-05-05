package main

import (
	"datacleanser/internal/server/handler"
	"datacleanser/internal/server/service"
	"datacleanser/internal/server/storage"
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8080, "server port")
	flag.Parse()

	store := storage.NewStorage()
	svc := service.NewCleanerService(store)
	h := handler.NewHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("data cleanser server starting on %s...", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
