package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	"hungarian-match/internal/hungarian"
	"hungarian-match/pkg/api"
)

type matchCache struct {
	mu     sync.RWMutex
	result *hungarian.MatchResult
}

var cache = &matchCache{}

func main() {
	port := getPort()

	http.HandleFunc("/match", handleMatch)
	http.HandleFunc("/query", handleQuery)

	addr := ":" + port
	fmt.Printf("Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func getPort() string {
	var port int
	flag.IntVar(&port, "port", 8080, "Server port")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			return strconv.Itoa(p)
		}
	}
	return strconv.Itoa(port)
}

func handleMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	edges := make([]hungarian.Edge, 0, len(req.Edges))
	for _, e := range req.Edges {
		edges = append(edges, hungarian.Edge{
			Intern:   e.Intern,
			Position: e.Position,
		})
	}

	g := hungarian.NewBipartiteGraph(req.Interns, req.Positions, edges)
	result := g.MaxMatch()

	cache.mu.Lock()
	cache.result = &result
	cache.mu.Unlock()

	resp := api.MatchResponse{
		MaxMatch:    result.MaxMatch,
		Assignments: make([]api.Assignment, 0, len(result.Assignments)),
	}
	for _, a := range result.Assignments {
		resp.Assignments = append(resp.Assignments, api.Assignment{
			Intern:   a.Intern,
			Position: a.Position,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	cache.mu.RLock()
	result := cache.result
	cache.mu.RUnlock()

	resp := api.QueryResponse{Matched: false}
	if result != nil {
		if pos, found := result.GetInternAssignment(req.Intern); found {
			resp.Matched = true
			resp.Position = pos
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
