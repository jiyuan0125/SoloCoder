package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"log-shipper/internal/dispatcher"
	"log-shipper/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8624"
	}

	d := dispatcher.New()
	s := server.New(d)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := ":" + port
		if err := s.Start(addr); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	log.Printf("log-shipper is running on port %s", port)
	log.Printf("endpoints:")
	log.Printf("  POST /ingest      - submit logs")
	log.Printf("  GET  /targets     - list targets")
	log.Printf("  POST /targets     - add target")
	log.Printf("  DELETE /targets/{id} - remove target")
	log.Printf("  GET  /stats       - throughput statistics")
	log.Printf("  GET  /health      - health check")

	<-stop
	log.Printf("shutting down...")

	s.Stop()
	d.Stop()
	log.Printf("gracefully shutdown")
}
