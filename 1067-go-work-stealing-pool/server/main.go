package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":8080", "server address")
	workers := flag.Int("workers", 4, "initial worker count")
	flag.Parse()

	server := NewTaskServer(*workers)

	http.HandleFunc("/submit", server.SubmitHandler)
	http.HandleFunc("/result", server.GetResultHandler)
	http.HandleFunc("/stats", server.StatsHandler)
	http.HandleFunc("/adjust", server.AdjustWorkersHandler)
	http.HandleFunc("/shutdown", server.ShutdownHandler)

	log.Printf("Server starting on %s with %d workers", *addr, *workers)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
