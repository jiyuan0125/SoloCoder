package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	"go-log-cleaner/common"
)

type Cleaner struct {
	request common.CleanTaskRequest
}

func NewCleaner(request common.CleanTaskRequest) *Cleaner {
	return &Cleaner{
		request: request,
	}
}

func (c *Cleaner) Execute() (*common.CleanTaskResult, error) {
	result := &common.CleanTaskResult{
		Status:    common.TaskStatusRunning,
		StartTime: time.Now(),
	}

	scanner := NewFileScanner(
		c.request.TargetDir,
		c.request.FileExtensions,
		c.request.Recursive,
	)

	files, err := scanner.Scan()
	if err != nil {
		result.Status = common.TaskStatusFailed
		result.Error = err.Error()
		result.EndTime = time.Now()
		return result, fmt.Errorf("failed to scan files: %w", err)
	}

	filesToDelete, filesSkipped := c.filterFiles(files, result)
	result.FilesToDelete = filesToDelete
	result.FilesSkipped = filesSkipped

	for _, file := range filesToDelete {
		result.TotalSizeToDelete += file.DiskUsage
	}

	if c.request.DryRun || !c.request.Confirm {
		result.Status = common.TaskStatusCompleted
		result.EndTime = time.Now()
		return result, nil
	}

	filesDeleted, err := c.deleteFiles(filesToDelete, result)
	if err != nil {
		result.Status = common.TaskStatusFailed
		result.Error = err.Error()
		result.EndTime = time.Now()
		return result, err
	}

	result.FilesDeleted = filesDeleted
	for _, file := range filesDeleted {
		result.TotalSizeDeleted += file.DiskUsage
	}

	result.Status = common.TaskStatusCompleted
	result.EndTime = time.Now()
	return result, nil
}

func (c *Cleaner) filterFiles(files []common.FileInfo, result *common.CleanTaskResult) ([]common.FileInfo, []common.FileInfo) {
	var filesToDelete []common.FileInfo
	var filesSkipped []common.FileInfo

	var filteredFiles []common.FileInfo
	for _, file := range files {
		if file.IsInUse {
			filesSkipped = append(filesSkipped, file)
			continue
		}
		filteredFiles = append(filteredFiles, file)
	}

	if c.request.Mode&common.CleanModeTime != 0 {
		filteredFiles = c.filterByTime(filteredFiles)
	}

	if c.request.Mode&common.CleanModeCapacity != 0 {
		filteredFiles = c.filterByCapacity(filteredFiles, result)
	}

	if c.request.MaxDelete > 0 && len(filteredFiles) > c.request.MaxDelete {
		filteredFiles = filteredFiles[:c.request.MaxDelete]
	}

	filesToDelete = filteredFiles
	return filesToDelete, filesSkipped
}

func (c *Cleaner) filterByTime(files []common.FileInfo) []common.FileInfo {
	var filtered []common.FileInfo
	cutoffTime := time.Now().AddDate(0, 0, -c.request.RetainDays)

	for _, file := range files {
		if file.ModTime.Before(cutoffTime) {
			filtered = append(filtered, file)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ModTime.Before(filtered[j].ModTime)
	})

	return filtered
}

func (c *Cleaner) filterByCapacity(files []common.FileInfo, result *common.CleanTaskResult) []common.FileInfo {
	used, total, err := getDirUsage(c.request.TargetDir)
	if err != nil {
		return files
	}

	usageRatio := float64(used) / float64(total)
	if usageRatio <= c.request.CapacityThreshold {
		return []common.FileInfo{}
	}

	safeUsed := int64(float64(total) * c.request.SafeWaterLevel)
	needToFree := used - safeUsed

	if needToFree <= 0 {
		return []common.FileInfo{}
	}

	sortedFiles := make([]common.FileInfo, len(files))
	copy(sortedFiles, files)
	sort.Slice(sortedFiles, func(i, j int) bool {
		return sortedFiles[i].ModTime.Before(sortedFiles[j].ModTime)
	})

	var filtered []common.FileInfo
	var freedSpace int64

	for _, file := range sortedFiles {
		if freedSpace >= needToFree {
			break
		}
		filtered = append(filtered, file)
		freedSpace += file.DiskUsage
	}

	return filtered
}

func (c *Cleaner) deleteFiles(files []common.FileInfo, result *common.CleanTaskResult) ([]common.FileInfo, error) {
	var deleted []common.FileInfo

	for _, file := range files {
		err := os.Remove(file.Path)
		if err != nil {
			result.FilesSkipped = append(result.FilesSkipped, file)
			continue
		}
		deleted = append(deleted, file)
	}

	return deleted, nil
}
