package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"network-planner/common"
	"network-planner/kruskal"
)

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Server port")
	flag.Parse()

	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}

	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("Invalid port: %s", port)
	}

	http.HandleFunc("/api/mst", handleMSTRequest)

	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func handleMSTRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.TopologyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if len(req.Nodes) == 0 {
		writeErrorResponse(w, http.StatusBadRequest, "Nodes cannot be empty")
		return
	}

	for _, edge := range req.Edges {
		if edge.Cost <= 0 {
			writeErrorResponse(w, http.StatusBadRequest, "Edge cost must be positive")
			return
		}
	}

	kruskalEdges := make([]kruskal.Edge, len(req.Edges))
	for i, e := range req.Edges {
		kruskalEdges[i] = kruskal.Edge{
			From: e.From,
			To:   e.To,
			Cost: e.Cost,
		}
	}

	uniqueEdges := kruskal.RemoveDuplicateEdges(kruskalEdges)

	graph := kruskal.Graph{
		Nodes: req.Nodes,
		Edges: uniqueEdges,
	}

	result := kruskal.KruskalMST(graph)

	if !result.IsConnected {
		response := common.MSTResponse{
			Success:  false,
			Message:  "Graph is not connected. Some devices are isolated.",
			Isolated: result.Isolated,
		}
		writeJSONResponse(w, http.StatusBadRequest, response)
		return
	}

	responseEdges := make([]common.Edge, len(result.Edges))
	for i, e := range result.Edges {
		responseEdges[i] = common.Edge{
			From: e.From,
			To:   e.To,
			Cost: e.Cost,
		}
	}

	response := common.MSTResponse{
		Success:   true,
		Edges:     responseEdges,
		TotalCost: result.TotalCost,
		Message:   "MST computed successfully",
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func writeErrorResponse(w http.ResponseWriter, status int, message string) {
	response := common.MSTResponse{
		Success: false,
		Message: message,
	}
	writeJSONResponse(w, status, response)
}

func writeJSONResponse(w http.ResponseWriter, status int, response common.MSTResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}
