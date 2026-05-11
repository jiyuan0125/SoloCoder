package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"sexpr/internal/api"
	"sexpr/internal/sexpr"
)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) handleParse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var req api.ParseRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	exprs, err := sexpr.Parse(req.Input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := sexpr.ToJSONAll(exprs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to serialize result")
		return
	}

	resp := api.ParseResponse{
		Success: true,
		Result:  json.RawMessage(result),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleEval(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var req api.EvalRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	result, err := sexpr.EvalString(req.Input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := api.EvalResponse{
		Success: true,
		Result:  result.String(),
	}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleFormat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read request body")
		return
	}
	defer r.Body.Close()

	var req api.FormatRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	exprs, err := sexpr.Parse(req.Input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := api.FormatResponse{
		Success: true,
		Result:  sexpr.FormatAll(exprs),
	}
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	resp := api.ParseResponse{
		Success: false,
		Error:   msg,
	}
	json.NewEncoder(w).Encode(resp)
}

func main() {
	port := flag.String("port", "", "port to listen on (default: 8080)")
	flag.Parse()

	if *port == "" {
		if envPort := os.Getenv("SEXP_PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = "8080"
		}
	}

	server := NewServer()

	http.HandleFunc("/sexp/parse", server.handleParse)
	http.HandleFunc("/sexp/eval", server.handleEval)
	http.HandleFunc("/sexp/format", server.handleFormat)

	fmt.Printf("Server listening on port %s...\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
