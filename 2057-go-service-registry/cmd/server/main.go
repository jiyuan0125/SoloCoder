package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"serviceregistry/internal/server"
)

func main() {
	var (
		port   int
		dbPath string
	)

	flag.IntVar(&port, "port", 9201, "server port")
	flag.StringVar(&dbPath, "db", "registry.db", "sqlite database path")
	flag.Parse()

	srv, err := server.New(dbPath, port)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.Start()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	case <-stop:
		fmt.Println("\nShutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Stop(ctx); err != nil {
			log.Fatalf("Server shutdown error: %v", err)
		}
	}

	fmt.Println("Server stopped gracefully")
}
