package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"filedownload/internal/model"
	"filedownload/internal/service"
)

const (
	MaxUploadSize   = 100 * 1024 * 1024
	DefaultExpire   = 24
	DefaultMaxDL    = 0
	AdminUserID     = "admin"
)

type Handler struct {
	fileService *service.FileService
}

func NewHandler(fs *service.FileService) *Handler {
	return &Handler{fileService: fs}
}

func (h *Handler) getUserID(r *http.Request) string {
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}
	return userID
}

func (h *Handler) isAdmin(userID string) bool {
	return userID == AdminUserID
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

type UploadResult struct {
	Filename string `json:"filename,omitempty"`
	FileID   string `json:"file_id,omitempty"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
}

type UploadResponse struct {
	Results []UploadResult `json:"results"`
}

type GenerateLinkRequest struct {
	ExpireHours int `json:"expire_hours"`
	MaxDownloads int `json:"max_downloads"`
}

type GenerateLinkResponse struct {
	FileID   string `json:"file_id"`
	DownloadURL string `json:"download_url"`
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := h.getUserID(r)

	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "file too large, maximum 100MB")
		return
	}

	expireHours := DefaultExpire
	maxDownloads := DefaultMaxDL

	if v := r.FormValue("expire_hours"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			expireHours = val
		}
	}

	if expireHours <= 0 {
		writeError(w, http.StatusBadRequest, "expire time must be greater than 0")
		return
	}

	if v := r.FormValue("max_downloads"); v != "" {
		if val, err := strconv.Atoi(v); err == nil {
			maxDownloads = val
		}
	}

	if maxDownloads < 0 {
		writeError(w, http.StatusBadRequest, "max downloads must be non-negative")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		files = r.MultipartForm.File["file"]
	}

	if len(files) == 0 {
		writeError(w, http.StatusBadRequest, "no files uploaded")
		return
	}

	results := make([]UploadResult, 0, len(files))
	hasError := false

	for _, fh := range files {
		result := UploadResult{Filename: fh.Filename}

		if !h.fileService.Storage().IsValidFilename(fh.Filename) {
			result.Success = false
			result.Error = "invalid filename: contains path traversal characters"
			hasError = true
			results = append(results, result)
			continue
		}

		if fh.Size > MaxUploadSize {
			result.Success = false
			result.Error = fmt.Sprintf("file too large: %s (max 100MB)", fh.Filename)
			hasError = true
			results = append(results, result)
			continue
		}

		file, err := fh.Open()
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to open file: %v", err)
			hasError = true
			results = append(results, result)
			continue
		}

		storedID, size, err := h.fileService.Storage().SaveFile(userID, fh.Filename, file)
		file.Close()

		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("failed to save file: %v", err)
			hasError = true
			results = append(results, result)
			continue
		}

		fileRecord, err := h.fileService.CreateFile(userID, fh.Filename, expireHours, maxDownloads)
		if err != nil {
			h.fileService.Storage().DeleteFile(userID, storedID)
			result.Success = false
			result.Error = fmt.Sprintf("failed to create file record: %v", err)
			hasError = true
			results = append(results, result)
			continue
		}

		fileRecord.StoredPath = storedID
		fileRecord.Size = size

		if err := h.fileService.SaveFileToDB(fileRecord); err != nil {
			h.fileService.Storage().DeleteFile(userID, storedID)
			result.Success = false
			result.Error = fmt.Sprintf("failed to save metadata: %v", err)
			hasError = true
			results = append(results, result)
			continue
		}

		result.Success = true
		result.FileID = fileRecord.ID
		results = append(results, result)
	}

	response := UploadResponse{Results: results}

	if hasError && len(files) > 1 {
		writeJSON(w, http.StatusMultiStatus, response)
	} else if hasError {
		writeJSON(w, http.StatusInternalServerError, response)
	} else {
		writeJSON(w, http.StatusOK, response)
	}
}

func (h *Handler) DownloadByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	fileID := strings.TrimPrefix(r.URL.Path, "/api/download/")
	if fileID == "" {
		writeError(w, http.StatusBadRequest, "file id required")
		return
	}

	file, err := h.fileService.GetFileByID(fileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if file == nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	if !h.fileService.CanDownload(file) {
		writeError(w, http.StatusGone, "link expired or download limit reached")
		return
	}

	f, err := h.fileService.Storage().OpenFile(file.UserID, file.StoredPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to open file")
		return
	}
	defer f.Close()

	ip := getClientIP(r)
	userAgent := r.UserAgent()
	h.fileService.RecordDownload(file, ip, userAgent)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))
	io.Copy(w, f)
}

func (h *Handler) DownloadByPath(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := h.getUserID(r)
	prefix := "/api/files/"
	path := strings.TrimPrefix(r.URL.Path, prefix)
	if path == "" {
		writeError(w, http.StatusBadRequest, "file path required")
		return
	}

	if !h.fileService.Storage().IsValidFilename(path) {
		writeError(w, http.StatusBadRequest, "invalid filename")
		return
	}

	file, err := h.fileService.GetFileByPath(userID, path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if file == nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	if !h.fileService.CanDownload(file) {
		writeError(w, http.StatusGone, "link expired or download limit reached")
		return
	}

	f, err := h.fileService.Storage().OpenFile(file.UserID, file.StoredPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to open file")
		return
	}
	defer f.Close()

	ip := getClientIP(r)
	userAgent := r.UserAgent()
	h.fileService.RecordDownload(file, ip, userAgent)

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", file.Size))
	io.Copy(w, f)
}

func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := h.getUserID(r)

	if keyword := r.URL.Query().Get("q"); keyword != "" {
		files, err := h.fileService.SearchFiles(userID, keyword)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if files == nil {
			files = []*model.File{}
		}
		writeJSON(w, http.StatusOK, files)
		return
	}

	files, err := h.fileService.ListUserFiles(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if files == nil {
		files = []*model.File{}
	}
	writeJSON(w, http.StatusOK, files)
}

func (h *Handler) AdminListAllFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := h.getUserID(r)
	if !h.isAdmin(userID) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	files, err := h.fileService.ListAllFiles()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if files == nil {
		files = []*model.File{}
	}
	writeJSON(w, http.StatusOK, files)
}

func (h *Handler) AdminListDownloadRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID := h.getUserID(r)
	if !h.isAdmin(userID) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	records, err := h.fileService.ListAllDownloadRecords()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if records == nil {
		records = []*model.DownloadRecord{}
	}
	writeJSON(w, http.StatusOK, records)
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}
