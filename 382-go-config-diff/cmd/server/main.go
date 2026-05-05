package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/config-diff/internal/server"
)

const (
    defaultHost = "127.0.0.1"
    defaultPort = 8765
)

func main() {
    host := flag.String("host", defaultHost, "Host to listen on")
    port := flag.Int("port", defaultPort, "Port to listen on")
    help := flag.Bool("h", false, "Show help message")

    flag.Usage = showHelp
    flag.Parse()

    if *help {
        showHelp()
        os.Exit(0)
    }

    srv := server.NewServer(*host, *port)
    if err := srv.Start(); err != nil {
        fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
        os.Exit(1)
    }
}

func showHelp() {
    fmt.Println("Config Diff Server - Backend service for comparing configuration files")
    fmt.Println()
    fmt.Println("Usage: config-diff-server [options]")
    fmt.Println()
    fmt.Println("Options:")
    fmt.Println("  -host string    Host to listen on (default \"127.0.0.1\")")
    fmt.Println("  -port int       Port to listen on (default 8765)")
    fmt.Println("  -h              Show this help message")
    fmt.Println()
    fmt.Println("Description:")
    fmt.Println("  This is the backend server that handles configuration file comparison requests.")
    fmt.Println("  It listens for TCP connections from the config-diff client and performs the actual")
    fmt.Println("  parsing and comparison of configuration files.")
}
