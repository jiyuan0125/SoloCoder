package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/fuzzysearch/api"
	"github.com/fuzzysearch/fuzzy"
)

type FuzzyServer struct {
	dict *fuzzy.Dictionary
}

func NewFuzzyServer() *FuzzyServer {
	return &FuzzyServer{
		dict: fuzzy.NewDictionary(),
	}
}

func (s *FuzzyServer) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/add", s.handleAdd)
	mux.HandleFunc("/remove", s.handleRemove)
	mux.HandleFunc("/import", s.handleImport)
	mux.HandleFunc("/search", s.handleSearch)
	mux.HandleFunc("/wildcard", s.handleWildcard)
	mux.HandleFunc("/dict", s.handleGetDict)

	return mux
}

func (s *FuzzyServer) Start(port string) error {
	addr := fmt.Sprintf(":%s", port)
	mux := s.setupRoutes()

	log.Printf("Fuzzy search server listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *FuzzyServer) handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.AddWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	success := s.dict.Add(req.Word)
	response := api.AddWordResponse{
		Success: success,
	}
	if !success {
		response.Message = "Word already exists"
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *FuzzyServer) handleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.RemoveWordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	success := s.dict.Remove(req.Word)
	response := api.RemoveWordResponse{
		Success: success,
	}
	if !success {
		response.Message = "Word not found"
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *FuzzyServer) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.ImportWordsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	imported := s.dict.Import(req.Words)
	response := api.ImportWordsResponse{
		Success:  true,
		Imported: imported,
		Total:    len(req.Words),
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *FuzzyServer) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Threshold < 0 {
		writeError(w, http.StatusBadRequest, "Threshold must be non-negative")
		return
	}

	results := s.dict.Search(req.Query, req.Threshold)
	apiResults := make([]api.SearchResult, len(results))
	for i, r := range results {
		apiResults[i] = api.SearchResult{
			Word:     r.Word,
			Distance: r.Distance,
		}
	}

	response := api.SearchResponse{
		Success: true,
		Results: apiResults,
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *FuzzyServer) handleWildcard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.WildcardSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Threshold < 0 {
		writeError(w, http.StatusBadRequest, "Threshold must be non-negative")
		return
	}

	results := s.dict.WildcardSearch(req.Pattern, req.Threshold)
	apiResults := make([]api.SearchResult, len(results))
	for i, r := range results {
		apiResults[i] = api.SearchResult{
			Word:     r.Word,
			Distance: r.Distance,
		}
	}

	response := api.WildcardSearchResponse{
		Success: true,
		Results: apiResults,
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *FuzzyServer) handleGetDict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	words := s.dict.GetAll()
	response := api.GetDictionaryResponse{
		Success: true,
		Size:    len(words),
		Words:   words,
	}

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	response := api.ErrorResponse{
		Success: false,
		Error:   message,
	}
	writeJSON(w, status, response)
}

func GetPort() string {
	if port := os.Getenv("FUZZY_PORT"); port != "" {
		return port
	}
	return "8080"
}

func ParsePort(portStr string) (string, error) {
	if portStr == "" {
		return GetPort(), nil
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", err
	}
	if port < 1 || port > 65535 {
		return "", fmt.Errorf("port must be between 1 and 65535")
	}
	return strconv.Itoa(port), nil
}
