package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ftp-client/common"
	"ftp-client/ftp"
)

type Handler struct {
	transferManager *TransferManager
}

func NewHandler(tm *TransferManager) *Handler {
	return &Handler{transferManager: tm}
}

func (h *Handler) HandleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config common.FTPConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mode := ftp.ModePassive
	if config.Mode == common.ModeActive {
		mode = ftp.ModeActive
	}

	client := ftp.NewClient(config.Host, config.Port, config.Username, config.Password, mode)
	if err := client.Connect(); err != nil {
		json.NewEncoder(w).Encode(common.SimpleResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	client.Close()

	json.NewEncoder(w).Encode(common.SimpleResponse{
		Success: true,
		Message: "Connection successful",
	})
}

func (h *Handler) HandleOperation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.FileOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	switch req.Operation {
	case common.OpListDir, common.OpChangeDir, common.OpMakeDir, common.OpRemoveDir, common.OpGetCurrentDir:
		h.handleSyncOperation(w, req)
	case common.OpDownload:
		h.handleDownload(w, req)
	case common.OpUpload:
		h.handleUpload(w, req)
	default:
		json.NewEncoder(w).Encode(common.SimpleResponse{
			Success: false,
			Error:   fmt.Sprintf("Unknown operation: %s", req.Operation),
		})
	}
}

func (h *Handler) handleSyncOperation(w http.ResponseWriter, req common.FileOperationRequest) {
	mode := ftp.ModePassive
	if req.Config.Mode == common.ModeActive {
		mode = ftp.ModeActive
	}

	client := ftp.NewClient(req.Config.Host, req.Config.Port, req.Config.Username, req.Config.Password, mode)
	if err := client.Connect(); err != nil {
		json.NewEncoder(w).Encode(common.SimpleResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	defer client.Close()

	switch req.Operation {
	case common.OpListDir:
		h.handleListDir(w, client, req)
	case common.OpChangeDir:
		h.handleChangeDir(w, client, req)
	case common.OpMakeDir:
		h.handleMakeDir(w, client, req)
	case common.OpRemoveDir:
		h.handleRemoveDir(w, client, req)
	case common.OpGetCurrentDir:
		h.handleGetCurrentDir(w, client)
	}
}

func (h *Handler) handleListDir(w http.ResponseWriter, client *ftp.Client, req common.FileOperationRequest) {
	var entries []ftp.FileEntry
	var err error

	if req.UseMLSD {
		entries, err = client.MLSD(req.RemotePath)
	} else {
		entries, err = client.List(req.RemotePath)
	}

	if err != nil {
		json.NewEncoder(w).Encode(common.ListDirResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	result := make([]common.FileEntry, len(entries))
	for i, e := range entries {
		result[i] = common.FileEntry{
			Name:        e.Name,
			IsDirectory: e.IsDirectory,
			Size:        e.Size,
			Modified:    e.Modified,
			Permissions: e.Permissions,
		}
	}

	json.NewEncoder(w).Encode(common.ListDirResponse{
		Success:    true,
		Entries:    result,
		CurrentDir: client.CurrentDir(),
	})
}

func (h *Handler) handleChangeDir(w http.ResponseWriter, client *ftp.Client, req common.FileOperationRequest) {
	err := client.ChangeDir(req.RemotePath)
	if err != nil {
		json.NewEncoder(w).Encode(common.SimpleResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	json.NewEncoder(w).Encode(common.SimpleResponse{
		Success: true,
		Message: "Directory changed",
	})
}

func (h *Handler) handleMakeDir(w http.ResponseWriter, client *ftp.Client, req common.FileOperationRequest) {
	err := client.MakeDir(req.RemotePath)
	if err != nil {
		json.NewEncoder(w).Encode(common.SimpleResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	json.NewEncoder(w).Encode(common.SimpleResponse{
		Success: true,
		Message: "Directory created",
	})
}

func (h *Handler) handleRemoveDir(w http.ResponseWriter, client *ftp.Client, req common.FileOperationRequest) {
	err := client.RemoveDir(req.RemotePath)
	if err != nil {
		json.NewEncoder(w).Encode(common.SimpleResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}
	json.NewEncoder(w).Encode(common.SimpleResponse{
		Success: true,
		Message: "Directory removed",
	})
}

func (h *Handler) handleGetCurrentDir(w http.ResponseWriter, client *ftp.Client) {
	json.NewEncoder(w).Encode(common.ListDirResponse{
		Success:    true,
		CurrentDir: client.CurrentDir(),
	})
}

func (h *Handler) handleDownload(w http.ResponseWriter, req common.FileOperationRequest) {
	transferID := generateID()

	ctx, cancel := context.WithCancel(context.Background())

	info := common.TransferInfo{
		ID:          transferID,
		Type:        common.TransferDownload,
		RemotePath:  req.RemotePath,
		LocalPath:   req.LocalPath,
		TotalSize:   0,
		Transferred: 0,
		Status:      common.StatusQueued,
		StartTime:   time.Now(),
	}

	job := &transferJob{
		info:   info,
		ctx:    ctx,
		cancel: cancel,
		config: req.Config,
	}

	h.transferManager.Add(transferID, job)

	go func() {
		info.Status = common.StatusRunning
		h.transferManager.UpdateInfo(transferID, info)

		mode := ftp.ModePassive
		if req.Config.Mode == common.ModeActive {
			mode = ftp.ModeActive
		}

		client := ftp.NewClient(req.Config.Host, req.Config.Port, req.Config.Username, req.Config.Password, mode)
		if err := client.Connect(); err != nil {
			info.Status = common.StatusFailed
			info.Error = fmt.Sprintf("connect failed: %v", err)
			info.EndTime = time.Now()
			h.transferManager.UpdateInfo(transferID, info)
			return
		}
		defer client.Close()

		if totalSize, err := client.GetSize(req.RemotePath); err == nil {
			info.TotalSize = totalSize
			h.transferManager.UpdateInfo(transferID, info)
		}

		callback := func(transferred, total int64) {
			h.transferManager.UpdateProgress(transferID, transferred)
		}

		var err error
		if req.Recursive {
			err = client.DownloadDirectoryWithContext(ctx, req.RemotePath, req.LocalPath, callback)
		} else {
			err = client.DownloadFileWithContext(ctx, req.RemotePath, req.LocalPath, callback)
		}

		currentInfo, _ := h.transferManager.Get(transferID)
		if err != nil {
			if err == context.Canceled {
				currentInfo.info.Status = common.StatusCancelled
			} else {
				currentInfo.info.Status = common.StatusFailed
				currentInfo.info.Error = err.Error()
			}
		} else {
			currentInfo.info.Status = common.StatusCompleted
			currentInfo.info.Transferred = currentInfo.info.TotalSize
		}
		currentInfo.info.EndTime = time.Now()
		h.transferManager.UpdateInfo(transferID, currentInfo.info)
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"transfer_id": transferID,
	})
}

func (h *Handler) handleUpload(w http.ResponseWriter, req common.FileOperationRequest) {
	transferID := generateID()

	ctx, cancel := context.WithCancel(context.Background())

	totalSize := int64(0)
	if info, err := getLocalFileSize(req.LocalPath); err == nil {
		totalSize = info
	}

	info := common.TransferInfo{
		ID:          transferID,
		Type:        common.TransferUpload,
		RemotePath:  req.RemotePath,
		LocalPath:   req.LocalPath,
		TotalSize:   totalSize,
		Transferred: 0,
		Status:      common.StatusQueued,
		StartTime:   time.Now(),
	}

	job := &transferJob{
		info:   info,
		ctx:    ctx,
		cancel: cancel,
		config: req.Config,
	}

	h.transferManager.Add(transferID, job)

	go func() {
		info.Status = common.StatusRunning
		h.transferManager.UpdateInfo(transferID, info)

		mode := ftp.ModePassive
		if req.Config.Mode == common.ModeActive {
			mode = ftp.ModeActive
		}

		client := ftp.NewClient(req.Config.Host, req.Config.Port, req.Config.Username, req.Config.Password, mode)
		if err := client.Connect(); err != nil {
			info.Status = common.StatusFailed
			info.Error = fmt.Sprintf("connect failed: %v", err)
			info.EndTime = time.Now()
			h.transferManager.UpdateInfo(transferID, info)
			return
		}
		defer client.Close()

		callback := func(transferred, total int64) {
			h.transferManager.UpdateProgress(transferID, transferred)
		}

		var err error
		if req.Recursive {
			err = client.UploadDirectoryWithContext(ctx, req.LocalPath, req.RemotePath, callback)
		} else {
			err = client.UploadFileWithContext(ctx, req.LocalPath, req.RemotePath, callback)
		}

		currentInfo, _ := h.transferManager.Get(transferID)
		if err != nil {
			if err == context.Canceled {
				currentInfo.info.Status = common.StatusCancelled
			} else {
				currentInfo.info.Status = common.StatusFailed
				currentInfo.info.Error = err.Error()
			}
		} else {
			currentInfo.info.Status = common.StatusCompleted
			currentInfo.info.Transferred = currentInfo.info.TotalSize
		}
		currentInfo.info.EndTime = time.Now()
		h.transferManager.UpdateInfo(transferID, currentInfo.info)
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"transfer_id": transferID,
	})
}

func (h *Handler) HandleQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	transfers := h.transferManager.List()
	json.NewEncoder(w).Encode(common.QueueResponse{
		Success:   true,
		Transfers: transfers,
	})
}

func (h *Handler) HandleCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.TransferCancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if h.transferManager.Cancel(req.TransferID) {
		json.NewEncoder(w).Encode(common.TransferCancelResponse{
			Success: true,
		})
	} else {
		json.NewEncoder(w).Encode(common.TransferCancelResponse{
			Success: false,
			Error:   "Transfer not found",
		})
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getLocalFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return info.Size(), nil
	}

	return getDirSize(path)
}

func getDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}
