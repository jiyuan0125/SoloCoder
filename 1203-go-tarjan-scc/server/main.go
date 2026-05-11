package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/user/tarjan-scc/pkg/api"
	"github.com/user/tarjan-scc/pkg/tarjan"
)

func getPort() int {
	port := 8080

	envPort := os.Getenv("PORT")
	if envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	flagPort := flag.Int("port", port, "Server port")
	flag.Parse()

	if *flagPort != 0 {
		port = *flagPort
	}

	return port
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		resp := api.GraphResponse{
			Success: false,
			Error:   "method not allowed, use POST",
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(resp)
		return
	}

	var req api.GraphRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := api.GraphResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	graph := tarjan.NewGraph()

	for _, node := range req.Nodes {
		graph.AddNode(node)
	}

	for _, edge := range req.Edges {
		graph.AddEdge(edge.From, edge.To)
	}

	sccs, err := graph.TarjanSCC()
	if err != nil {
		resp := api.GraphResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	responseItems := make([]api.SCCResponseItem, len(sccs))
	for i, scc := range sccs {
		responseItems[i] = api.SCCResponseItem{
			Nodes:   scc.Nodes,
			IsCycle: scc.IsCycle,
		}
	}

	resp := api.GraphResponse{
		Success: true,
		SCCs:    responseItems,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func main() {
	port := getPort()

	http.HandleFunc("/analyze", analyzeHandler)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Tarjan SCC server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
