package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tomlmerge/common"
	"tomlmerge/tomlmerge"
)

type storedFile struct {
	ID          string
	Name        string
	Environment string
	Content     string
	Size        int64
	ModifiedAt  time.Time
}

type editSession struct {
	ID          string
	Content     string
	BaseFileID  string
	EnvFileIDs  []string
	CreatedAt   time.Time
	LastUpdated time.Time
}

type Server struct {
	files     map[string]*storedFile
	filesMu   sync.RWMutex
	sessions  map[string]*editSession
	sessionsMu sync.RWMutex
	addr      string
	storageDir string
}

func NewServer(addr, storageDir string) *Server {
	s := &Server{
		files:      make(map[string]*storedFile),
		sessions:   make(map[string]*editSession),
		addr:       addr,
		storageDir: storageDir,
	}
	if storageDir != "" {
		os.MkdirAll(storageDir, 0755)
		s.loadFromStorage()
	}
	return s
}

func (s *Server) loadFromStorage() {
	if s.storageDir == "" {
		return
	}
	entries, err := os.ReadDir(s.storageDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}
		path := filepath.Join(s.storageDir, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		info, _ := entry.Info()
		id := strings.TrimSuffix(entry.Name(), ".toml")
		s.files[id] = &storedFile{
			ID:         id,
			Name:       entry.Name(),
			Content:    string(content),
			Size:       int64(len(content)),
			ModifiedAt: info.ModTime(),
		}
	}
}

func (s *Server) saveFile(id string, file *storedFile) {
	if s.storageDir == "" {
		return
	}
	path := filepath.Join(s.storageDir, id+".toml")
	os.WriteFile(path, []byte(file.Content), 0644)
}

func (s *Server) generateID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	var req common.UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.UploadResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}
	if req.Content == "" {
		writeJSON(w, http.StatusBadRequest, common.UploadResponse{
			Success: false,
			Error:   "content is required",
		})
		return
	}
	if _, err := tomlmerge.DecodeString(req.Content); err != nil {
		writeJSON(w, http.StatusBadRequest, common.UploadResponse{
			Success: false,
			Error:   "invalid TOML: " + err.Error(),
		})
		return
	}
	id := s.generateID("file")
	file := &storedFile{
		ID:          id,
		Name:        req.FileName,
		Environment: req.Environment,
		Content:     req.Content,
		Size:        int64(len(req.Content)),
		ModifiedAt:  time.Now(),
	}
	s.filesMu.Lock()
	s.files[id] = file
	s.filesMu.Unlock()
	s.saveFile(id, file)
	writeJSON(w, http.StatusOK, common.UploadResponse{
		Success: true,
		Message: "file uploaded successfully",
		FileID:  id,
	})
}

func (s *Server) handleListFiles(w http.ResponseWriter, r *http.Request) {
	s.filesMu.RLock()
	defer s.filesMu.RUnlock()
	files := make([]common.FileEntry, 0, len(s.files))
	for _, f := range s.files {
		files = append(files, common.FileEntry{
			ID:          f.ID,
			Name:        f.Name,
			Environment: f.Environment,
			Size:        f.Size,
			ModifiedAt:  f.ModifiedAt.Format(time.RFC3339),
		})
	}
	writeJSON(w, http.StatusOK, common.ListFilesResponse{
		Success: true,
		Files:   files,
	})
}

func (s *Server) handleMerge(w http.ResponseWriter, r *http.Request) {
	var req common.MergeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.MergeResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}
	result, err := tomlmerge.MergeStrings(req.BaseContent, req.EnvContent)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, common.MergeResponse{
			Success: false,
			Error:   "merge failed: " + err.Error(),
		})
		return
	}
	var buf bytes.Buffer
	if err := tomlmerge.Encode(&buf, result); err != nil {
		writeJSON(w, http.StatusInternalServerError, common.MergeResponse{
			Success: false,
			Error:   "encode failed: " + err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, common.MergeResponse{
		Success: true,
		Result:  buf.String(),
	})
}

func (s *Server) handleMergeMultiple(w http.ResponseWriter, r *http.Request) {
	var req common.MergeMultipleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.MergeResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}
	s.filesMu.RLock()
	defer s.filesMu.RUnlock()
	var result map[string]interface{}
	if req.BaseID != "" {
		baseFile, exists := s.files[req.BaseID]
		if !exists {
			writeJSON(w, http.StatusBadRequest, common.MergeResponse{
				Success: false,
				Error:   fmt.Sprintf("base file %s not found", req.BaseID),
			})
			return
		}
		var err error
		result, err = tomlmerge.DecodeString(baseFile.Content)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, common.MergeResponse{
				Success: false,
				Error:   "invalid base TOML: " + err.Error(),
			})
			return
		}
		for _, envID := range req.EnvIDs {
			envFile, exists := s.files[envID]
			if !exists {
				writeJSON(w, http.StatusBadRequest, common.MergeResponse{
					Success: false,
					Error:   fmt.Sprintf("env file %s not found", envID),
				})
				return
			}
			envMap, err := tomlmerge.DecodeString(envFile.Content)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, common.MergeResponse{
					Success: false,
					Error:   fmt.Sprintf("invalid env TOML %s: %s", envID, err.Error()),
				})
				return
			}
			result, err = tomlmerge.Merge(result, envMap)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, common.MergeResponse{
					Success: false,
					Error:   "merge failed: " + err.Error(),
				})
				return
			}
		}
	} else {
		if len(req.FileIDs) < 2 {
			writeJSON(w, http.StatusBadRequest, common.MergeResponse{
				Success: false,
				Error:   "at least 2 files required for multiple merge",
			})
			return
		}
		var err error
		var fileMap map[string]interface{}
		for i, fileID := range req.FileIDs {
			file, exists := s.files[fileID]
			if !exists {
				writeJSON(w, http.StatusBadRequest, common.MergeResponse{
					Success: false,
					Error:   fmt.Sprintf("file %s not found", fileID),
				})
				return
			}
			fileMap, err = tomlmerge.DecodeString(file.Content)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, common.MergeResponse{
					Success: false,
					Error:   fmt.Sprintf("invalid TOML %s: %s", fileID, err.Error()),
				})
				return
			}
			if i == 0 {
				result = fileMap
			} else {
				result, err = tomlmerge.Merge(result, fileMap)
				if err != nil {
					writeJSON(w, http.StatusInternalServerError, common.MergeResponse{
						Success: false,
						Error:   "merge failed: " + err.Error(),
					})
					return
				}
			}
		}
	}
	var buf bytes.Buffer
	if err := tomlmerge.Encode(&buf, result); err != nil {
		writeJSON(w, http.StatusInternalServerError, common.MergeResponse{
			Success: false,
			Error:   "encode failed: " + err.Error(),
		})
		return
	}
	sessionID := s.generateID("session")
	session := &editSession{
		ID:          sessionID,
		Content:     buf.String(),
		BaseFileID:  req.BaseID,
		EnvFileIDs:  req.EnvIDs,
		CreatedAt:   time.Now(),
		LastUpdated: time.Now(),
	}
	s.sessionsMu.Lock()
	s.sessions[sessionID] = session
	s.sessionsMu.Unlock()
	writeJSON(w, http.StatusOK, common.MergeResponse{
		Success: true,
		Result:  buf.String(),
	})
}

func (s *Server) handleEdit(w http.ResponseWriter, r *http.Request) {
	var req common.EditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.EditResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}
	if req.SessionID != "" {
		s.sessionsMu.Lock()
		defer s.sessionsMu.Unlock()
		session, exists := s.sessions[req.SessionID]
		if !exists {
			writeJSON(w, http.StatusBadRequest, common.EditResponse{
				Success: false,
				Error:   fmt.Sprintf("session %s not found", req.SessionID),
			})
			return
		}
		if req.Content != "" {
			if _, err := tomlmerge.DecodeString(req.Content); err != nil {
				writeJSON(w, http.StatusBadRequest, common.EditResponse{
					Success: false,
					Error:   "invalid TOML: " + err.Error(),
				})
				return
			}
			session.Content = req.Content
			session.LastUpdated = time.Now()
		}
		writeJSON(w, http.StatusOK, common.EditResponse{
			Success: true,
			Result:  session.Content,
		})
		return
	}
	if req.Content != "" {
		if _, err := tomlmerge.DecodeString(req.Content); err != nil {
			writeJSON(w, http.StatusBadRequest, common.EditResponse{
				Success: false,
				Error:   "invalid TOML: " + err.Error(),
			})
			return
		}
	}
	writeJSON(w, http.StatusOK, common.EditResponse{
		Success: true,
		Result: req.Content,
	})
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	var req common.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.ExportResponse{
			Success: false,
			Error:   "invalid request body: " + err.Error(),
		})
		return
	}
	if req.Content == "" {
		writeJSON(w, http.StatusBadRequest, common.ExportResponse{
			Success: false,
			Error:   "content is required",
		})
		return
	}
	config, err := tomlmerge.DecodeString(req.Content)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.ExportResponse{
			Success: false,
			Error:   "invalid TOML: " + err.Error(),
		})
		return
	}
	var buf bytes.Buffer
	if err := tomlmerge.Encode(&buf, config); err != nil {
		writeJSON(w, http.StatusInternalServerError, common.ExportResponse{
			Success: false,
			Error:   "encode failed: " + err.Error(),
		})
		return
	}
	w.Header().Set("Content-Type", "application/toml")
	w.Header().Set("Content-Disposition", "attachment; filename=merged.toml")
	io.Copy(w, &buf)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>TOML Merge Server</title>
</head>
<body>
	<h1>TOML Merge Server</h1>
	<p>Available endpoints:</p>
	<ul>
		<li>POST /api/upload - Upload TOML file</li>
		<li>GET /api/files - List uploaded files</li>
		<li>POST /api/merge - Merge two TOML contents</li>
		<li>POST /api/merge-multiple - Merge multiple stored files</li>
		<li>POST /api/edit - Edit merged configuration</li>
		<li>POST /api/export - Export configuration</li>
	</ul>
</body>
</html>`))
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/api/upload", s.handleUpload)
	mux.HandleFunc("/api/files", s.handleListFiles)
	mux.HandleFunc("/api/merge", s.handleMerge)
	mux.HandleFunc("/api/merge-multiple", s.handleMergeMultiple)
	mux.HandleFunc("/api/edit", s.handleEdit)
	mux.HandleFunc("/api/export", s.handleExport)
	fmt.Printf("Server starting on %s\n", s.addr)
	return http.ListenAndServe(s.addr, mux)
}

func main() {
	var addr string
	var storageDir string
	flag.StringVar(&addr, "addr", ":8100", "server address")
	flag.StringVar(&storageDir, "storage", "", "storage directory for uploaded files (optional)")
	flag.Parse()
	server := NewServer(addr, storageDir)
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
