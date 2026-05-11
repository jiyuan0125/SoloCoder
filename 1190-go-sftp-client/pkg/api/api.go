package api

type ConnectionConfig struct {
	ID       string `json:"id,omitempty"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password,omitempty"`
	KeyPath  string `json:"key_path,omitempty"`
}

type ConnectionConfigResponse struct {
	Success bool   `json:"success"`
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

type FileOperationRequest struct {
	ConnectionID string `json:"connection_id"`
	Operation    string `json:"operation"`
	LocalPath    string `json:"local_path,omitempty"`
	RemotePath   string `json:"remote_path,omitempty"`
	SourcePath   string `json:"source_path,omitempty"`
	DestPath     string `json:"dest_path,omitempty"`
}

type FileOperationResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type DirEntry struct {
	Filename  string `json:"filename"`
	Longname  string `json:"longname"`
	Type      uint8  `json:"type"`
	Size      uint64 `json:"size"`
	Perms     uint32 `json:"perms"`
	IsDir     bool   `json:"is_dir"`
	IsSymlink bool   `json:"is_symlink"`
}

type FileAttrs struct {
	Type      uint8  `json:"type"`
	Size      uint64 `json:"size"`
	UID       uint32 `json:"uid"`
	GID       uint32 `json:"gid"`
	Perms     uint32 `json:"perms"`
	Atime     uint32 `json:"atime"`
	Mtime     uint32 `json:"mtime"`
	IsDir     bool   `json:"is_dir"`
	IsSymlink bool   `json:"is_symlink"`
}

type ProgressResponse struct {
	Success      bool   `json:"success"`
	ConnectionID string `json:"connection_id"`
	Operation    string `json:"operation,omitempty"`
	TotalBytes   int64  `json:"total_bytes"`
	Transferred  int64  `json:"transferred"`
	Percent      float64 `json:"percent"`
}
