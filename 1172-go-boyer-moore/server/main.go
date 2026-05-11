package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/solocoder/boyermoore/bmsearch"
	"github.com/solocoder/boyermoore/shared"
)

const DefaultPort = "8301"

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req shared.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	matches := bmsearch.Search(req.Text, req.Pattern)
	resp := shared.SearchResponse{
		Pattern: req.Pattern,
		TextLen: len(req.Text),
		Count:   len(matches),
		Matches: matches,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func preprocessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req shared.PreprocessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	prep, err := bmsearch.Preprocess(req.Pattern)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	bmBcMap := make(map[rune]int)
	m := len(req.Pattern)
	for i := 0; i < bmsearch.ASCIIMax; i++ {
		if prep.BmBc[i] != m {
			bmBcMap[rune(i)] = prep.BmBc[i]
		}
	}
	resp := shared.PreprocessResponse{
		Pattern:    prep.Pattern,
		PatternLen: m,
		BmBc:       bmBcMap,
		BmGs:       prep.BmGs,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getPort() string {
	port := flag.String("port", "", "listen port (overrides PORT env var)")
	flag.Parse()
	if *port != "" {
		return *port
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		return envPort
	}
	return DefaultPort
}

func main() {
	port := getPort()
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/preprocess", preprocessHandler)
	addr := ":" + port
	fmt.Printf("server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
