package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	
	"port-forward/protocol"
)

var (
	controlPort = flag.Int("control-port", protocol.DefaultControlPort, "Control port for client commands")
	verbose     = flag.Bool("verbose", false, "Enable verbose logging")
)

func main() {
	flag.Parse()
	
	server := NewServer(*controlPort, *verbose)
	
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		if err := server.Start(); err != nil {
			fmt.Printf("Server error: %v\n", err)
			os.Exit(1)
		}
	}()
	
	<-sigCh
	fmt.Println("\nShutting down server...")
	
	server.Shutdown()
	
	fmt.Println("Server stopped gracefully")
}

func logf(format string, v ...interface{}) {
	if *verbose {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("[%s] %s\n", timestamp, fmt.Sprintf(format, v...))
	}
}
