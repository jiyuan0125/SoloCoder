package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
)

const defaultPort = 8215

func getPort() int {
	portStr := os.Getenv("CMS_PORT")
	if portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			return port
		}
	}

	port := flag.Int("port", defaultPort, "server port")
	flag.Parse()
	return *port
}

func main() {
	port := getPort()
	addr := ":" + strconv.Itoa(port)

	handler := NewServer()

	log.Printf("Count-Min Sketch server starting on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
