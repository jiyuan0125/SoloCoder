package logger

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"log-rotate/internal/config"
	"log-rotate/internal/model"
	"log-rotate/internal/repository"
)

type LogWriter struct {
	repo         *repository.Repository
	logDir       string
	currentFile  *os.File
	currentPath  string
	mu           sync.RWMutex
	wg           sync.WaitGroup
	rotateCh     chan struct{}
	compressCh   chan string
	ctx          context.Context
	cancel       context.CancelFunc
	configMu     sync.RWMutex
	cfg          *model.LogConfig
	diskCheckMu  sync.Mutex
}

func NewLogWriter(repo *repository.Repository) (*LogWriter, error) {
	cfg, err := repo.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("get config failed: %w", err)
	}

	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return nil, fmt.Errorf("create log dir failed: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	lw := &LogWriter{
		repo:       repo,
		logDir:     cfg.LogDir,
		rotateCh:   make(chan struct{}, 1),
		compressCh: make(chan string, 10),
		ctx:        ctx,
		cancel:     cancel,
		cfg:        cfg,
	}

	if err := lw.openCurrentFile(); err != nil {
		cancel()
		return nil, err
	}

	go lw.run()
	go lw.runCompressor()

	return lw, nil
}

func (lw *LogWriter) getCurrentFilePath() string {
	return filepath.Join(lw.logDir, "app.log")
}

func (lw *LogWriter) openCurrentFile() error {
	lw.mu.Lock()
	defer lw.mu.Unlock()

	path := lw.getCurrentFilePath()
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file failed: %w", err)
	}
	lw.currentFile = file
	lw.currentPath = path
	return nil
}

func (lw *LogWriter) Write(level, message string) error {
	truncated := false
	if len(message) > config.MaxLogEntrySize-len(level)-len(time.RFC3339Nano)-20-len(config.TruncateMarker) {
		message = message[:config.MaxLogEntrySize-len(level)-len(time.RFC3339Nano)-20-len(config.TruncateMarker)] + config.TruncateMarker
		truncated = true
	}

	logLine := fmt.Sprintf("[%s] %s %s", time.Now().Format(time.RFC3339Nano), strings.ToUpper(level), message)
	if truncated {
		logLine += " " + config.TruncateMarker
	}
	logLine += "\n"

	lw.mu.RLock()
	file := lw.currentFile
	lw.mu.RUnlock()

	if _, err := file.WriteString(logLine); err != nil {
		return fmt.Errorf("write log failed: %w", err)
	}

	lw.checkRotateTrigger()
	return nil
}

func (lw *LogWriter) checkRotateTrigger() {
	lw.configMu.RLock()
	cfg := lw.cfg
	lw.configMu.RUnlock()

	lw.mu.RLock()
	file := lw.currentFile
	lw.mu.RUnlock()

	info, err := file.Stat()
	if err != nil {
		return
	}

	maxSize := int64(cfg.MaxFileSizeMB) * 1024 * 1024
	if info.Size() >= maxSize {
		lw.triggerRotate()
		return
	}

	now := time.Now()
	if now.Hour() == cfg.RotateHour && now.Minute() == cfg.RotateMinute {
		lw.triggerRotate()
	}
}

func (lw *LogWriter) triggerRotate() {
	select {
	case lw.rotateCh <- struct{}{}:
	default:
	}
}

func (lw *LogWriter) TriggerRotate() {
	lw.triggerRotate()
}

func (lw *LogWriter) run() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-lw.ctx.Done():
			return
		case <-lw.rotateCh:
			lw.doRotate()
		case <-ticker.C:
			lw.checkDiskSpace()
		}
	}
}

func (lw *LogWriter) doRotate() {
	lw.configMu.RLock()
	cfg := lw.cfg
	lw.configMu.RUnlock()

	oldPath := lw.getCurrentFilePath()

	info, err := os.Stat(oldPath)
	if err != nil || info.Size() == 0 {
		return
	}

	archiveName := fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02"))
	archivePath := filepath.Join(lw.logDir, archiveName)

	lw.mu.Lock()
	if err := lw.currentFile.Sync(); err != nil {
		log.Printf("sync file failed: %v", err)
	}
	if err := lw.currentFile.Close(); err != nil {
		log.Printf("close file failed: %v", err)
	}
	lw.currentFile = nil
	lw.mu.Unlock()

	if err := os.Rename(oldPath, archivePath); err != nil {
		log.Printf("rename log file failed: %v", err)
		lw.openCurrentFile()
		return
	}

	if err := lw.openCurrentFile(); err != nil {
		log.Printf("open new file failed: %v", err)
		return
	}

	if err := lw.repo.AddArchive(archiveName, info.Size(), false); err != nil {
		log.Printf("add archive record failed: %v", err)
	}

	select {
	case lw.compressCh <- archivePath:
	default:
		go lw.compressFile(archivePath)
	}

	lw.cleanupOldArchives(cfg.MaxArchiveFiles)
}

func (lw *LogWriter) runCompressor() {
	for {
		select {
		case <-lw.ctx.Done():
			return
		case path := <-lw.compressCh:
			lw.compressFile(path)
		}
	}
}

func (lw *LogWriter) compressFile(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return
	}

	fileName := filepath.Base(path)
	gzPath := path + ".gz"
	inFile, err := os.Open(path)
	if err != nil {
		log.Printf("open file for compress failed: %v", err)
		return
	}
	defer inFile.Close()

	outFile, err := os.Create(gzPath)
	if err != nil {
		log.Printf("create gz file failed: %v", err)
		return
	}

	gzw := gzip.NewWriter(outFile)
	if _, err := io.Copy(gzw, inFile); err != nil {
		log.Printf("compress failed: %v", err)
		gzw.Close()
		outFile.Close()
		os.Remove(gzPath)
		return
	}

	if err := gzw.Close(); err != nil {
		log.Printf("close gzip writer failed: %v", err)
		outFile.Close()
		os.Remove(gzPath)
		return
	}

	if err := outFile.Close(); err != nil {
		log.Printf("close gz file failed: %v", err)
		os.Remove(gzPath)
		return
	}

	if err := lw.repo.UpdateArchiveCompressed(fileName); err != nil {
		log.Printf("update archive compressed status failed: %v", err)
	}

	if err := os.Remove(path); err != nil {
		log.Printf("remove original file failed: %v", err)
	}
}

func (lw *LogWriter) cleanupOldArchives(maxFiles int) {
	archives, err := lw.repo.GetArchives(maxFiles + 10)
	if err != nil {
		log.Printf("get archives failed: %v", err)
		return
	}

	if len(archives) <= maxFiles {
		return
	}

	for i := maxFiles; i < len(archives); i++ {
		archive := archives[i]
		filePath := filepath.Join(lw.logDir, archive.FileName)
		gzPath := filePath + ".gz"

		if archive.Compressed {
			os.Remove(gzPath)
		} else {
			os.Remove(filePath)
		}

		if err := lw.repo.MarkArchiveDeleted(archive.ID); err != nil {
			log.Printf("mark archive deleted failed: %v", err)
		}
	}
}

func (lw *LogWriter) checkDiskSpace() {
	lw.diskCheckMu.Lock()
	defer lw.diskCheckMu.Unlock()

	archives, err := lw.repo.GetArchives(100)
	if err != nil {
		return
	}

	var totalSize int64
	for _, a := range archives {
		totalSize += a.FileSize
	}

	if len(archives) > 5 {
		for i := len(archives) - 1; i >= len(archives)-3 && i >= 0; i-- {
			archive := archives[i]
			filePath := filepath.Join(lw.logDir, archive.FileName)
			gzPath := filePath + ".gz"

			if archive.Compressed {
				os.Remove(gzPath)
			} else {
				os.Remove(filePath)
			}
			lw.repo.MarkArchiveDeleted(archive.ID)
		}
	}
}

func (lw *LogWriter) UpdateConfig(cfg *model.LogConfig) error {
	lw.configMu.Lock()
	lw.cfg = cfg
	lw.configMu.Unlock()
	return nil
}

func (lw *LogWriter) GetConfig() *model.LogConfig {
	lw.configMu.RLock()
	defer lw.configMu.RUnlock()
	return lw.cfg
}

func (lw *LogWriter) GetStatus() map[string]interface{} {
	lw.mu.RLock()
	var fileSize int64
	if lw.currentFile != nil {
		if info, err := lw.currentFile.Stat(); err == nil {
			fileSize = info.Size()
		}
	}
	lw.mu.RUnlock()

	return map[string]interface{}{
		"current_file_size": fileSize,
		"current_file_path": lw.currentPath,
		"config":            lw.GetConfig(),
	}
}

func (lw *LogWriter) Close() error {
	lw.cancel()

	lw.mu.Lock()
	if lw.currentFile != nil {
		lw.currentFile.Sync()
		lw.currentFile.Close()
	}
	lw.mu.Unlock()

	return nil
}
