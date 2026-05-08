package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":8080", "server address")
	flag.Parse()

	server := NewServer()

	http.HandleFunc("/operate", server.handleOperation)
	http.HandleFunc("/batch", server.handleQueryBatch)

	log.Printf("interval server listening on %s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
