package common

import (
	"time"
	"vfs/vfs"
)

type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	IsDir   bool      `json:"is_dir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
}

type ListDirResponse struct {
	Path    string      `json:"path"`
	Entries []*FileInfo `json:"entries"`
}

type TreeNodeResponse struct {
	Name     string               `json:"name"`
	IsDir    bool                 `json:"is_dir"`
	Path     string               `json:"path"`
	Children []*TreeNodeResponse  `json:"children,omitempty"`
}

type TreeResponse struct {
	Root *TreeNodeResponse `json:"root"`
}

type FileDiffResponse struct {
	Path    string `json:"path"`
	Status  string `json:"status"`
	Details string `json:"details"`
}

type DiffResponse struct {
	Diffs []*FileDiffResponse `json:"diffs"`
}

type ReloadResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func ConvertTreeNodeToResponse(node *vfs.TreeNode) *TreeNodeResponse {
	if node == nil {
		return nil
	}
	resp := &TreeNodeResponse{
		Name:  node.Name,
		IsDir: node.IsDir,
		Path:  node.Path,
	}
	for _, child := range node.Children {
		resp.Children = append(resp.Children, ConvertTreeNodeToResponse(child))
	}
	return resp
}

func ConvertFileDiffToResponse(diff vfs.FileDiff) *FileDiffResponse {
	return &FileDiffResponse{
		Path:    diff.Path,
		Status:  diff.Status,
		Details: diff.Details,
	}
}
