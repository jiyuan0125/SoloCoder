package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		flagPort := flag.String("port", "9001", "Server port")
		flag.Parse()
		port = *flagPort
	}
	return port
}

func main() {
	port := getPort()
	server := NewServer()

	mux := http.NewServeMux()

	mux.HandleFunc("/vehicles", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			server.handleCreateVehicle(w, r)
		case http.MethodGet:
			server.handleListVehicles(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/vehicles/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		server.handleGetVehicle(w, r)
	})

	mux.HandleFunc("/evaluations", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			server.handleEvaluateVehicle(w, r)
		case http.MethodGet:
			server.handleListEvaluations(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/deposits", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		server.handlePayDeposit(w, r)
	})

	mux.HandleFunc("/transfers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		server.handlePayFull(w, r)
	})

	addr := ":" + port
	fmt.Printf("Server starting on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
