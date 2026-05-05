package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-feedback-handler/internal/server"
)

var (
	port    = flag.Int("port", 8080, "HTTP server port")
	runOnce = flag.Bool("run-once", false, "Run scheduler tasks once and exit")
)

func main() {
	flag.Parse()

	store := server.NewStore()
	svc := server.NewService(store)
	handler := server.NewHandler(svc)
	scheduler := server.NewScheduler(svc)

	if *runOnce {
		log.Println("Running scheduler tasks once...")
		scheduler.RunNow()
		log.Println("Done.")
		os.Exit(0)
	}

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	addr := fmt.Sprintf(":%d", *port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Feedback handler server starting on %s", addr)
		scheduler.Start()
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	scheduler.Stop()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}
