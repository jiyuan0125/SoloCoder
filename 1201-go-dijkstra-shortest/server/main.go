package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/example/shortestpath/common"
	"github.com/example/shortestpath/shortestpath"
)

type Server struct {
	graph *shortestpath.Graph
	mu    sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		graph: shortestpath.NewGraph(),
	}
}

func (s *Server) importGraphHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var graphData common.GraphData
	if err := json.NewDecoder(r.Body).Decode(&graphData); err != nil {
		respondJSON(w, http.StatusBadRequest, common.ImportGraphResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid JSON: %v", err),
		})
		return
	}

	newGraph := shortestpath.NewGraph()

	for _, node := range graphData.Nodes {
		newGraph.AddNode(node)
	}

	for i, edge := range graphData.Edges {
		if err := newGraph.AddEdge(edge.From, edge.To, edge.Weight); err != nil {
			respondJSON(w, http.StatusBadRequest, common.ImportGraphResponse{
				Success: false,
				Message: fmt.Sprintf("Edge %d (from=%s, to=%s, weight=%.2f): %v", i, edge.From, edge.To, edge.Weight, err),
			})
			return
		}
	}

	s.mu.Lock()
	s.graph = newGraph
	s.mu.Unlock()

	respondJSON(w, http.StatusOK, common.ImportGraphResponse{
		Success: true,
		Message: fmt.Sprintf("Graph imported successfully: %d nodes, %d edges", newGraph.NodeCount(), len(graphData.Edges)),
	})
}

func (s *Server) shortestPathHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.ShortestPathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, common.ShortestPathResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid JSON: %v", err),
		})
		return
	}

	s.mu.RLock()
	graph := s.graph
	s.mu.RUnlock()

	result, err := graph.ShortestPath(req.Start, req.End)
	if err != nil {
		var message string
		switch err {
		case shortestpath.ErrEmptyGraph:
			message = "No graph has been imported yet"
		case shortestpath.ErrNodeNotFound:
			message = "Start or end node not found in graph"
		case shortestpath.ErrNoPath:
			message = "No path exists between the specified nodes"
		default:
			message = err.Error()
		}
		respondJSON(w, http.StatusOK, common.ShortestPathResponse{
			Success: false,
			Message: message,
		})
		return
	}

	respondJSON(w, http.StatusOK, common.ShortestPathResponse{
		Success:   true,
		TotalTime: result.TotalTime,
		Path:      result.Path,
		EdgeCount: result.EdgeCount,
	})
}

func (s *Server) allShortestPathsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.AllShortestPathsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, common.AllShortestPathsResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid JSON: %v", err),
		})
		return
	}

	s.mu.RLock()
	graph := s.graph
	s.mu.RUnlock()

	result, err := graph.AllShortestPaths(req.Start)
	if err != nil {
		var message string
		switch err {
		case shortestpath.ErrEmptyGraph:
			message = "No graph has been imported yet"
		case shortestpath.ErrNodeNotFound:
			message = "Start node not found in graph"
		default:
			message = err.Error()
		}
		respondJSON(w, http.StatusOK, common.AllShortestPathsResponse{
			Success: false,
			Message: message,
		})
		return
	}

	respondJSON(w, http.StatusOK, common.AllShortestPathsResponse{
		Success:   true,
		Distances: result.Distances,
	})
}

func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func main() {
	var port string
	flag.StringVar(&port, "port", "", "Server port (default: 8080, or SERVER_PORT env)")
	flag.Parse()

	if port == "" {
		port = os.Getenv("SERVER_PORT")
		if port == "" {
			port = "8080"
		}
	}

	server := NewServer()

	http.HandleFunc("/graph/import", server.importGraphHandler)
	http.HandleFunc("/path/shortest", server.shortestPathHandler)
	http.HandleFunc("/path/all", server.allShortestPathsHandler)

	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
