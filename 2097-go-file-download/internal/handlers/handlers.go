package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"filedownload/internal/middleware"
	"filedownload/internal/models"
	"filedownload/internal/storage"
)

const (
	MaxFileSize    = 100 * 1024 * 1024
	UploadDir      = "uploads"
)

type MultiStatusResponse struct {
	Results []models.UploadResult `json:"results"`
}

type GenerateLinkResponse struct {
	DownloadURL string    `json:"download_url"`
	LinkID      string    `json:"link_id"`
	ExpiresAt   time.Time `json:"expires_at"`
	MaxDownloads int      `json:"max_downloads"`
}

func generateID() string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%d-%s", time.Now().UnixNano(), "filedownload")))
	return hex.EncodeToString(h.Sum(nil))[:32]
}

func isValidFileName(name string) bool {
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return false
	}
	return true
}

func ensureUserDir(userID string) (string, error) {
	userDir := filepath.Join(UploadDir, userID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", err
	}
	return userDir, nil
}

func UploadFile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	err := r.ParseMultipartForm(MaxFileSize)
	if err != nil {
		http.Error(w, "Bad request or file too large", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	userDir, err := ensureUserDir(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	results := make([]models.UploadResult, 0, len(files))

	for _, fh := range files {
		result := models.UploadResult{
			OriginalName: fh.Filename,
			Success:      false,
		}

		if !isValidFileName(fh.Filename) {
			result.Error = "Invalid filename: contains path traversal characters"
			results = append(results, result)
			continue
		}

		if fh.Size > MaxFileSize {
			result.Error = "File too large"
			results = append(results, result)
			continue
		}

		fileID := generateID()
		storedFileName := fileID + filepath.Ext(fh.Filename)
		destPath := filepath.Join(userDir, storedFileName)

		src, err := fh.Open()
		if err != nil {
			result.Error = fmt.Sprintf("Failed to open file: %v", err)
			results = append(results, result)
			continue
		}

		dst, err := os.Create(destPath)
		if err != nil {
			src.Close()
			result.Error = fmt.Sprintf("Failed to create file: %v", err)
			results = append(results, result)
			continue
		}

		size, err := io.Copy(dst, src)
		src.Close()
		dst.Close()

		if err != nil {
			os.Remove(destPath)
			result.Error = fmt.Sprintf("Failed to write file: %v", err)
			results = append(results, result)
			continue
		}

		file := &models.File{
			ID:           fileID,
			UserID:       userID,
			FileName:     storedFileName,
			OriginalName: fh.Filename,
			Size:         size,
			UploadedAt:   time.Now(),
		}

		if err := storage.GlobalStorage.CreateFile(file); err != nil {
			os.Remove(destPath)
			result.Error = fmt.Sprintf("Failed to save metadata: %v", err)
			results = append(results, result)
			continue
		}

		result.Success = true
		result.FileID = fileID
		results = append(results, result)
	}

	if len(files) == 1 {
		if results[0].Success {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(results[0])
		} else {
			http.Error(w, results[0].Error, http.StatusBadRequest)
		}
		return
	}

	hasSuccess := false
	for _, r := range results {
		if r.Success {
			hasSuccess = true
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if hasSuccess {
		w.WriteHeader(http.StatusMultiStatus)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(MultiStatusResponse{Results: results})
}

func GenerateDownloadLink(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req models.GenerateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ExpiresIn <= 0 {
		http.Error(w, "ExpiresIn must be positive", http.StatusBadRequest)
		return
	}

	if req.MaxDownloads <= 0 {
		req.MaxDownloads = 1
	}

	file, err := storage.GlobalStorage.GetFileByID(req.FileID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if file == nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if file.UserID != userID {
		http.Error(w, "Forbidden: cannot access another user's file", http.StatusForbidden)
		return
	}

	linkID := generateID()
	link := &models.DownloadLink{
		ID:           linkID,
		FileID:       req.FileID,
		UserID:       userID,
		MaxDownloads: req.MaxDownloads,
		Downloaded:   0,
		ExpiresAt:    time.Now().Add(time.Duration(req.ExpiresIn) * time.Second),
		CreatedAt:    time.Now(),
	}

	if err := storage.GlobalStorage.CreateDownloadLink(link); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := GenerateLinkResponse{
		DownloadURL:  fmt.Sprintf("/api/download/%s", linkID),
		LinkID:       linkID,
		ExpiresAt:    link.ExpiresAt,
		MaxDownloads: link.MaxDownloads,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func DownloadFile(w http.ResponseWriter, r *http.Request) {
	linkID := r.PathValue("linkID")
	if linkID == "" {
		http.Error(w, "Missing link ID", http.StatusBadRequest)
		return
	}

	link, err := storage.GlobalStorage.GetDownloadLink(linkID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if link == nil {
		http.Error(w, "Link not found", http.StatusNotFound)
		return
	}

	if time.Now().After(link.ExpiresAt) {
		http.Error(w, "Link expired", http.StatusGone)
		return
	}

	if link.Downloaded >= link.MaxDownloads {
		http.Error(w, "Download limit exceeded", http.StatusGone)
		return
	}

	file, err := storage.GlobalStorage.GetFileByID(link.FileID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if file == nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	filePath := filepath.Join(UploadDir, link.UserID, file.FileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	ip := middleware.GetIPAddress(r)
	if r.Context().Value(middleware.UserIDKey) == nil {
		ip = r.RemoteAddr
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
	}

	if err := storage.GlobalStorage.IncrementDownloads(linkID); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	record := &models.DownloadRecord{
		ID:           generateID(),
		LinkID:       linkID,
		FileID:       link.FileID,
		UserID:       link.UserID,
		DownloadedAt: time.Now(),
		IPAddress:    ip,
	}
	if err := storage.GlobalStorage.CreateDownloadRecord(record); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.OriginalName))
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, f)
}

func DownloadByPath(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	filePath := r.PathValue("path")

	if filePath == "" {
		http.Error(w, "Missing file path", http.StatusBadRequest)
		return
	}

	fileName := filepath.Base(filePath)
	if !isValidFileName(fileName) {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	file, err := storage.GlobalStorage.GetFileByUserAndName(userID, fileName)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if file == nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	realPath := filepath.Join(UploadDir, userID, file.FileName)
	if _, err := os.Stat(realPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	f, err := os.Open(realPath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.OriginalName))
	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, f)
}

func GetUserFiles(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	keyword := r.URL.Query().Get("keyword")

	var files []*models.File
	var err error

	if keyword != "" {
		files, err = storage.GlobalStorage.SearchFiles(userID, keyword)
	} else {
		files, err = storage.GlobalStorage.GetFilesByUserID(userID)
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func GetFileInfo(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	isAdmin := middleware.GetIsAdmin(r)
	fileID := r.PathValue("fileID")

	file, err := storage.GlobalStorage.GetFileByID(fileID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if file == nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if file.UserID != userID && !isAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(file)
}

func AdminGetAllFiles(w http.ResponseWriter, r *http.Request) {
	files, err := storage.GlobalStorage.GetAllFiles()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func AdminGetAllDownloadRecords(w http.ResponseWriter, r *http.Request) {
	records, err := storage.GlobalStorage.GetAllDownloadRecords()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func AdminGetUserFiles(w http.ResponseWriter, r *http.Request) {
	targetUserID := r.PathValue("userID")
	if targetUserID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	keyword := r.URL.Query().Get("keyword")

	var files []*models.File
	var err error

	if keyword != "" {
		files, err = storage.GlobalStorage.SearchFiles(targetUserID, keyword)
	} else {
		files, err = storage.GlobalStorage.GetFilesByUserID(targetUserID)
	}

	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func AdminGetUserDownloadRecords(w http.ResponseWriter, r *http.Request) {
	targetUserID := r.PathValue("userID")
	if targetUserID == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	records, err := storage.GlobalStorage.GetDownloadRecordsByUserID(targetUserID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}
