package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"go-fuzz-corpus/api"
	"go-fuzz-corpus/corpus"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, api.ErrorResponse{Error: err.Error()})
}

func handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.RootPath == "" {
		req.RootPath = "."
	}

	stats, err := corpus.ScanRoot(req.RootPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, api.ScanResponse{Targets: stats})
}

func handleDedup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.DedupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.RootPath == "" {
		req.RootPath = "."
	}

	if req.TargetName == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("target name is required"))
		return
	}

	resp, err := corpus.DedupCorpus(req.TargetName, req.RootPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func handleMutate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.MutateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.RootPath == "" {
		req.RootPath = "."
	}

	if req.TargetName == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("target name is required"))
		return
	}

	if req.Count <= 0 {
		req.Count = 100
	}

	mutator := corpus.NewMutator()
	resp, err := mutator.MutateCorpus(req.TargetName, req.RootPath, req.Count)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.RootPath == "" {
		req.RootPath = "."
	}

	if req.TargetName == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("target name is required"))
		return
	}

	if req.CrashDir == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("crash directory is required"))
		return
	}

	resp, err := corpus.ImportFromCrashDir(req.CrashDir, req.TargetName, req.RootPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func handleListTargets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.ListTargetsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.RootPath == "" {
		req.RootPath = "."
	}

	stats, err := corpus.ScanRoot(req.RootPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, api.ListTargetsResponse{Targets: stats})
}

func handleGetTarget(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.GetTargetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.RootPath == "" {
		req.RootPath = "."
	}

	if req.TargetName == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("target name is required"))
		return
	}

	corpusDir, err := corpus.FindCorpusDirForTarget(req.RootPath, req.TargetName)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	stats, err := corpus.ScanCorpusDir(corpusDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, api.GetTargetResponse{Target: *stats})
}

func handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	var req api.DeleteFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if req.FilePath == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("file path is required"))
		return
	}

	resp, err := corpus.DeleteCorpusFile(req.FilePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func main() {
	port := flag.String("port", "", "Server port")
	flag.Parse()

	envPort := os.Getenv("FUZZ_CORPUS_PORT")
	finalPort := *port
	if finalPort == "" {
		finalPort = envPort
	}
	if finalPort == "" {
		finalPort = "8410"
	}

	http.HandleFunc("/api/scan", handleScan)
	http.HandleFunc("/api/dedup", handleDedup)
	http.HandleFunc("/api/mutate", handleMutate)
	http.HandleFunc("/api/import", handleImport)
	http.HandleFunc("/api/targets", handleListTargets)
	http.HandleFunc("/api/target", handleGetTarget)
	http.HandleFunc("/api/delete", handleDeleteFile)

	addr := fmt.Sprintf(":%s", finalPort)
	fmt.Printf("Fuzz corpus server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}
