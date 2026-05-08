package common

type MergeRequest struct {
	BaseContent string   `json:"base_content"`
	EnvContent  string   `json:"env_content"`
	Order       []string `json:"order,omitempty"`
}

type MergeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

type UploadRequest struct {
	FileName    string `json:"file_name"`
	Content     string `json:"content"`
	Environment string `json:"environment,omitempty"`
}

type UploadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	FileID  string `json:"file_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ListFilesResponse struct {
	Success bool        `json:"success"`
	Files   []FileEntry `json:"files,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type FileEntry struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Environment string `json:"environment,omitempty"`
	Size        int64  `json:"size"`
	ModifiedAt  string `json:"modified_at"`
}

type MergeMultipleRequest struct {
	FileIDs  []string `json:"file_ids"`
	BaseID   string   `json:"base_id"`
	EnvIDs   []string `json:"env_ids"`
}

type EditRequest struct {
	SessionID string `json:"session_id"`
	Content   string `json:"content"`
}

type EditResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Result  string `json:"result,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ExportRequest struct {
	Content string `json:"content"`
	Format  string `json:"format,omitempty"`
}

type ExportResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}
