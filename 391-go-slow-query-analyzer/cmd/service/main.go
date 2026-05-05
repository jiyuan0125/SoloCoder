package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"slowquery/internal/server"
	"slowquery/protocol"
)

func main() {
	port := flag.String("port", protocol.DefaultPort, "Server port")
	flag.Parse()

	addr := fmt.Sprintf(":%s", *port)

	srv := server.NewServer(addr)

	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down server...")
	srv.Stop()
}
