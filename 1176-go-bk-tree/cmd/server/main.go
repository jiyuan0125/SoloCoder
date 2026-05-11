package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"bktree-app/internal/bktree"
	"bktree-app/internal/models"
)

var tree = bktree.New()

func getPort() string {
	port := "8201"
	if envPort := os.Getenv("BK_TREE_PORT"); envPort != "" {
		port = envPort
	}
	flag.StringVar(&port, "port", port, "Server port")
	flag.Parse()
	return ":" + port
}

func main() {
	port := getPort()

	mux := http.NewServeMux()

	mux.HandleFunc("/add", handleAdd)
	mux.HandleFunc("/import", handleImport)
	mux.HandleFunc("/search", handleSearch)
	mux.HandleFunc("/remove", handleRemove)
	mux.HandleFunc("/info", handleInfo)

	log.Printf("Server listening on %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tree.Insert(req.Word)

	resp := models.AddResponse{Success: true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for _, word := range req.Words {
		tree.Insert(word)
	}

	resp := models.ImportResponse{
		Success:  true,
		Imported: len(req.Words),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	matches := tree.Search(req.Query, req.MaxDistance)
	searchMatches := make([]models.SearchMatch, len(matches))
	for i, m := range matches {
		searchMatches[i] = models.SearchMatch{
			Word:     m.Word,
			Distance: m.Distance,
		}
	}

	resp := models.SearchResponse{
		Success: true,
		Matches: searchMatches,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.RemoveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	removed := tree.Remove(req.Word)

	resp := models.RemoveResponse{
		Success: true,
		Removed: removed,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := models.InfoResponse{
		Success: true,
		Size:    tree.Size(),
		Depth:   tree.Depth(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
