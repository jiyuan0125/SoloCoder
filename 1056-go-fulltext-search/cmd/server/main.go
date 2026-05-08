package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"fulltext-search/api"
	"fulltext-search/search"
)

type Server struct {
	engine *search.Engine
}

func NewServer() *Server {
	return &Server{
		engine: search.NewEngine(),
	}
}

func (s *Server) AddDocHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.AddDocRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DocID == "" {
		writeError(w, http.StatusBadRequest, "doc_id is required")
		return
	}

	s.engine.AddDoc(req.DocID, req.Content)

	resp := api.AddDocResponse{Success: true, Message: "document added successfully"}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) DeleteDocHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.DeleteDocRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DocID == "" {
		writeError(w, http.StatusBadRequest, "doc_id is required")
		return
	}

	s.engine.DeleteDoc(req.DocID)

	resp := api.DeleteDocResponse{Success: true, Message: "document deleted successfully"}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Query == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return
	}

	results, err := s.engine.Search(req.Query)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items := make([]api.SearchResultItem, 0, len(results))
	for _, r := range results {
		items = append(items, api.SearchResultItem{
			DocID: r.DocID,
			Score: r.Score,
		})
	}

	resp := api.SearchResponse{
		Success: true,
		Results: items,
	}
	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	resp := api.ErrorResponse{
		Success: false,
		Error:   message,
	}
	writeJSON(w, status, resp)
}

func main() {
	server := NewServer()

	http.HandleFunc("/add", server.AddDocHandler)
	http.HandleFunc("/delete", server.DeleteDocHandler)
	http.HandleFunc("/search", server.SearchHandler)

	port := ":8080"
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
