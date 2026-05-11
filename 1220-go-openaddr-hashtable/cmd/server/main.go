package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	
	"hashtable/pkg/hashtable"
)

type Server struct {
	ht    *hashtable.HashTable
	mutex sync.RWMutex
}

func main() {
	port := flag.String("port", "", "Port to listen on")
	flag.Parse()
	
	if *port == "" {
		*port = os.Getenv("PORT")
		if *port == "" {
			*port = "8080"
		}
	}
	
	server := &Server{}
	
	http.HandleFunc("/create", server.handleCreate)
	http.HandleFunc("/put", server.handlePut)
	http.HandleFunc("/get", server.handleGet)
	http.HandleFunc("/remove", server.handleRemove)
	http.HandleFunc("/stats", server.handleStats)
	
	log.Printf("Server starting on port %s...", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), nil))
}
