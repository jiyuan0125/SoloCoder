package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/solocoder/resp-parser/api"
	"github.com/solocoder/resp-parser/resp"
)

type Server struct {
	processor *resp.CommandProcessor
}

func NewServer() *Server {
	return &Server{
		processor: resp.NewCommandProcessor(),
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleCommand(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	var commands [][]string
	var err error

	if strings.Contains(contentType, "application/json") {
		var req api.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON request: "+err.Error(), http.StatusBadRequest)
			return
		}
		commands = req.Commands
	} else {
		commands, err = resp.ParseCommandsFromReader(r.Body)
		if err != nil {
			if pe, ok := err.(*resp.ProtocolError); ok {
				http.Error(w, fmt.Sprintf("protocol error at byte %d: %s", pe.Pos, pe.Message), http.StatusBadRequest)
				return
			}
			http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	if len(commands) == 0 {
		http.Error(w, "no commands provided", http.StatusBadRequest)
		return
	}

	results := s.processor.ProcessBatch(commands)

	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/resp") || strings.Contains(contentType, "application/resp") {
		s.writeRESPResponse(w, results)
		return
	}

	s.writeJSONResponse(w, results)
}

func (s *Server) writeRESPResponse(w http.ResponseWriter, results []resp.Value) {
	w.Header().Set("Content-Type", "application/resp")
	enc := resp.NewEncoder(w)
	for _, v := range results {
		if err := enc.Encode(v); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
	if err := enc.Flush(); err != nil {
		http.Error(w, "failed to flush response", http.StatusInternalServerError)
	}
}

func (s *Server) writeJSONResponse(w http.ResponseWriter, results []resp.Value) {
	w.Header().Set("Content-Type", "application/json")
	
	response := api.Response{
		Results: make([]api.ResponseValue, len(results)),
	}
	
	for i, v := range results {
		response.Results[i] = convertToAPIValue(v)
	}
	
	json.NewEncoder(w).Encode(response)
}

func convertToAPIValue(v resp.Value) api.ResponseValue {
	val := api.ResponseValue{
		Type:   api.ValueTypeToString(int(v.Type)),
		IsNull: v.IsNull,
	}

	switch v.Type {
	case resp.TypeSimpleString, resp.TypeError, resp.TypeBulkString:
		val.Str = v.Str
	case resp.TypeInteger:
		num := v.Int
		val.Int = &num
	case resp.TypeArray:
		val.Array = make([]api.ResponseValue, len(v.Array))
		for j, elem := range v.Array {
			val.Array[j] = convertToAPIValue(elem)
		}
	}

	return val
}

func (s *Server) handleRaw(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	commands, err := resp.ParseCommandsFromReader(strings.NewReader(string(body)))
	if err != nil {
		if pe, ok := err.(*resp.ProtocolError); ok {
			http.Error(w, fmt.Sprintf("protocol error at byte %d: %s", pe.Pos, pe.Message), http.StatusBadRequest)
			return
		}
		http.Error(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if len(commands) == 0 {
		http.Error(w, "no commands provided", http.StatusBadRequest)
		return
	}

	results := s.processor.ProcessBatch(commands)

	w.Header().Set("Content-Type", "application/resp")
	enc := resp.NewEncoder(w)
	for _, v := range results {
		if err := enc.Encode(v); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
	if err := enc.Flush(); err != nil {
		http.Error(w, "failed to flush response", http.StatusInternalServerError)
	}
}

func main() {
	server := NewServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.handleHealth)
	mux.HandleFunc("/command", server.handleCommand)
	mux.HandleFunc("/raw", server.handleRaw)

	port := 8080
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("RESP Server listening on %s\n", addr)
	fmt.Printf("  POST /command  - Accept JSON or RESP, return JSON or RESP\n")
	fmt.Printf("  POST /raw     - Accept raw RESP or inline, return raw RESP\n")
	fmt.Printf("  GET  /health - Health check\n")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}
