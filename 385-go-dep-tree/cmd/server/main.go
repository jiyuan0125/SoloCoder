package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	
	"dep-tree/internal/server"
)

func main() {
	port := flag.Int("port", 9876, "Server port")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting dependency tree server on %s...", addr)

	srv := server.New()
	
	go func() {
		if err := srv.Start(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")
	srv.Stop()
	log.Println("Server stopped")
}
