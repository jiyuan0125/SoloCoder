package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"vfs/common"
	"vfs/vfs"
)

type Server struct {
	vfs *vfs.VirtualFileSystem
}

func NewServer(vfs *vfs.VirtualFileSystem) *Server {
	return &Server{vfs: vfs}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) writeError(w http.ResponseWriter, status int, err error) {
	s.writeJSON(w, status, common.ErrorResponse{Error: err.Error()})
}

func (s *Server) handleListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	path := r.URL.Path
	if path == "/files" || path == "/files/" {
		path = "."
	} else {
		path = strings.TrimPrefix(path, "/files/")
		path = strings.TrimPrefix(path, "/files")
		if path == "" {
			path = "."
		}
	}

	info, err := s.vfs.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			s.writeError(w, http.StatusNotFound, err)
		} else {
			s.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	if !info.IsDir() {
		data, err := s.vfs.ReadFile(path)
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", info.Name()))
		w.Write(data)
		return
	}

	entries, err := s.vfs.ReadDir(path)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	response := common.ListDirResponse{Path: path}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		entryPath := path
		if path == "." {
			entryPath = entry.Name()
		} else {
			entryPath = strings.Join([]string{path, entry.Name()}, "/")
		}
		response.Entries = append(response.Entries, &common.FileInfo{
			Name:    entry.Name(),
			Path:    entryPath,
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}

	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		path = "."
	}

	tree, err := vfs.BuildTree(s.vfs, path)
	if err != nil {
		if os.IsNotExist(err) {
			s.writeError(w, http.StatusNotFound, err)
		} else {
			s.writeError(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := common.TreeResponse{
		Root: common.ConvertTreeNodeToResponse(tree),
	}

	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleDiff(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	diffs, err := vfs.Compare(s.vfs)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err)
		return
	}

	if diffs == nil {
		diffs = []vfs.FileDiff{}
	}

	response := common.DiffResponse{}
	for _, d := range diffs {
		response.Diffs = append(response.Diffs, common.ConvertFileDiffToResponse(d))
	}

	s.writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
		return
	}

	s.vfs.Reload()
	s.writeJSON(w, http.StatusOK, common.ReloadResponse{
		Success: true,
		Message: "Cache reloaded successfully",
	})
}

func getPort() string {
	port := os.Getenv("VFS_PORT")
	if port == "" {
		flagPort := flag.String("port", "8420", "Server port")
		flag.Parse()
		port = *flagPort
	}
	return port
}

func main() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable path: %v\n", err)
		os.Exit(1)
	}
	exeDir := filepath.Dir(exePath)

	assetsDir := os.Getenv("VFS_ASSETS_DIR")
	if assetsDir == "" {
		assetsDir = filepath.Join(exeDir, "assets")
	}

	devMode := os.Getenv("VFS_DEV") == "true"

	var diskDir string
	if devMode {
		if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
			diskDir = assetsDir
		}
	}

	subFS, err := fs.Sub(embedFS, "assets")
	if err != nil {
		fmt.Printf("Failed to get sub FS: %v\n", err)
		os.Exit(1)
	}

	config := vfs.Config{
		EmbedFS: subFS,
		DiskDir: diskDir,
	}

	vfsInstance, err := vfs.New(config)
	if err != nil {
		fmt.Printf("Failed to create VFS: %v\n", err)
		os.Exit(1)
	}

	server := NewServer(vfsInstance)

	mux := http.NewServeMux()
	mux.HandleFunc("/files", server.handleListFiles)
	mux.HandleFunc("/files/", server.handleListFiles)
	mux.HandleFunc("/tree", server.handleTree)
	mux.HandleFunc("/diff", server.handleDiff)
	mux.HandleFunc("/reload", server.handleReload)

	port := getPort()
	addr := fmt.Sprintf(":%s", port)

	fmt.Printf("VFS Server starting on %s\n", addr)
	fmt.Printf("Development mode: %v\n", devMode)
	fmt.Printf("Disk overlay: %s\n", diskDir)

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}
