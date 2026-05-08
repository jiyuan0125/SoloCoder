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

	mux := http.NewServeMux()

	mux.HandleFunc("/index/create", server.CreateIndex)
	mux.HandleFunc("/index/delete", server.DeleteIndex)
	mux.HandleFunc("/index/list", server.ListIndices)

	mux.HandleFunc("/bit/set", server.SetBit)
	mux.HandleFunc("/bit/clear", server.ClearBit)
	mux.HandleFunc("/bit/get", server.GetBit)

	mux.HandleFunc("/count", server.Count)
	mux.HandleFunc("/indices", server.Indices)

	mux.HandleFunc("/op", server.Operation)

	mux.HandleFunc("/serialize", server.Serialize)
	mux.HandleFunc("/deserialize", server.Deserialize)

	log.Printf("Starting bitmap index server on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
