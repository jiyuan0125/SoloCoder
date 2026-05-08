package main

import (
	"autocomplete/common"
	"autocomplete/trie"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

var t *trie.Trie

func initTrie() {
	t = trie.New()
	defaultWords := []string{
		"apple", "application", "apply", "app",
		"banana", "book", "boy",
		"cat", "car", "computer", "code",
		"dog", "door", "data",
		"elephant", "egg", "email",
		"fish", "food", "friend",
		"go", "golang", "good", "great",
		"hello", "help", "home",
		"internet", "idea", "input",
		"java", "javascript", "job",
		"key", "keyboard", "kitchen",
		"linux", "learn", "love",
		"music", "movie", "money",
		"node", "network", "new",
		"open", "orange", "order",
		"python", "program", "project",
		"query", "question", "quick",
		"run", "read", "room",
		"server", "search", "service",
		"test", "type", "time",
		"update", "user", "url",
		"value", "view", "video",
		"window", "word", "work", "world",
		"xml", "xcode",
		"youtube", "yellow",
		"zoo", "zero", "zone",
	}
	t.BuildFromList(defaultWords)
}

func main() {
	initTrie()

	http.HandleFunc("/search", handleSearch)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/delete", handleDelete)

	port := getEnv("PORT", "8080")
	addr := ":" + port
	fmt.Printf("Server starting on %s...\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultValue
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SearchRequest

	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	} else {
		req.Prefix = r.URL.Query().Get("prefix")
		limitStr := r.URL.Query().Get("limit")
		if limitStr != "" {
			var err error
			req.Limit, err = strconv.Atoi(limitStr)
			if err != nil {
				http.Error(w, "Invalid limit", http.StatusBadRequest)
				return
			}
		}
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	words := t.Search(req.Prefix, req.Limit)

	resp := common.SearchResponse{
		Words: words,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.AddRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Word == "" {
		http.Error(w, "Word is required", http.StatusBadRequest)
		return
	}

	t.Insert(req.Word)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.Response{
		Success: true,
		Message: fmt.Sprintf("Word '%s' added successfully", req.Word),
	})
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Word == "" {
		http.Error(w, "Word is required", http.StatusBadRequest)
		return
	}

	t.Delete(req.Word)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.Response{
		Success: true,
		Message: fmt.Sprintf("Word '%s' deleted successfully", req.Word),
	})
}
