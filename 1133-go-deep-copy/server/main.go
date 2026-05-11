package main

import (
	"log"
	"net/http"
)

const defaultPort = ":8080"

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/prototypes/register", handleRegisterPrototype)
	mux.HandleFunc("/api/prototypes", handleListPrototypes)
	mux.HandleFunc("/api/prototypes/clone", handleClone)
	mux.HandleFunc("/api/deepcopy", handleDeepCopy)
	mux.HandleFunc("/api/compare", handleCompare)

	log.Printf("Server starting on %s", defaultPort)
	if err := http.ListenAndServe(defaultPort, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
