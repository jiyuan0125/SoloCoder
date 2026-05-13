package main

import (
	"archive/tar"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"tar-archiver/archiver"
	"tar-archiver/db"
	"time"
)

type CreateArchiveRequest struct {
	SourceDir       string `json:"source_dir"`
	ArchivePath     string `json:"archive_path"`
	IncrementalDate string `json:"incremental_date,omitempty"`
	Password        string `json:"password,omitempty"`
}

type CreateArchiveResponse struct {
	Success       bool   `json:"success"`
	ArchivePath   string `json:"archive_path,omitempty"`
	FilesArchived int    `json:"files_archived,omitempty"`
	TotalBytes    int64  `json:"total_bytes,omitempty"`
	Error         string `json:"error,omitempty"`
}

type ExtractArchiveRequest struct {
	ArchivePath string `json:"archive_path"`
	TargetDir   string `json:"target_dir"`
	Password    string `json:"password,omitempty"`
}

type ExtractArchiveResponse struct {
	Success        bool     `json:"success"`
	FilesExtracted []string `json:"files_extracted,omitempty"`
	FilesSkipped   []string `json:"files_skipped,omitempty"`
	Error          string   `json:"error,omitempty"`
}

type ListArchiveRequest struct {
	ArchivePath string `json:"archive_path"`
	Password    string `json:"password,omitempty"`
}

type ListArchiveResponse struct {
	Success bool        `json:"success"`
	Entries []EntryInfo `json:"entries,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type EntryInfo struct {
	Name    string    `json:"name"`
	Size    int64     `json:"size"`
	Mode    int64     `json:"mode"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func createArchiveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		sendError(w, "Password cannot be empty", http.StatusBadRequest)
		return
	}

	var incrementalDate *time.Time
	if req.IncrementalDate != "" {
		t, err := time.Parse(time.RFC3339, req.IncrementalDate)
		if err != nil {
			sendError(w, fmt.Sprintf("Invalid incremental date: %v", err), http.StatusBadRequest)
			return
		}
		incrementalDate = &t
	}

	op := &db.ArchiveOperation{
		Operation:   "archive",
		ArchivePath: req.ArchivePath,
		SourcePath:  req.SourceDir,
		StartTime:   time.Now(),
		Status:      "running",
	}

	if err := db.CreateArchiveOperation(op); err != nil {
		sendError(w, fmt.Sprintf("Failed to create operation record: %v", err), http.StatusInternalServerError)
		return
	}

	progressChan := make(chan archiver.Progress)
	go func() {
		for progress := range progressChan {
			percentage := 0.0
			if progress.TotalBytes > 0 {
				percentage = float64(progress.ProcessedBytes) / float64(progress.TotalBytes) * 100
			}
			log.Printf("Archive progress: %.1f%% (%d / %d bytes)", percentage, progress.ProcessedBytes, progress.TotalBytes)
		}
	}()

	result, err := archiver.CreateArchive(req.SourceDir, req.ArchivePath, incrementalDate, req.Password, progressChan)
	close(progressChan)

	endTime := time.Now()
	op.EndTime = &endTime

	if err != nil {
		op.Status = "failed"
		op.ErrorMsg = err.Error()
		db.UpdateArchiveOperation(op)

		errorMsg := err.Error()
		if strings.Contains(errorMsg, "source directory does not exist") {
			sendError(w, "Source directory does not exist", http.StatusBadRequest)
		} else if strings.Contains(errorMsg, "existing file exists but is not a valid tar.gz") {
			sendError(w, "Existing file exists but is not a valid tar.gz", http.StatusBadRequest)
		} else {
			sendError(w, errorMsg, http.StatusInternalServerError)
		}
		return
	}

	op.Status = "completed"
	db.UpdateArchiveOperation(op)

	logCleanup(op.ID, fmt.Sprintf("Archive created successfully: %s", req.ArchivePath))

	response := CreateArchiveResponse{
		Success:       true,
		ArchivePath:   result.ArchivePath,
		FilesArchived: result.FilesArchived,
		TotalBytes:    result.TotalBytes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func extractArchiveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExtractArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		sendError(w, "Password cannot be empty", http.StatusBadRequest)
		return
	}

	op := &db.ArchiveOperation{
		Operation:   "extract",
		ArchivePath: req.ArchivePath,
		SourcePath:  req.TargetDir,
		StartTime:   time.Now(),
		Status:      "running",
	}

	if err := db.CreateArchiveOperation(op); err != nil {
		sendError(w, fmt.Sprintf("Failed to create operation record: %v", err), http.StatusInternalServerError)
		return
	}

	if err := archiver.VerifyArchiveIntegrity(req.ArchivePath, req.Password); err != nil {
		endTime := time.Now()
		op.EndTime = &endTime
		op.Status = "failed"
		op.ErrorMsg = fmt.Sprintf("Archive integrity check failed: %v", err)
		db.UpdateArchiveOperation(op)
		sendError(w, op.ErrorMsg, http.StatusBadRequest)
		return
	}

	result, err := archiver.ExtractArchive(req.ArchivePath, req.TargetDir, req.Password)

	endTime := time.Now()
	op.EndTime = &endTime

	if err != nil {
		op.Status = "failed"
		op.ErrorMsg = err.Error()
		db.UpdateArchiveOperation(op)
		sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	op.Status = "completed"
	db.UpdateArchiveOperation(op)

	logCleanup(op.ID, fmt.Sprintf("Extracted %d files to %s", len(result.FilesExtracted), req.TargetDir))

	response := ExtractArchiveResponse{
		Success:        true,
		FilesExtracted: result.FilesExtracted,
		FilesSkipped:   result.FilesSkipped,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func listArchiveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ListArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Password == "" {
		sendError(w, "Password cannot be empty", http.StatusBadRequest)
		return
	}

	entries, err := archiver.ListArchive(req.ArchivePath, req.Password)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	entryInfos := make([]EntryInfo, 0, len(entries))
	for _, entry := range entries {
		entryInfos = append(entryInfos, EntryInfo{
			Name:    entry.Name,
			Size:    entry.Size,
			Mode:    entry.Mode,
			ModTime: entry.ModTime,
			IsDir:   entry.Typeflag == tar.TypeDir,
		})
	}

	response := ListArchiveResponse{
		Success: true,
		Entries: entryInfos,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func logCleanup(operationID int64, details string) {
	cleanupLog := &db.CleanupLog{
		OperationID: operationID,
		CleanupTime: time.Now(),
		Details:     details,
	}

	if err := db.CreateCleanupLog(cleanupLog); err != nil {
		log.Printf("Warning: Failed to create cleanup log: %v", err)
	}
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func main() {
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.DB.Close()

	http.HandleFunc("/archive", createArchiveHandler)
	http.HandleFunc("/extract", extractArchiveHandler)
	http.HandleFunc("/list", listArchiveHandler)

	log.Println("Starting server on port 9501")
	log.Fatal(http.ListenAndServe(":9501", nil))
}
