package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"wire-di-gen/internal/api"
	"wire-di-gen/internal/wiregen"
)

const (
	defaultPort = 8080
	envPort     = "WIREGEN_PORT"
)

func main() {
	port := getPort()

	http.HandleFunc("/generate", handleGenerate)
	http.HandleFunc("/health", handleHealth)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("wire-di-gen server listening on %s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func getPort() int {
	if envVal := os.Getenv(envPort); envVal != "" {
		if p, err := strconv.Atoi(envVal); err == nil && p > 0 && p < 65536 {
			return p
		}
	}

	if len(os.Args) > 1 {
		if p, err := strconv.Atoi(os.Args[1]); err == nil && p > 0 && p < 65536 {
			return p
		}
	}

	return defaultPort
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req api.GenerateRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse request: %v", err), http.StatusBadRequest)
		return
	}

	if len(req.Files) == 0 {
		http.Error(w, "no files provided", http.StatusBadRequest)
		return
	}

	tempDir, err := os.MkdirTemp("", "wiregen-*")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create temp dir: %v", err), http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tempDir)

	for filename, content := range req.Files {
		filePath := filepath.Join(tempDir, filename)
		dir := filepath.Dir(filePath)
		if dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				http.Error(w, fmt.Sprintf("failed to create directory: %v", err), http.StatusInternalServerError)
				return
			}
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			http.Error(w, fmt.Sprintf("failed to write file: %v", err), http.StatusInternalServerError)
			return
		}
	}

	code, err := wiregen.Generate(tempDir)

	resp := api.GenerateResponse{}
	if err != nil {
		resp.Success = false
		resp.Error = err.Error()
		if wiregen.IsCircularDependencyError(err) {
			resp.CyclePath = wiregen.GetCircularDependencyPath(err)
		}
	} else {
		resp.Success = true
		resp.Code = code
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
