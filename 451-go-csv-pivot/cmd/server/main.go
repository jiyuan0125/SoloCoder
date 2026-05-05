// Package main implements the CSV join server.
// This server provides HTTP endpoints for joining CSV files.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"csvjoin/pkg/api"
	"csvjoin/pkg/csvjoin"
)

// Server represents the CSV join HTTP server.
type Server struct {
	uploadDir string
}

// NewServer creates a new CSV join server.
func NewServer(uploadDir string) (*Server, error) {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	return &Server{uploadDir: uploadDir}, nil
}

// HealthHandler handles health check requests.
func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.HealthResponse{Status: "ok"})
}

// JoinHandler handles CSV join requests.
func (s *Server) JoinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req api.JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	// Validate request
	if len(req.Files) < 2 {
		s.sendErrorResponse(w, http.StatusBadRequest, "at least 2 files required")
		return
	}
	if req.JoinKey == "" {
		s.sendErrorResponse(w, http.StatusBadRequest, "join key is required")
		return
	}

	// Create a unique directory for this request
	requestID := fmt.Sprintf("req_%d", time.Now().UnixNano())
	requestDir := filepath.Join(s.uploadDir, requestID)
	if err := os.MkdirAll(requestDir, 0755); err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "failed to create request directory: "+err.Error())
		return
	}
	defer os.RemoveAll(requestDir)

	// Prepare join options
	joinOptions := csvjoin.JoinOptions{
		JoinKey:  req.JoinKey,
		TrimKeys: req.TrimKeys,
	}

	switch req.JoinType {
	case api.InnerJoin:
		joinOptions.JoinType = csvjoin.InnerJoin
	case api.LeftJoin:
		joinOptions.JoinType = csvjoin.LeftJoin
	default:
		joinOptions.JoinType = csvjoin.LeftJoin // Default to left join
	}

	// Process each file
	for _, fileConfig := range req.Files {
		// Check if file exists
		if _, err := os.Stat(fileConfig.Path); os.IsNotExist(err) {
			s.sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("file %q does not exist", fileConfig.Path))
			return
		}

		joinFileOpts := csvjoin.FileOptions{
			Path:         fileConfig.Path,
			SkipHeader:   fileConfig.SkipHeader,
			ColumnNames:  fileConfig.ColumnNames,
			FileSuffix:   fileConfig.FileSuffix,
		}

		joinOptions.Files = append(joinOptions.Files, joinFileOpts)
	}

	// Generate output path
	outputPath := filepath.Join(requestDir, "output.csv")

	// Create joiner and perform join
	joiner, err := csvjoin.NewJoiner(joinOptions)
	if err != nil {
		s.sendErrorResponse(w, http.StatusBadRequest, "failed to create joiner: "+err.Error())
		return
	}

	result, err := joiner.Join(outputPath)
	if err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "join failed: "+err.Error())
		return
	}

	// Copy the output to a persistent location
	finalOutputPath := filepath.Join(s.uploadDir, fmt.Sprintf("result_%s.csv", requestID))
	if err := copyFile(outputPath, finalOutputPath); err != nil {
		s.sendErrorResponse(w, http.StatusInternalServerError, "failed to save result: "+err.Error())
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.JoinResponse{
		Success:    true,
		OutputPath: finalOutputPath,
		RowCount:   result.RowCount,
		Columns:    result.Columns,
	})
}

// UploadHandler handles file uploads.
func (s *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	uploadedFiles := make([]string, 0)

	// Get all files from the form
	for key := range r.MultipartForm.File {
		for _, fileHeader := range r.MultipartForm.File[key] {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "failed to open uploaded file: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// Create a unique filename
			safeFilename := filepath.Base(fileHeader.Filename)
			safeFilename = strings.ReplaceAll(safeFilename, "..", "")
			safeFilename = strings.ReplaceAll(safeFilename, string(filepath.Separator), "")

			outputPath := filepath.Join(s.uploadDir, fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeFilename))

			dst, err := os.Create(outputPath)
			if err != nil {
				http.Error(w, "failed to create output file: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, "failed to save uploaded file: "+err.Error(), http.StatusInternalServerError)
				return
			}

			uploadedFiles = append(uploadedFiles, outputPath)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"files":   uploadedFiles,
	})
}

// DownloadHandler handles file downloads.
func (s *Server) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "file parameter is required", http.StatusBadRequest)
		return
	}

	// Sanitize filename
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, "..", "")

	filePath := filepath.Join(s.uploadDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	// Set headers for download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	// Serve the file
	http.ServeFile(w, r, filePath)
}

// sendErrorResponse sends an error response.
func (s *Server) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(api.JoinResponse{
		Success: false,
		Error:   message,
	})
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func main() {
	// Get configuration from environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	// Create server
	server, err := NewServer(uploadDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Register handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.HealthHandler)
	mux.HandleFunc("/join", server.JoinHandler)
	mux.HandleFunc("/upload", server.UploadHandler)
	mux.HandleFunc("/download", server.DownloadHandler)

	// Start server
	addr := ":" + port
	fmt.Printf("Server starting on %s...\n", addr)
	fmt.Printf("Upload directory: %s\n", uploadDir)
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /health          - Health check")
	fmt.Println("  POST /join            - Join CSV files")
	fmt.Println("  POST /upload          - Upload CSV files")
	fmt.Println("  GET  /download?file=  - Download result file")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server failed: %v\n", err)
		os.Exit(1)
	}
}
