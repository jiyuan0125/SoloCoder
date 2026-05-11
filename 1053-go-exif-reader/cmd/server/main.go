package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/exif-reader/pkg/api"
	"github.com/exif-reader/pkg/exif"
)

type sessionStore struct {
	mu    sync.RWMutex
	data  map[string]*exif.EXIFData
}

func newSessionStore() *sessionStore {
	return &sessionStore{
		data: make(map[string]*exif.EXIFData),
	}
}

func (s *sessionStore) set(id string, data *exif.EXIFData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[id] = data
}

func (s *sessionStore) get(id string) (*exif.EXIFData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.data[id]
	return data, ok
}

var store = newSessionStore()

func main() {
	http.HandleFunc("/parse", handleParse)
	http.HandleFunc("/query", handleQuery)
	http.HandleFunc("/thumbnail", handleThumbnail)

	fmt.Println("Server starting on :8203")
	http.ListenAndServe(":8203", nil)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err string) {
	writeJSON(w, status, &api.ErrorResponse{Error: err})
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 50*1024*1024)

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read file from request")
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read file content")
		return
	}

	reader := bytes.NewReader(fileData)
	exifData, err := exif.Parse(reader)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse EXIF: %v", err))
		return
	}

	sessionID := fmt.Sprintf("%p", exifData)
	store.set(sessionID, exifData)

	metadata := exifData.FormatAll()

	response := &api.ParseResponse{
		Success:       true,
		Metadata:      metadata,
		HasThumbnail:  len(exifData.Thumbnail) > 0,
	}

	w.Header().Set("X-Session-Id", sessionID)
	writeJSON(w, http.StatusOK, response)
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	sessionID := r.Header.Get("X-Session-Id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "missing X-Session-Id header")
		return
	}

	exifData, ok := store.get(sessionID)
	if !ok {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	var req api.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	metadata := exifData.FormatAll()
	value, found := metadata[req.Field]

	response := &api.QueryResponse{
		Success: true,
		Field:   req.Field,
		Value:   value,
		Found:   found,
	}

	writeJSON(w, http.StatusOK, response)
}

func handleThumbnail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	sessionID := r.Header.Get("X-Session-Id")
	if sessionID == "" {
		writeError(w, http.StatusBadRequest, "missing X-Session-Id header")
		return
	}

	exifData, ok := store.get(sessionID)
	if !ok {
		writeError(w, http.StatusNotFound, "session not found")
		return
	}

	if len(exifData.Thumbnail) == 0 {
		writeError(w, http.StatusNotFound, "no thumbnail available")
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	w.Write(exifData.Thumbnail)
}
