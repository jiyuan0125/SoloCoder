package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/example/callgraph/internal/api"
	"github.com/example/callgraph/internal/core"
	"github.com/example/callgraph/internal/model"
)

const (
	defaultPort = 8300
	envPort     = "CALLGRAPH_PORT"
)

type server struct {
	logger *log.Logger
}

func newServer() *server {
	return &server{
		logger: log.New(os.Stdout, "[callgraph-server] ", log.LstdFlags),
	}
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *server) analyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Printf("read request body: %v", err)
		writeJSONResponse(w, http.StatusBadRequest, api.AnalyzeResponse{
			Success: false,
			Message: fmt.Sprintf("read request: %v", err),
		})
		return
	}
	defer r.Body.Close()

	req, err := api.DecodeRequest(body)
	if err != nil {
		s.logger.Printf("decode request: %v", err)
		writeJSONResponse(w, http.StatusBadRequest, api.AnalyzeResponse{
			Success: false,
			Message: fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	s.logger.Printf("analyzing package: %s, files: %d, format: %s",
		req.PackagePath, len(req.Files), req.Format)

	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = "dot"
	}

	var files []model.File
	for _, f := range req.Files {
		files = append(files, model.File{
			Name:    f.Name,
			Content: f.Content,
		})
	}

	analyzer := core.NewAnalyzer(req.PackagePath, files)
	graph, err := analyzer.Analyze()
	if err != nil {
		s.logger.Printf("analyze error: %v", err)
		writeJSONResponse(w, http.StatusInternalServerError, api.AnalyzeResponse{
			Success: false,
			Message: fmt.Sprintf("analyze: %v", err),
		})
		return
	}

	output, err := core.FormatGraph(graph, format)
	if err != nil {
		s.logger.Printf("format error: %v", err)
		writeJSONResponse(w, http.StatusBadRequest, api.AnalyzeResponse{
			Success: false,
			Message: fmt.Sprintf("format: %v", err),
		})
		return
	}

	s.logger.Printf("analysis complete: %d nodes, %d edges", len(graph.Nodes), len(graph.Edges))

	writeJSONResponse(w, http.StatusOK, api.AnalyzeResponse{
		Success: true,
		Format:  format,
		Output:  output,
	})
}

func writeJSONResponse(w http.ResponseWriter, code int, resp api.AnalyzeResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	data, _ := api.EncodeResponse(resp)
	_, _ = w.Write(data)
}

func getPort() int {
	if v := os.Getenv(envPort); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			return p
		}
	}
	return defaultPort
}

func main() {
	portFlag := flag.Int("port", 0, "port to listen on (default: 8300, or $CALLGRAPH_PORT)")
	flag.Parse()

	port := getPort()
	if *portFlag > 0 {
		port = *portFlag
	}

	srv := newServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.health)
	mux.HandleFunc("/analyze", srv.analyze)

	addr := fmt.Sprintf(":%d", port)
	srv.logger.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
