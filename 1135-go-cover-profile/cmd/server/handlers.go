package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/solocoder/coverage-analyzer/pkg/api"
	"github.com/solocoder/coverage-analyzer/pkg/coverage"
)

type Server struct {
	store *AnalysisStore
}

func NewServer(store *AnalysisStore) *Server {
	return &Server{store: store}
}

func (s *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", 0, "")
		return
	}

	contentType := r.Header.Get("Content-Type")
	var content string

	if strings.Contains(contentType, "multipart/form-data") {
		file, _, err := r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read uploaded file", 0, err.Error())
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read file content", 0, err.Error())
			return
		}
		content = string(data)
	} else {
		var req api.UploadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body", 0, err.Error())
			return
		}
		content = req.Content
	}

	if content == "" {
		writeError(w, http.StatusBadRequest, "empty content", 0, "")
		return
	}

	profile, err := coverage.Parse(content)
	if err != nil {
		if pe, ok := err.(*coverage.ParseError); ok {
			writeError(w, http.StatusBadRequest, pe.Message, pe.Line, "")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error(), 0, "")
		return
	}

	result := coverage.NewAnalysisResult(profile)
	id := generateID()
	s.store.Save(id, result)

	resp := api.UploadResponse{
		ID:      id,
		Mode:    string(profile.Mode),
		Message: "successfully uploaded and parsed",
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) FunctionCoverageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", 0, "")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id parameter", 0, "")
		return
	}

	result, exists := s.store.Get(id)
	if !exists {
		writeError(w, http.StatusNotFound, "analysis not found", 0, "")
		return
	}

	report := result.CalculateFunctionCoverage()
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) PackageCoverageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", 0, "")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id parameter", 0, "")
		return
	}

	result, exists := s.store.Get(id)
	if !exists {
		writeError(w, http.StatusNotFound, "analysis not found", 0, "")
		return
	}

	report := result.CalculatePackageCoverage()
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) LineCoverageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", 0, "")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id parameter", 0, "")
		return
	}

	result, exists := s.store.Get(id)
	if !exists {
		writeError(w, http.StatusNotFound, "analysis not found", 0, "")
		return
	}

	report := result.CalculateLineCoverage()
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) SummaryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", 0, "")
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id parameter", 0, "")
		return
	}

	result, exists := s.store.Get(id)
	if !exists {
		writeError(w, http.StatusNotFound, "analysis not found", 0, "")
		return
	}

	summary := result.CalculateSummary()
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) DiffHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", 0, "")
		return
	}

	oldID := r.URL.Query().Get("old")
	newID := r.URL.Query().Get("new")

	if oldID == "" || newID == "" {
		writeError(w, http.StatusBadRequest, "missing old or new parameter", 0, "")
		return
	}

	oldResult, exists := s.store.Get(oldID)
	if !exists {
		writeError(w, http.StatusNotFound, "old analysis not found", 0, "")
		return
	}

	newResult, exists := s.store.Get(newID)
	if !exists {
		writeError(w, http.StatusNotFound, "new analysis not found", 0, "")
		return
	}

	oldReport := oldResult.CalculateFunctionCoverage()
	newReport := newResult.CalculateFunctionCoverage()

	diff := coverage.DiffReports(oldReport, newReport)
	diff.OldID = oldID
	diff.NewID = newID
	diff.OldSummary = *oldResult.CalculateSummary()
	diff.NewSummary = *newResult.CalculateSummary()

	writeJSON(w, http.StatusOK, diff)
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string, line int, details string) {
	resp := api.ErrorResponse{
		Error:   message,
		Line:    line,
		Details: details,
	}
	writeJSON(w, status, resp)
}
