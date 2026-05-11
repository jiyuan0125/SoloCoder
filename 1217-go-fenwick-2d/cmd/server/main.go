package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"fenwick2d/api"
	"fenwick2d/fenwick2d"
)

var (
	ft   *fenwick2d.Fenwick2D
	mu   sync.RWMutex
	port string
)

func main() {
	flagPort := flag.String("port", "", "server port")
	flag.Parse()

	if *flagPort != "" {
		port = *flagPort
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	} else {
		port = "8300"
	}

	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/sum", sumHandler)
	http.HandleFunc("/init", initHandler)

	fmt.Printf("Server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	newFt, err := fenwick2d.New(req.Rows, req.Cols)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	ft = newFt

	json.NewEncoder(w).Encode(api.CreateResponse{Success: true})
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if ft == nil {
		sendError(w, "matrix not created", http.StatusBadRequest)
		return
	}

	if err := ft.Update(req.X, req.Y, req.Delta); err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(api.UpdateResponse{Success: true})
}

func sumHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.SumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	mu.RLock()
	defer mu.RUnlock()

	if ft == nil {
		sendError(w, "matrix not created", http.StatusBadRequest)
		return
	}

	sum, err := ft.QueryRange(req.Lx, req.Ly, req.Rx, req.Ry)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(api.SumResponse{Success: true, Sum: sum})
}

func initHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.InitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	newFt, err := fenwick2d.NewWithMatrix(req.Matrix)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	ft = newFt

	json.NewEncoder(w).Encode(api.InitResponse{Success: true})
}

func sendError(w http.ResponseWriter, msg string, status int) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   msg,
	})
}
