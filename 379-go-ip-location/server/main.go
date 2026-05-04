package main

import (
	"fmt"
	"log"
	"net/http"

	"ip-location/iplocation"
)

const defaultPort = 8080

var db *iplocation.LocationDB

func init() {
	db = iplocation.NewBuiltinDB()
}

func main() {
	addr := fmt.Sprintf(":%d", defaultPort)

	mux := http.NewServeMux()

	setupRoutes(mux)

	log.Printf("IP Location Server starting on %s", addr)
	log.Printf("Loaded %d IP ranges", db.RangeCount())

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
