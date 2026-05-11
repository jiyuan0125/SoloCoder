package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"sqlparser/pkg/common"
	"sqlparser/pkg/parser"
)

func main() {
	port := getPort()

	mux := http.NewServeMux()
	mux.HandleFunc("/sql/parse", handleParse)
	mux.HandleFunc("/sql/match", handleMatch)
	mux.HandleFunc("/sql/validate", handleValidate)
	mux.HandleFunc("/sql/explain", handleExplain)

	addr := fmt.Sprintf(":%s", port)
	fmt.Printf("SQL WHERE Parser Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func getPort() string {
	var port string
	flag.StringVar(&port, "port", "", "Server port (default 8080)")
	flag.StringVar(&port, "p", "", "Server port (short)")
	flag.Parse()

	if port != "" {
		return port
	}

	port = os.Getenv("SQL_PARSER_PORT")
	if port != "" {
		return port
	}

	port = os.Getenv("PORT")
	if port != "" {
		return port
	}

	return "8080"
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.ParseResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req common.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ParseResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	where := strings.TrimSpace(req.Where)
	astJSON, err := parser.ParseToJSON(where)
	if err != nil {
		writeJSON(w, http.StatusOK, common.ParseResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.ParseResponse{
		Success: true,
		AST:     astJSON,
	})
}

func handleMatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.MatchResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req common.MatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.MatchResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	where := strings.TrimSpace(req.Where)
	match, err := parser.MatchData(where, req.Data)
	if err != nil {
		writeJSON(w, http.StatusOK, common.MatchResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.MatchResponse{
		Success: true,
		Match:   match,
	})
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.ValidateResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req common.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ValidateResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	where := strings.TrimSpace(req.Where)
	err := parser.Validate(where)
	if err != nil {
		writeJSON(w, http.StatusOK, common.ValidateResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.ValidateResponse{
		Success: true,
	})
}

func handleExplain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.ExplainResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req common.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ExplainResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	where := strings.TrimSpace(req.Where)
	_, steps, err := parser.Explain(where)
	if err != nil {
		writeJSON(w, http.StatusOK, common.ExplainResponse{
			Success: false,
			Steps:   steps,
			Error:   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, common.ExplainResponse{
		Success: true,
		Steps:   steps,
	})
}
