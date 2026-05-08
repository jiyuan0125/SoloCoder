package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	state := NewServerState()
	handler := NewHandler(state)

	mux := http.NewServeMux()

	mux.HandleFunc("/map/set", handler.HandleSetMap)
	mux.HandleFunc("/map/random", handler.HandleRandomMap)
	mux.HandleFunc("/map", handler.HandleGetMap)
	mux.HandleFunc("/pathfind", handler.HandlePathfind)
	mux.HandleFunc("/result", handler.HandleGetResult)

	log.Printf("A* pathfinding server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
