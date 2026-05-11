package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"idgen/api"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "server port (default: 8080)")
	flag.Parse()

	if port == 0 {
		port = getEnvPort()
		if port == 0 {
			port = api.DefaultPort
		}
	}

	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/register", server.RegisterNode)
	mux.HandleFunc("/unregister", server.UnregisterNode)
	mux.HandleFunc("/generate", server.GenerateID)
	mux.HandleFunc("/batch-generate", server.BatchGenerate)
	mux.HandleFunc("/nodes", server.ListNodes)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("ID generator server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnvPort() int {
	portStr := os.Getenv("IDGEN_PORT")
	if portStr == "" {
		return 0
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		log.Printf("warning: invalid IDGEN_PORT value: %s, using default", portStr)
		return 0
	}
	return port
}
