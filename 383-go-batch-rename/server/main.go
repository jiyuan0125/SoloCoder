package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
)

const (
    defaultPort = "8888"
)

func main() {
    port := defaultPort
    if len(os.Args) > 1 {
        port = os.Args[1]
    }

    server, err := NewServer(":" + port)
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }

    fmt.Printf("Batch rename server started on port %s\n", port)

    go func() {
        if err := server.Start(); err != nil {
            log.Fatalf("Server error: %v", err)
        }
    }()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    <-sigCh
    fmt.Println("\nShutting down server...")
    server.Stop()
    fmt.Println("Server stopped")
}
