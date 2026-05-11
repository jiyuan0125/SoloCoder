package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "port to listen on")
	flag.Parse()

	if port == "" {
		port = os.Getenv("SPLAYTREE_PORT")
	}
	if port == "" {
		port = "8080"
	}

	handler := NewHandler()

	http.HandleFunc("/put", handler.PutHandler)
	http.HandleFunc("/get", handler.GetHandler)
	http.HandleFunc("/delete", handler.DeleteHandler)
	http.HandleFunc("/range/query", handler.RangeQueryHandler)
	http.HandleFunc("/range/add", handler.RangeAddHandler)
	http.HandleFunc("/range/set", handler.RangeSetHandler)

	addr := ":" + port
	log.Printf("splaytree server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
