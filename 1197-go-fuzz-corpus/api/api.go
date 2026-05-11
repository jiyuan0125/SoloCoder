package api

import "time"

type CorpusFile struct {
	Path       string    `json:"path"`
	Hash       string    `json:"hash"`
	Size       int64     `json:"size"`
	AddedTime  time.Time `json:"added_time"`
	Source     string    `json:"source"`
	IsManual   bool      `json:"is_manual"`
}

type TargetStats struct {
	TargetName string       `json:"target_name"`
	FilePath   string       `json:"file_path"`
	FileCount  int          `json:"file_count"`
	TotalSize  int64        `json:"total_size"`
	Files      []CorpusFile `json:"files"`
}

type ScanRequest struct {
	RootPath string `json:"root_path"`
}

type ScanResponse struct {
	Targets []TargetStats `json:"targets"`
}

type DedupRequest struct {
	TargetName string `json:"target_name"`
	RootPath   string `json:"root_path"`
}

type DedupResponse struct {
	TargetName    string   `json:"target_name"`
	RemovedCount  int      `json:"removed_count"`
	RemovedFiles  []string `json:"removed_files"`
	RetainedCount int      `json:"retained_count"`
}

type MutateRequest struct {
	TargetName string `json:"target_name"`
	RootPath   string `json:"root_path"`
	Count      int    `json:"count"`
}

type MutateResponse struct {
	TargetName   string   `json:"target_name"`
	Generated    int      `json:"generated"`
	Failed       int      `json:"failed"`
	NewFiles     []string `json:"new_files"`
}

type ImportRequest struct {
	CrashDir   string `json:"crash_dir"`
	TargetName string `json:"target_name"`
	RootPath   string `json:"root_path"`
}

type ImportResponse struct {
	TargetName string   `json:"target_name"`
	Imported   int      `json:"imported"`
	Skipped    int      `json:"skipped"`
	NewFiles   []string `json:"new_files"`
}

type ListTargetsRequest struct {
	RootPath string `json:"root_path"`
}

type ListTargetsResponse struct {
	Targets []TargetStats `json:"targets"`
}

type GetTargetRequest struct {
	TargetName string `json:"target_name"`
	RootPath   string `json:"root_path"`
}

type GetTargetResponse struct {
	Target TargetStats `json:"target"`
}

type DeleteFileRequest struct {
	FilePath string `json:"file_path"`
}

type DeleteFileResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
