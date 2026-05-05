package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"csv-merger/common"
	"csv-merger/server/network"
)

func main() {
	host := flag.String("host", common.DefaultServerHost, "Server host to listen on")
	port := flag.String("port", common.DefaultServerPort, "Server port to listen on")
	flag.Parse()

	server := network.NewServer(*host, *port)

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	<-sigCh
	log.Println("Shutting down server...")
	server.Stop()
}
