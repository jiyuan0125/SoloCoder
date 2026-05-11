package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/example/ldap-filter/api"
	"github.com/example/ldap-filter/ldapfilter"
)

func main() {
	port := flag.String("port", "8502", "HTTP server port")
	flag.Parse()

	if envPort := os.Getenv("LDAP_FILTER_PORT"); envPort != "" {
		*port = envPort
	}

	http.HandleFunc("/ldap/parse", handleParse)
	http.HandleFunc("/ldap/match", handleMatch)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("LDAP Filter Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(api.ParseResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.ParseResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	node, err := ldapfilter.Parse(req.Filter)
	if err != nil {
		json.NewEncoder(w).Encode(api.ParseResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(api.ParseResponse{
		Success: true,
		AST:     ldapfilter.ToASTJSON(node),
	})
}

func handleMatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(api.MatchResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(api.MatchResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	node, err := ldapfilter.Parse(req.Filter)
	if err != nil {
		json.NewEncoder(w).Encode(api.MatchResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	matched := ldapfilter.Match(node, req.Attributes)
	json.NewEncoder(w).Encode(api.MatchResponse{
		Success: true,
		Matched: matched,
	})
}
