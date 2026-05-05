package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"go-file-indexer/common"
	"go-file-indexer/indexer"
)

var idx *indexer.Indexer

func main() {
	var indexFile string
	var bindAddr string
	var initialDir string

	flag.StringVar(&indexFile, "index", "/tmp/file-index.gob", "Path to index file")
	flag.StringVar(&bindAddr, "bind", ":8080", "Bind address")
	flag.StringVar(&initialDir, "dir", "", "Initial directory to index")
	flag.Parse()

	idx = indexer.NewIndexer(indexFile)

	log.Println("Loading existing index...")
	if err := idx.LoadIndex(); err != nil {
		log.Printf("Warning: Could not load index: %v", err)
	}

	if initialDir != "" {
		log.Printf("Initial indexing of directory: %s", initialDir)
		if err := idx.BuildIndex(initialDir); err != nil {
			log.Printf("Error during initial indexing: %v", err)
		} else {
			idx.SaveIndex()
		}
	}

	http.HandleFunc("/index", indexHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/stats", statsHandler)
	http.HandleFunc("/update", updateHandler)

	log.Printf("Server starting on %s", bindAddr)
	if err := http.ListenAndServe(bindAddr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.IndexRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dirPath := req.Directory
	if dirPath == "" {
		http.Error(w, "Directory path required", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		http.Error(w, "Directory does not exist", http.StatusNotFound)
		return
	}

	log.Printf("Building index for directory: %s", dirPath)
	if err := idx.BuildIndex(dirPath); err != nil {
		resp := common.IndexResponse{
			Success: false,
			Message: fmt.Sprintf("Error building index: %v", err),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if err := idx.SaveIndex(); err != nil {
		log.Printf("Warning: Could not save index: %v", err)
	}

	stats := idx.Stats()
	resp := common.IndexResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully indexed %d files, %d words", stats["indexed_files"], stats["indexed_words"]),
	}
	json.NewEncoder(w).Encode(resp)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var query string

	if r.Method == http.MethodGet {
		query = r.URL.Query().Get("q")
	} else {
		var req common.SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		query = req.Query
	}

	if query == "" {
		resp := common.SearchResponse{
			Success: false,
			Message: "Query parameter required",
			Hits:    []common.SearchHit{},
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	hits, err := idx.Search(query)
	if err != nil {
		resp := common.SearchResponse{
			Success: false,
			Message: fmt.Sprintf("Search error: %v", err),
			Hits:    []common.SearchHit{},
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := common.SearchResponse{
		Success: true,
		Message: fmt.Sprintf("Found %d results", len(hits)),
		Hits:    hits,
	}
	json.NewEncoder(w).Encode(resp)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats := idx.Stats()
	resp := common.StatsResponse{
		Success:         true,
		IndexedFiles:    stats["indexed_files"].(int),
		IndexedWords:    stats["indexed_words"].(int),
		LastIndexedTime: stats["last_indexed_time"].(string),
	}
	json.NewEncoder(w).Encode(resp)
}

func updateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Performing incremental update...")
	if err := idx.IncrementalUpdate(); err != nil {
		resp := common.IndexResponse{
			Success: false,
			Message: fmt.Sprintf("Error during incremental update: %v", err),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if err := idx.SaveIndex(); err != nil {
		log.Printf("Warning: Could not save index: %v", err)
	}

	stats := idx.Stats()
	resp := common.IndexResponse{
		Success: true,
		Message: fmt.Sprintf("Incremental update complete. Now %d files, %d words", stats["indexed_files"], stats["indexed_words"]),
	}
	json.NewEncoder(w).Encode(resp)
}
