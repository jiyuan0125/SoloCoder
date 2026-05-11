package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"rk-matcher/pkg/api"
	"rk-matcher/pkg/rabinkarp"
)

const (
	defaultPort = 8402
	version     = "1.0.0"
)

type Server struct {
	matcher *rabinkarp.Matcher
}

func NewServer() *Server {
	return &Server{
		matcher: rabinkarp.NewMatcher(),
	}
}

func (s *Server) handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	success := s.matcher.Add(req.Pattern)
	resp := api.AddResponse{
		Success: success,
	}
	if !success {
		if len(req.Pattern) == 0 {
			resp.Message = "Empty pattern not allowed"
		} else {
			resp.Message = "Pattern already exists"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	added, skipped := s.matcher.AddBatch(req.Patterns)
	resp := api.ImportResponse{
		Added:   added,
		Skipped: skipped,
		Message: fmt.Sprintf("Added %d patterns, skipped %d", added, skipped),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	matches := s.matcher.Search(req.Text)
	searchMatches := make([]api.SearchMatch, len(matches))
	for i, m := range matches {
		searchMatches[i] = api.SearchMatch{
			Pattern: m.Pattern,
			Index:   m.Index,
		}
	}

	resp := api.SearchResponse{
		Matches: searchMatches,
		Count:   len(matches),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handlePatterns(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	patterns := s.matcher.Patterns()
	byLength := s.matcher.PatternsByLength()

	patternList := make([]api.Pattern, 0, len(patterns))
	for _, p := range patterns {
		patternList = append(patternList, api.Pattern{
			Pattern: p,
			Length:  len(p),
		})
	}

	resp := api.PatternsResponse{
		Patterns: patternList,
		ByLength: byLength,
		Total:    len(patterns),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := s.matcher.CollisionStats()
	buckets := make(map[string]int, len(stats.HashBuckets))
	for k, v := range stats.HashBuckets {
		buckets[strconv.FormatUint(k, 10)] = v
	}

	resp := api.CollisionStatsResponse{
		TotalChecks:    stats.TotalChecks,
		FalsePositives: stats.FalsePositives,
		HashBuckets:    buckets,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	patterns := s.matcher.Patterns()
	byLength := s.matcher.PatternsByLength()
	stats := s.matcher.CollisionStats()

	patternList := make([]api.Pattern, 0, len(patterns))
	for _, p := range patterns {
		patternList = append(patternList, api.Pattern{
			Pattern: p,
			Length:  len(p),
		})
	}

	buckets := make(map[string]int, len(stats.HashBuckets))
	for k, v := range stats.HashBuckets {
		buckets[strconv.FormatUint(k, 10)] = v
	}

	resp := api.StatusResponse{
		Version: version,
		Patterns: api.PatternsResponse{
			Patterns: patternList,
			ByLength: byLength,
			Total:    len(patterns),
		},
		Stats: api.CollisionStatsResponse{
			TotalChecks:    stats.TotalChecks,
			FalsePositives: stats.FalsePositives,
			HashBuckets:    buckets,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleHash(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.HashRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var h uint64
	var success bool
	var message string

	if req.Length > 0 {
		h, success = rabinkarp.HashWithLen(req.Text, req.Length)
		if !success {
			message = fmt.Sprintf("Invalid length: %d for text of length %d", req.Length, len(req.Text))
		}
	} else {
		h = rabinkarp.Hash(req.Text)
		success = true
	}

	resp := api.HashResponse{
		Hash:    strconv.FormatUint(h, 10),
		Success: success,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getPort() int {
	portStr := os.Getenv("PORT")
	if portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 && p < 65536 {
			return p
		}
	}

	var port int
	flag.IntVar(&port, "port", defaultPort, "Server port")
	flag.IntVar(&port, "p", defaultPort, "Server port (shorthand)")
	flag.Parse()

	if port <= 0 || port >= 65536 {
		return defaultPort
	}
	return port
}

func main() {
	port := getPort()
	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/add", server.handleAdd)
	mux.HandleFunc("/import", server.handleImport)
	mux.HandleFunc("/search", server.handleSearch)
	mux.HandleFunc("/patterns", server.handlePatterns)
	mux.HandleFunc("/stats", server.handleStats)
	mux.HandleFunc("/status", server.handleStatus)
	mux.HandleFunc("/hash", server.handleHash)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Rabin-Karp Matcher Server v%s\n", version)
	fmt.Printf("Listening on %s\n", addr)
	fmt.Printf("Endpoints:\n")
	fmt.Printf("  POST /add      - Add a pattern\n")
	fmt.Printf("  POST /import   - Import patterns in batch\n")
	fmt.Printf("  POST /search   - Search text for patterns\n")
	fmt.Printf("  POST /hash     - Compute hash of a string\n")
	fmt.Printf("  GET  /patterns - List all patterns\n")
	fmt.Printf("  GET  /stats    - Show collision statistics\n")
	fmt.Printf("  GET  /status   - Show full status\n")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
