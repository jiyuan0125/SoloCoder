package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"go-log-cleaner/common"
)

type FileScanner struct {
	targetDir      string
	fileExtensions []string
	recursive      bool
}

func NewFileScanner(targetDir string, fileExtensions []string, recursive bool) *FileScanner {
	return &FileScanner{
		targetDir:      targetDir,
		fileExtensions: fileExtensions,
		recursive:      recursive,
	}
}

func (fs *FileScanner) Scan() ([]common.FileInfo, error) {
	var files []common.FileInfo

	err := filepath.Walk(fs.targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if path != fs.targetDir && !fs.recursive {
				return filepath.SkipDir
			}
			return nil
		}

		if !fs.matchesExtension(path) {
			return nil
		}

		fileInfo, err := fs.getFileInfo(path, info)
		if err != nil {
			return err
		}

		files = append(files, fileInfo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %w", err)
	}

	return files, nil
}

func (fs *FileScanner) matchesExtension(path string) bool {
	if len(fs.fileExtensions) == 0 {
		return true
	}

	for _, ext := range fs.fileExtensions {
		if strings.HasSuffix(strings.ToLower(path), strings.ToLower(ext)) {
			return true
		}
	}
	return false
}

func (fs *FileScanner) getFileInfo(path string, info os.FileInfo) (common.FileInfo, error) {
	diskUsage, err := getDiskUsage(path)
	if err != nil {
		return common.FileInfo{}, err
	}

	isInUse, err := isFileInUse(path)
	if err != nil {
		return common.FileInfo{}, err
	}

	return common.FileInfo{
		Path:      path,
		Size:      info.Size(),
		DiskUsage: diskUsage,
		ModTime:   info.ModTime(),
		IsInUse:   isInUse,
	}, nil
}

func getDiskUsage(path string) (int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("failed to open file for disk usage check: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return 0, fmt.Errorf("failed to get file stat: %w", err)
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("failed to get system stat")
	}

	blockSize := int64(stat.Blksize)
	return stat.Blocks * blockSize / 512, nil
}

func isFileInUse(path string) (bool, error) {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		if os.IsPermission(err) {
			return true, nil
		}
		return false, fmt.Errorf("failed to check file usage: %w", err)
	}
	file.Close()
	return false, nil
}

func getDirUsage(path string) (int64, int64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, 0, fmt.Errorf("failed to get directory usage: %w", err)
	}

	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bfree * uint64(stat.Bsize)
	used := total - free

	return int64(used), int64(total), nil
}
