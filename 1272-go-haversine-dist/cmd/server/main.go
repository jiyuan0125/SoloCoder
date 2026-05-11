package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	var port string

	flag.StringVar(&port, "port", "", "Port to listen on (e.g., :8080)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
		if port == "" {
			port = ":8080"
		}
		if port[0] != ':' {
			port = ":" + port
		}
	}

	router := NewRouter()

	fmt.Printf("Server starting on port %s\n", port)
	log.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}
