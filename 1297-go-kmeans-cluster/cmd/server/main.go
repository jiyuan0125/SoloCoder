package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"kmeans-cluster/server"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "Server port (default: 8080 or KMEANS_PORT env)")
	flag.Parse()

	if port == 0 {
		if envPort := os.Getenv("KMEANS_PORT"); envPort != "" {
			if p, err := strconv.Atoi(envPort); err == nil {
				port = p
			}
		}
		if port == 0 {
			port = 8080
		}
	}

	addr := ":" + strconv.Itoa(port)
	log.Printf("Starting K-Means server on %s", addr)

	http.HandleFunc("/cluster", server.ClusterHandler)
	http.HandleFunc("/wcss", server.WCSSHandler)
	http.HandleFunc("/health", server.HealthHandler)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
