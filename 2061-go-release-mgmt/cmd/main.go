package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"release-mgmt/internal/database"
	"release-mgmt/internal/router"
	"syscall"
)

func main() {
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	r := router.New()

	server := &http.Server{
		Addr:    ":8302",
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Starting server on :8302")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")
	server.Close()
	log.Println("Server stopped")
}
