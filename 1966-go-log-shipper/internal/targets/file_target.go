package targets

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"log-shipper/internal/types"
)

type FileTarget struct {
	id         string
	config     types.FileTargetConfig
	file       *os.File
	fileSize   int64
	fileCreate time.Time

	lastWriteTime time.Time
	successCount  int64
	failureCount  int64
	writeCount    int64
	startTime     time.Time

	mu     sync.Mutex
	stopCh chan struct{}
}

func NewFileTarget(id string, config types.FileTargetConfig) (*FileTarget, error) {
	t := &FileTarget{
		id:        id,
		config:    config,
		stopCh:    make(chan struct{}),
		startTime: time.Now(),
	}
	if err := t.openFile(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *FileTarget) ID() string { return t.id }
func (t *FileTarget) Type() string { return "file" }
func (t *FileTarget) Config() interface{} { return t.config }

func (t *FileTarget) Status() types.TargetStatus {
	elapsed := time.Since(t.startTime).Seconds()
	var rate float64
	if elapsed > 0 {
		rate = float64(atomic.LoadInt64(&t.writeCount)) / elapsed
	}
	return types.TargetStatus{
		ID:            t.id,
		Type:          "file",
		Config:        t.config,
		LastWriteTime: t.lastWriteTime,
		SuccessCount:  atomic.LoadInt64(&t.successCount),
		FailureCount:  atomic.LoadInt64(&t.failureCount),
		WriteRate:     rate,
	}
}

func (t *FileTarget) Start() error { return nil }

func (t *FileTarget) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.file != nil {
		t.file.Sync()
		t.file.Close()
	}
	return nil
}

func (t *FileTarget) Write(entries []*types.LogEntry) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if err := t.rotateIfNeeded(); err != nil {
		atomic.AddInt64(&t.failureCount, 1)
		log.Printf("[ERROR] file target %s: rotation failed: %v", t.id, err)
		return err
	}

	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		data = append(data, '\n')
		n, err := t.file.Write(data)
		if err != nil {
			atomic.AddInt64(&t.failureCount, 1)
			log.Printf("[ERROR] file target %s write failed (disk full?): %v", t.id, err)
			fmt.Fprintf(os.Stderr, "[ALERT] File target %s write failed: %v\n", t.id, err)
			return err
		}
		t.fileSize += int64(n)
		atomic.AddInt64(&t.writeCount, 1)
	}

	t.lastWriteTime = time.Now()
	atomic.AddInt64(&t.successCount, 1)
	t.file.Sync()
	return nil
}

func (t *FileTarget) openFile() error {
	dir := filepath.Dir(t.config.Path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(t.config.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	info, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}

	t.file = f
	t.fileSize = info.Size()
	t.fileCreate = info.ModTime()
	return nil
}

func (t *FileTarget) rotateIfNeeded() error {
	if t.config.Rotation.Strategy == "" {
		return nil
	}

	needRotate := false

	switch t.config.Rotation.Strategy {
	case "size":
		if t.config.Rotation.MaxSize > 0 && t.fileSize >= t.config.Rotation.MaxSize {
			needRotate = true
		}
	case "time":
		if t.config.Rotation.MaxAge > 0 && time.Since(t.fileCreate) >= t.config.Rotation.MaxAge {
			needRotate = true
		}
	}

	if !needRotate {
		return nil
	}

	t.file.Sync()
	t.file.Close()

	timestamp := time.Now().Format("20060102-150405")
	backupPath := t.config.Path + "." + timestamp
	if err := os.Rename(t.config.Path, backupPath); err != nil {
		return err
	}

	if err := t.openFile(); err != nil {
		return err
	}

	t.cleanupOldBackups()
	return nil
}

func (t *FileTarget) cleanupOldBackups() {
	if t.config.Rotation.MaxBackups <= 0 {
		return
	}

	pattern := t.config.Path + ".*"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	if len(matches) <= t.config.Rotation.MaxBackups {
		return
	}

	for i := 0; i < len(matches)-t.config.Rotation.MaxBackups; i++ {
		os.Remove(matches[i])
	}
}
