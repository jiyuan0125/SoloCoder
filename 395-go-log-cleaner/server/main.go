package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-log-cleaner/common"
)

func main() {
	host := flag.String("host", common.DefaultServerHost, "Server host")
	port := flag.String("port", common.DefaultServerPort, "Server port")
	flag.Parse()

	srv := NewServer(*host, *port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	<-sigChan
	fmt.Println("\nShutting down server...")
	if err := srv.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error stopping server: %v\n", err)
	}
	fmt.Println("Server stopped")
}
