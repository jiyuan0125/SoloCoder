package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "Server port")
	flag.Parse()

	if port == 0 {
		envPort := os.Getenv("PING_SERVER_PORT")
		if envPort != "" {
			p, err := strconv.Atoi(envPort)
			if err != nil {
				log.Fatalf("Invalid port in PING_SERVER_PORT: %v", err)
			}
			port = p
		} else {
			port = 8080
		}
	}

	manager := NewTaskManager()
	handler := NewPingHandler(manager)

	r := mux.NewRouter()
	handler.RegisterRoutes(r)

	addr := ":" + strconv.Itoa(port)
	log.Printf("Ping server starting on %s...", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
