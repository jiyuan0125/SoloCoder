package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"gotestgen/api"
	"gotestgen/core"
)

type Server struct {
	history []*HistoryRecord
	mu      sync.RWMutex
}

type HistoryRecord struct {
	ID            string
	InputFileName string
	Structs       []string
	TestFunctions []string
	GeneratedAt   time.Time
}

func NewServer() *Server {
	return &Server{
		history: []*HistoryRecord{},
	}
}

func (s *Server) generateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.GenerateResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.FileName == "" || req.SourceCode == "" {
		writeJSON(w, http.StatusBadRequest, api.GenerateResponse{
			Success: false,
			Error:   "File name and source code are required",
		})
		return
	}

	result, err := core.GenerateTestCode(req.SourceCode, req.FileName)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.GenerateResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to generate test code: %v", err),
		})
		return
	}

	record := &HistoryRecord{
		ID:            fmt.Sprintf("%d", time.Now().UnixNano()),
		InputFileName: req.FileName,
		GeneratedAt:   time.Now(),
	}

	parsed, _ := core.ParseSource(req.SourceCode, req.FileName)
	for name := range parsed.Structs {
		record.Structs = append(record.Structs, name)
	}

	preview, _ := core.GeneratePreview(req.SourceCode, req.FileName)
	for _, tf := range preview.TestFunctions {
		record.TestFunctions = append(record.TestFunctions, tf.Name)
	}

	s.mu.Lock()
	s.history = append(s.history, record)
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, api.GenerateResponse{
		Success:  true,
		FileName: result.FileName,
		Code:     result.Code,
	})
}

func (s *Server) previewHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.PreviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.PreviewResponse{
			Success: false,
			Error:   "Invalid request body",
		})
		return
	}

	if req.FileName == "" || req.SourceCode == "" {
		writeJSON(w, http.StatusBadRequest, api.PreviewResponse{
			Success: false,
			Error:   "File name and source code are required",
		})
		return
	}

	result, err := core.GeneratePreview(req.SourceCode, req.FileName)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, api.PreviewResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to generate preview: %v", err),
		})
		return
	}

	testFuncs := make([]api.TestFuncInfo, len(result.TestFunctions))
	for i, tf := range result.TestFunctions {
		cases := make([]api.TestCaseInfo, len(tf.Cases))
		for j, tc := range tf.Cases {
			cases[j] = api.TestCaseInfo{
				Name:     tc.Name,
				Input:    tc.Input,
				Expected: tc.Expected,
			}
		}
		testFuncs[i] = api.TestFuncInfo{
			Name:     tf.Name,
			IsMethod: tf.IsMethod,
			Receiver: tf.Receiver,
			Cases:    cases,
		}
	}

	writeJSON(w, http.StatusOK, api.PreviewResponse{
		Success:   true,
		FileName:  result.FileName,
		Structs:   result.Structs,
		TestFuncs: testFuncs,
	})
}

func (s *Server) historyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	history := make([]api.HistoryEntry, len(s.history))
	for i, record := range s.history {
		history[i] = api.HistoryEntry{
			ID:            record.ID,
			InputFileName: record.InputFileName,
			Structs:       record.Structs,
			TestFunctions: record.TestFunctions,
			GeneratedAt:   record.GeneratedAt.Format(time.RFC3339),
		}
	}

	writeJSON(w, http.StatusOK, api.HistoryResponse{
		Success: true,
		History: history,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func main() {
	server := NewServer()

	http.HandleFunc("/api/generate", server.generateHandler)
	http.HandleFunc("/api/preview", server.previewHandler)
	http.HandleFunc("/api/history", server.historyHandler)

	port := "8080"
	log.Printf("Server starting on port %s...", port)
	log.Printf("Available endpoints:")
	log.Printf("  POST /api/generate  - Generate test code")
	log.Printf("  POST /api/preview   - Preview test functions")
	log.Printf("  GET  /api/history   - Get generation history")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
