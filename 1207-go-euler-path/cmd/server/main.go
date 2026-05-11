package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"euler-path/common"
	"euler-path/euler"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "server port")
	flag.Parse()

	if port == 0 {
		envPort := os.Getenv("PORT")
		if envPort != "" {
			fmt.Sscanf(envPort, "%d", &port)
		}
	}

	if port == 0 {
		port = 8080
	}

	http.HandleFunc("/euler-path", handleEulerPath)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleEulerPath(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(common.EulerPathResponse{
			Error: "method not allowed, use POST",
		})
		return
	}

	var req common.EulerPathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(common.EulerPathResponse{
			Error: fmt.Sprintf("invalid JSON: %v", err),
		})
		return
	}

	g := euler.NewGraph()
	for _, n := range req.Nodes {
		g.AddNode(euler.Node(n))
	}
	for _, e := range req.Edges {
		g.AddEdge(euler.Node(e.From), euler.Node(e.To))
	}

	var start euler.Node
	if req.Start != "" {
		start = euler.Node(req.Start)
	}

	result, err := g.FindEulerPath(start)

	resp := common.EulerPathResponse{
		HasEulerPath:    result != nil && result.HasEulerPath,
		HasEulerCircuit: result != nil && result.HasEulerCircuit,
	}

	if result != nil && result.Path != nil {
		path := make([]string, len(result.Path))
		for i, n := range result.Path {
			path[i] = string(n)
		}
		resp.Path = path
	}

	if err != nil {
		resp.Error = err.Error()
		w.WriteHeader(http.StatusBadRequest)
	}

	json.NewEncoder(w).Encode(resp)
}
