package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"sync"

	"ac-service/internal/api"
	"ac-service/pkg/ahocorasick"
)

type Server struct {
	matcher *ahocorasick.Matcher
	mu      sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		matcher: ahocorasick.NewMatcher(),
	}
}

func (s *Server) addPatternHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.AddPatternRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.matcher.Add(req.Pattern); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.AddPatternResponse{
		Success: true,
		Message: "Pattern added successfully",
	})
}

func (s *Server) addPatternsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.AddPatternsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	added, skipped, err := s.matcher.AddPatterns(req.Patterns)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.AddPatternsResponse{
		Success:      true,
		AddedCount:   added,
		SkippedCount: skipped,
		Message:      "Patterns added successfully",
	})
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(api.ErrorResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	s.mu.RLock()
	matches := s.matcher.Search(req.Text)
	s.mu.RUnlock()

	apiMatches := make([]api.Match, len(matches))
	for i, match := range matches {
		apiMatches[i] = api.Match{
			Pattern: match.Pattern,
			Start:   match.Start,
			End:     match.End,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.SearchResponse{
		Success: true,
		Matches: apiMatches,
	})
}

func (s *Server) dictStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	count := s.matcher.Count()
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.DictStatusResponse{
		Success: true,
		Count:  count,
	})
}

func (s *Server) clearHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	s.matcher.Clear()
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.ClearResponse{
		Success: true,
		Message: "Dictionary cleared successfully",
	})
}

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

	server := NewServer()

	http.HandleFunc("/api/pattern/add", server.addPatternHandler)
	http.HandleFunc("/api/pattern/add-batch", server.addPatternsHandler)
	http.HandleFunc("/api/search", server.searchHandler)
	http.HandleFunc("/api/dict", server.dictStatusHandler)
	http.HandleFunc("/api/clear", server.clearHandler)

	println("Server starting on port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
