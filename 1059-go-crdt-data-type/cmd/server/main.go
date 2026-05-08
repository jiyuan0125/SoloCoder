package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := ":8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		if !strings.HasPrefix(envPort, ":") {
			port = ":" + envPort
		} else {
			port = envPort
		}
	}

	server := NewServer()

	http.HandleFunc("/health", server.healthHandler)
	http.HandleFunc("/instances", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instances" {
			if r.Method == http.MethodPost {
				server.createInstanceHandler(w, r)
			} else if r.Method == http.MethodGet {
				server.listInstancesHandler(w, r)
			} else {
				server.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			}
			return
		}

		if strings.HasSuffix(r.URL.Path, "/operations") && r.Method == http.MethodPost {
			server.operationHandler(w, r)
			return
		}

		if r.Method == http.MethodGet {
			server.getInstanceHandler(w, r)
			return
		}

		server.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	})

	http.HandleFunc("/merge", server.mergeHandler)

	fmt.Printf("CRDT Server starting on port %s...\n", port)
	fmt.Println("Available endpoints:")
	fmt.Println("  POST /instances          - Create a new CRDT instance")
	fmt.Println("  GET  /instances          - List all instances")
	fmt.Println("  GET  /instances/{id}     - Get instance details")
	fmt.Println("  POST /instances/{id}/operations - Execute operation on instance")
	fmt.Println("  POST /merge              - Merge two instances")
	fmt.Println("  GET  /health             - Health check")

	log.Fatal(http.ListenAndServe(port, nil))
}
