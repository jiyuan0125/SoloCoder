package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"go-dep-analyzer/pkg/analyzer"
	"go-dep-analyzer/pkg/api"
)

type AppState struct {
	mu       sync.RWMutex
	analysis *analyzer.FullAnalysis
}

var state = &AppState{}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var req api.UploadRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if req.GoModContent == "" {
		writeError(w, http.StatusBadRequest, "go.mod content is required")
		return
	}

	analysis, err := analyzer.Analyze(req.GoModContent, req.GoSumContent)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Analysis failed: %v", err))
		return
	}

	state.mu.Lock()
	state.analysis = analysis
	state.mu.Unlock()

	writeJSON(w, http.StatusOK, &api.UploadResponse{
		Success: true,
		Message: "Analysis completed successfully",
	})
}

func requireAnalysis(w http.ResponseWriter) *analyzer.FullAnalysis {
	state.mu.RLock()
	defer state.mu.RUnlock()

	if state.analysis == nil {
		writeError(w, http.StatusPreconditionRequired, "No analysis data available. Please upload go.mod first.")
		return nil
	}
	return state.analysis
}

func handleGraph(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	analysis := requireAnalysis(w)
	if analysis == nil {
		return
	}

	resp := analyzer.ExportGraph(analysis.Graph)
	writeJSON(w, http.StatusOK, resp)
}

func handleCycles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	analysis := requireAnalysis(w)
	if analysis == nil {
		return
	}

	resp := analyzer.ExportCycles(analysis.Cycles)
	writeJSON(w, http.StatusOK, resp)
}

func handleTopology(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	analysis := requireAnalysis(w)
	if analysis == nil {
		return
	}

	resp := analyzer.ExportTopology(analysis.Topology)
	writeJSON(w, http.StatusOK, resp)
}

func handleDepths(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	analysis := requireAnalysis(w)
	if analysis == nil {
		return
	}

	resp := analyzer.ExportDepths(analysis.Depths)
	writeJSON(w, http.StatusOK, resp)
}

func handleConflicts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	analysis := requireAnalysis(w)
	if analysis == nil {
		return
	}

	resp := analyzer.ExportConflicts(analysis.Conflicts)
	writeJSON(w, http.StatusOK, resp)
}

func handleChains(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	analysis := requireAnalysis(w)
	if analysis == nil {
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}
	defer r.Body.Close()

	var req api.DependencyChainRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}

	if req.TargetModule == "" {
		writeError(w, http.StatusBadRequest, "target_module is required")
		return
	}

	chains := analyzer.FindDependencyChains(analysis.Graph, req.TargetModule)
	resp := analyzer.ExportDependencyChains(analysis.Graph, chains)
	writeJSON(w, http.StatusOK, resp)
}

func handleReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	analysis := requireAnalysis(w)
	if analysis == nil {
		return
	}

	resp := analyzer.ExportFullReport(analysis)
	writeJSON(w, http.StatusOK, resp)
}

func main() {
	http.HandleFunc("/upload", handleUpload)
	http.HandleFunc("/graph", handleGraph)
	http.HandleFunc("/cycles", handleCycles)
	http.HandleFunc("/topology", handleTopology)
	http.HandleFunc("/depths", handleDepths)
	http.HandleFunc("/conflicts", handleConflicts)
	http.HandleFunc("/chains", handleChains)
	http.HandleFunc("/report", handleReport)

	fmt.Println("Go Dependency Analyzer Server starting on :8080...")
	fmt.Println("Available endpoints:")
	fmt.Println("  POST /upload    - Upload go.mod and go.sum content")
	fmt.Println("  GET  /graph     - Get dependency graph")
	fmt.Println("  GET  /cycles    - Detect cyclic dependencies")
	fmt.Println("  GET  /topology  - Get topological sort result")
	fmt.Println("  GET  /depths    - Get dependency depth analysis")
	fmt.Println("  GET  /conflicts - Detect version conflicts")
	fmt.Println("  POST /chains    - Get dependency chains for a module")
	fmt.Println("  GET  /report    - Get full analysis report")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
