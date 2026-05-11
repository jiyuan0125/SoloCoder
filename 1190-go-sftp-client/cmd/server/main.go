package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/example/sftp-client/pkg/api"
	"github.com/example/sftp-client/pkg/sftp"
)

type Connection struct {
	Config *api.ConnectionConfig
	Client *sftp.Client
}

type Server struct {
	connections map[string]*Connection
	mu          sync.RWMutex
	progress    map[string]*ProgressInfo
	progressMu  sync.RWMutex
}

type ProgressInfo struct {
	Operation   string
	TotalBytes  int64
	Transferred int64
}

func NewServer() *Server {
	return &Server{
		connections: make(map[string]*Connection),
		progress:    make(map[string]*ProgressInfo),
	}
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cfg api.ConnectionConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	client, err := sftp.NewClient(&sftp.Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		KeyPath:  cfg.KeyPath,
	})
	if err != nil {
		resp := api.ConnectionConfigResponse{
			Success: false,
			Message: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	id := fmt.Sprintf("conn_%d", len(s.connections)+1)
	cfg.ID = id

	s.mu.Lock()
	s.connections[id] = &Connection{
		Config: &cfg,
		Client: client,
	}
	s.mu.Unlock()

	resp := api.ConnectionConfigResponse{
		Success: true,
		ID:      id,
		Message: "Connection established successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleFileOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.FileOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	conn, ok := s.connections[req.ConnectionID]
	s.mu.RUnlock()

	if !ok {
		resp := api.FileOperationResponse{
			Success: false,
			Message: "Connection not found",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	client := conn.Client

	switch req.Operation {
	case "list":
		entries, err := client.ListDir(req.RemotePath)
		if err != nil {
			resp := api.FileOperationResponse{
				Success: false,
				Message: err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		apiEntries := make([]api.DirEntry, len(entries))
		for i, entry := range entries {
			apiEntries[i] = api.DirEntry{
				Filename:  entry.Filename,
				Longname:  entry.Longname,
				Type:      entry.Attrs.Type,
				Size:      entry.Attrs.Size,
				Perms:     entry.Attrs.Perms,
				IsDir:     entry.Attrs.Type == sftp.SSH_FILEXFER_TYPE_DIRECTORY,
				IsSymlink: entry.Attrs.Type == sftp.SSH_FILEXFER_TYPE_SYMLINK,
			}
		}

		resp := api.FileOperationResponse{
			Success: true,
			Data:    apiEntries,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	case "stat":
		attrs, err := client.Stat(req.RemotePath)
		if err != nil {
			resp := api.FileOperationResponse{
				Success: false,
				Message: err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		apiAttrs := api.FileAttrs{
			Type:      attrs.Type,
			Size:      attrs.Size,
			UID:       attrs.UID,
			GID:       attrs.GID,
			Perms:     attrs.Perms,
			Atime:     attrs.Atime,
			Mtime:     attrs.Mtime,
			IsDir:     attrs.Type == sftp.SSH_FILEXFER_TYPE_DIRECTORY,
			IsSymlink: attrs.Type == sftp.SSH_FILEXFER_TYPE_SYMLINK,
		}

		resp := api.FileOperationResponse{
			Success: true,
			Data:    apiAttrs,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	case "upload":
		progress := make(chan int64, 100)
		s.progressMu.Lock()
		s.progress[req.ConnectionID] = &ProgressInfo{
			Operation: "upload",
		}
		s.progressMu.Unlock()

		go func() {
			for p := range progress {
				s.progressMu.Lock()
				if info, ok := s.progress[req.ConnectionID]; ok {
					info.Transferred = p
				}
				s.progressMu.Unlock()
			}
		}()

		err := client.UploadFile(req.LocalPath, req.RemotePath, progress)
		close(progress)

		if err != nil {
			resp := api.FileOperationResponse{
				Success: false,
				Message: err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		resp := api.FileOperationResponse{
			Success: true,
			Message: "Upload completed successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	case "download":
		progress := make(chan int64, 100)
		s.progressMu.Lock()
		s.progress[req.ConnectionID] = &ProgressInfo{
			Operation: "download",
		}
		s.progressMu.Unlock()

		go func() {
			for p := range progress {
				s.progressMu.Lock()
				if info, ok := s.progress[req.ConnectionID]; ok {
					info.Transferred = p
				}
				s.progressMu.Unlock()
			}
		}()

		err := client.DownloadFile(req.RemotePath, req.LocalPath, progress)
		close(progress)

		if err != nil {
			resp := api.FileOperationResponse{
				Success: false,
				Message: err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		resp := api.FileOperationResponse{
			Success: true,
			Message: "Download completed successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)

	default:
		http.Error(w, "Unknown operation", http.StatusBadRequest)
	}
}

func (s *Server) handleViewFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	connID := r.URL.Query().Get("connection_id")
	remotePath := r.URL.Query().Get("path")

	if connID == "" || remotePath == "" {
		http.Error(w, "Missing connection_id or path", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	conn, ok := s.connections[connID]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return
	}

	client := conn.Client

	flags := uint32(sftp.SSH_FXF_READ)
	file, err := client.OpenFile(remotePath, flags)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.CloseFile(file)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", remotePath))

	var offset uint64 = 0
	bufSize := uint32(32 * 1024)

	for {
		data, err := client.ReadFile(file, offset, bufSize)
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		if len(data) == 0 {
			break
		}
		w.Write(data)
		offset += uint64(len(data))
	}
}

func (s *Server) handleProgress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	connID := r.URL.Query().Get("connection_id")
	if connID == "" {
		http.Error(w, "Missing connection_id", http.StatusBadRequest)
		return
	}

	s.progressMu.RLock()
	info, ok := s.progress[connID]
	s.progressMu.RUnlock()

	if !ok {
		resp := api.ProgressResponse{
			Success:      false,
			ConnectionID: connID,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	percent := 0.0
	if info.TotalBytes > 0 {
		percent = float64(info.Transferred) / float64(info.TotalBytes) * 100
	}

	resp := api.ProgressResponse{
		Success:      true,
		ConnectionID: connID,
		Operation:    info.Operation,
		TotalBytes:   info.TotalBytes,
		Transferred:  info.Transferred,
		Percent:      percent,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	port := flag.Int("port", 8204, "Server port")
	flag.Parse()

	if envPort := os.Getenv("SFTP_SERVER_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			*port = p
		}
	}

	server := NewServer()

	http.HandleFunc("/api/config", server.handleConfig)
	http.HandleFunc("/api/file", server.handleFileOperation)
	http.HandleFunc("/api/view", server.handleViewFile)
	http.HandleFunc("/api/progress", server.handleProgress)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("SFTP server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
