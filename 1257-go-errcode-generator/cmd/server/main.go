package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/example/errcode-gen/pkg/codegen"
	"github.com/example/errcode-gen/pkg/common"
)

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := common.GenerateResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		}
		writeResponse(w, http.StatusBadRequest, resp)
		return
	}

	if strings.TrimSpace(req.YAMLContent) == "" {
		resp := common.GenerateResponse{
			Success: false,
			Error:   "yaml_content is required",
		}
		writeResponse(w, http.StatusBadRequest, resp)
		return
	}

	packageName := req.PackageName
	if packageName == "" {
		packageName = "errors"
	}

	result, err := codegen.GenerateFromYAML([]byte(req.YAMLContent), packageName)
	if err != nil {
		resp := common.GenerateResponse{
			Success: false,
			Error:   err.Error(),
		}
		writeResponse(w, http.StatusBadRequest, resp)
		return
	}

	resp := common.GenerateResponse{
		Success:  true,
		Code:     result.Code,
		Warnings: result.Warnings,
	}
	writeResponse(w, http.StatusOK, resp)
}

func writeResponse(w http.ResponseWriter, statusCode int, resp common.GenerateResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

func getPort() string {
	var port string
	flag.StringVar(&port, "port", "", "server port (e.g., :8080)")
	flag.StringVar(&port, "p", "", "server port (short)")
	flag.Parse()

	if port != "" {
		if !strings.HasPrefix(port, ":") {
			port = ":" + port
		}
		return port
	}

	port = os.Getenv("ERRCODE_PORT")
	if port != "" {
		if !strings.HasPrefix(port, ":") {
			port = ":" + port
		}
		return port
	}

	return ":8080"
}

func main() {
	port := getPort()
	server := NewServer()

	http.HandleFunc("/generate", server.handleGenerate)

	fmt.Printf("errcode-gen server listening on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
