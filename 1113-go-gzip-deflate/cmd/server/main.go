package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"archiver/pkg/archive"
	"archiver/pkg/api"
)

type Server struct {
	config    api.ServerConfig
	requestID int64
}

func NewServer(config api.ServerConfig) *Server {
	if config.Address == "" {
		config.Address = ":8430"
	}
	if config.MaxFiles == 0 {
		config.MaxFiles = 10000
	}
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Minute
	}
	return &Server{config: config}
}

func (s *Server) nextRequestID() string {
	return fmt.Sprintf("%d", atomic.AddInt64(&s.requestID, 1))
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) listFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.ListFilesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	files, err := listFiles(&req)
	if err != nil {
		resp := api.ListFilesResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := api.ListFilesResponse{
		Success: true,
		Files:   files,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) archiveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	requestID := s.nextRequestID()
	log.Printf("[%s] received archive request", requestID)
	startTime := time.Now()

	var req api.ArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[%s] failed to decode request: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	opts := &archive.Options{
		CompressionLevel: req.Compression,
		IncludePatterns:  req.IncludePattern,
		ExcludePatterns:  req.ExcludePattern,
		StripPrefix:      req.StripPrefix,
		AddPrefix:        req.AddPrefix,
		FollowSymlinks:   req.FollowSymlinks,
		IncludeEmptyDirs: req.IncludeEmpty,
	}

	if req.TimeStart != nil {
		opts.TimeStart = *req.TimeStart
	}
	if req.TimeEnd != nil {
		opts.TimeEnd = *req.TimeEnd
	}

	if len(req.Renames) > 0 {
		opts.Renamer = func(originalPath, archivePath string) string {
			if newName, ok := req.Renames[originalPath]; ok {
				return newName
			}
			return archivePath
		}
	}

	buf := &bytes.Buffer{}
	writer, err := archive.NewWriter(buf, opts)
	if err != nil {
		log.Printf("[%s] failed to create writer: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalFiles := 0

	for _, dir := range req.Directories {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			log.Printf("[%s] invalid directory path %s: %v", requestID, dir, err)
			continue
		}
		if err := writer.AddFilesFromRoot(absDir); err != nil {
			log.Printf("[%s] failed to add directory %s: %v", requestID, dir, err)
			writer.Close()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	for srcPath, arcPath := range req.Files {
		absSrc, err := filepath.Abs(srcPath)
		if err != nil {
			log.Printf("[%s] invalid source path %s: %v", requestID, srcPath, err)
			continue
		}

		info, err := os.Lstat(absSrc)
		if err != nil {
			log.Printf("[%s] failed to stat %s: %v", requestID, absSrc, err)
			continue
		}

		fileInfo := &archiveFileInfo{FileInfo: info}
		if !opts.ShouldInclude(arcPath, fileInfo) {
			continue
		}

		transformedPath := opts.TransformPath(arcPath)

		if info.IsDir() {
			entry := &archive.FileEntry{
				Path:    transformedPath,
				Mode:    int64(info.Mode()),
				ModTime: info.ModTime().Unix(),
				IsDir:   true,
			}
			if err := writer.AddFile(entry); err != nil {
				log.Printf("[%s] failed to add directory %s: %v", requestID, absSrc, err)
				writer.Close()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			totalFiles++
		} else if isSymlink(info) {
			if opts.FollowSymlinks {
				target, err := filepath.EvalSymlinks(absSrc)
				if err == nil {
					targetInfo, err := os.Stat(target)
					if err == nil && !targetInfo.IsDir() {
						f, err := os.Open(target)
						if err == nil {
							entry := &archive.FileEntry{
								Path:    transformedPath,
								Reader:  f,
								Size:    targetInfo.Size(),
								Mode:    int64(targetInfo.Mode()),
								ModTime: targetInfo.ModTime().Unix(),
								IsDir:   false,
							}
							if err := writer.AddFile(entry); err != nil {
								f.Close()
								log.Printf("[%s] failed to add file %s: %v", requestID, absSrc, err)
								writer.Close()
								http.Error(w, err.Error(), http.StatusInternalServerError)
								return
							}
							f.Close()
							totalFiles++
							continue
						}
					}
				}
			}

			linkTarget, err := os.Readlink(absSrc)
			if err != nil {
				log.Printf("[%s] failed to read symlink %s: %v", requestID, absSrc, err)
				continue
			}
			entry := &archive.FileEntry{
				Path:       transformedPath,
				Mode:       int64(info.Mode()),
				ModTime:    info.ModTime().Unix(),
				IsDir:      false,
				LinkTarget: linkTarget,
			}
			if err := writer.AddFile(entry); err != nil {
				log.Printf("[%s] failed to add symlink %s: %v", requestID, absSrc, err)
				writer.Close()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			totalFiles++
		} else {
			f, err := os.Open(absSrc)
			if err != nil {
				log.Printf("[%s] failed to open %s: %v", requestID, absSrc, err)
				continue
			}
			defer f.Close()

			entry := &archive.FileEntry{
				Path:    transformedPath,
				Reader:  f,
				Size:    info.Size(),
				Mode:    int64(info.Mode()),
				ModTime: info.ModTime().Unix(),
				IsDir:   false,
			}
			if err := writer.AddFile(entry); err != nil {
				log.Printf("[%s] failed to add file %s: %v", requestID, absSrc, err)
				writer.Close()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			totalFiles++
		}
	}

	if err := writer.Close(); err != nil {
		log.Printf("[%s] failed to close writer: %v", requestID, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileCount, writtenSize := writer.Stats()

	duration := time.Since(startTime)
	log.Printf("[%s] completed: %d files, %d bytes in %v", requestID, fileCount, writtenSize, duration)

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=archive_%s.tar.gz", requestID))
	w.Header().Set("X-File-Count", fmt.Sprintf("%d", fileCount))
	w.Header().Set("X-Written-Size", fmt.Sprintf("%d", writtenSize))
	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, buf); err != nil {
		log.Printf("[%s] failed to write response: %v", requestID, err)
	}
}

func listFiles(req *api.ListFilesRequest) ([]api.FileInfo, error) {
	var result []api.FileInfo

	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", req.Path)
	}

	var walkFunc filepath.WalkFunc = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == absPath {
			return nil
		}

		relPath, err := filepath.Rel(absPath, path)
		if err != nil {
			return err
		}

		if len(req.Patterns) > 0 {
			matched := false
			for _, pattern := range req.Patterns {
				if m, _ := filepath.Match(pattern, relPath); m {
					matched = true
					break
				}
				if strings.Contains(pattern, "**") || strings.Contains(relPath, string(filepath.Separator)) {
					base := filepath.Base(relPath)
					if m, _ := filepath.Match(pattern, base); m {
						matched = true
						break
					}
				}
			}
			if !matched {
				return nil
			}
		}

		result = append(result, api.FileInfo{
			Path:    relPath,
			Size:    info.Size(),
			Mode:    uint32(info.Mode()),
			ModTime: info.ModTime(),
			IsDir:   info.IsDir(),
		})

		if !req.Recursive && info.IsDir() {
			return filepath.SkipDir
		}

		return nil
	}

	if err := filepath.Walk(absPath, walkFunc); err != nil {
		return nil, err
	}

	return result, nil
}

type archiveFileInfo struct {
	os.FileInfo
}

func (f *archiveFileInfo) Mode() uint32 {
	return uint32(f.FileInfo.Mode())
}

func (f *archiveFileInfo) IsSymlink() bool {
	return f.FileInfo.Mode()&os.ModeSymlink != 0
}

func isSymlink(info os.FileInfo) bool {
	return info.Mode()&os.ModeSymlink != 0
}

func main() {
	var (
		addr      = flag.String("addr", ":8430", "server address")
		maxFiles  = flag.Int("max-files", 10000, "maximum number of files per archive")
		timeout   = flag.Duration("timeout", 5*time.Minute, "request timeout")
	)
	flag.Parse()

	config := api.ServerConfig{
		Address:  *addr,
		MaxFiles: *maxFiles,
		Timeout:  *timeout,
	}

	server := NewServer(config)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.healthHandler)
	mux.HandleFunc("/list", server.listFilesHandler)
	mux.HandleFunc("/archive", server.archiveHandler)

	log.Printf("server starting on %s", config.Address)
	log.Printf("max files per archive: %d", config.MaxFiles)
	log.Printf("request timeout: %v", config.Timeout)

	srv := &http.Server{
		Addr:         config.Address,
		Handler:      mux,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
