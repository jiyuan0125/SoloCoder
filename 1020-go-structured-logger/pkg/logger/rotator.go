package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type RotatingFileWriter struct {
	mu         sync.Mutex
	basePath   string
	maxSize    int64
	maxBackups int
	file       *os.File
	fileSize   int64
}

func NewRotatingFileWriter(basePath string, maxSize int64, maxBackups int) (*RotatingFileWriter, error) {
	dir := filepath.Dir(basePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	w := &RotatingFileWriter{
		basePath:   basePath,
		maxSize:    maxSize,
		maxBackups: maxBackups,
	}
	if err := w.openOrCreate(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *RotatingFileWriter) openOrCreate() error {
	info, err := os.Stat(w.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			file, err := os.OpenFile(w.basePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return err
			}
			w.file = file
			w.fileSize = 0
			return nil
		}
		return err
	}
	file, err := os.OpenFile(w.basePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	w.file = file
	w.fileSize = info.Size()
	return nil
}

func (w *RotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n := len(p)
	if w.fileSize+int64(n) > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	written, err := w.file.Write(p)
	if written > 0 {
		w.fileSize += int64(written)
	}
	return written, err
}

func (w *RotatingFileWriter) rotate() error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
	}
	baseName := filepath.Base(w.basePath)
	dir := filepath.Dir(w.basePath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	files, err := filepath.Glob(filepath.Join(dir, nameWithoutExt+".*"+ext))
	if err != nil {
		return err
	}
	var numberedFiles []struct {
		path string
		num  int
	}
	for _, f := range files {
		base := filepath.Base(f)
		parts := strings.Split(base, ".")
		if len(parts) < 3 {
			continue
		}
		numStr := parts[len(parts)-2]
		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		numberedFiles = append(numberedFiles, struct {
			path string
			num  int
		}{path: f, num: num})
	}
	sort.Slice(numberedFiles, func(i, j int) bool {
		return numberedFiles[i].num > numberedFiles[j].num
	})
	for _, nf := range numberedFiles {
		if nf.num >= w.maxBackups-1 {
			os.Remove(nf.path)
		} else {
			newPath := filepath.Join(dir, fmt.Sprintf("%s.%d%s", nameWithoutExt, nf.num+1, ext))
			os.Rename(nf.path, newPath)
		}
	}
	rotatedPath := filepath.Join(dir, fmt.Sprintf("%s.1%s", nameWithoutExt, ext))
	if _, err := os.Stat(w.basePath); err == nil {
		if err := os.Rename(w.basePath, rotatedPath); err != nil {
			return err
		}
	}
	newFile, err := os.OpenFile(w.basePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	w.file = newFile
	w.fileSize = 0
	return nil
}

func (w *RotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
