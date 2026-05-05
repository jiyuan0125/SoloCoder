package relay

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type FileWatcher struct {
	filePath     string
	offsetStore  *OffsetStore
	forwarder    *Forwarder
	alertsParsed int64
	mu           sync.RWMutex
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

func NewFileWatcher(filePath string, offsetStore *OffsetStore, forwarder *Forwarder) *FileWatcher {
	return &FileWatcher{
		filePath:    filePath,
		offsetStore: offsetStore,
		forwarder:   forwarder,
		stopChan:    make(chan struct{}),
	}
}

func (w *FileWatcher) Start() {
	w.wg.Add(1)
	go w.watchLoop()
}

func (w *FileWatcher) Stop() {
	close(w.stopChan)
	w.wg.Wait()
}

func (w *FileWatcher) GetFilePath() string {
	return w.filePath
}

func (w *FileWatcher) GetCurrentOffset() int64 {
	return w.offsetStore.GetOffset(w.filePath)
}

func (w *FileWatcher) GetAlertsParsed() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.alertsParsed
}

func (w *FileWatcher) watchLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.checkFile()
		}
	}
}

func (w *FileWatcher) checkFile() {
	offset := w.offsetStore.GetOffset(w.filePath)

	file, err := os.Open(w.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("无法打开文件 %s: %v", w.filePath, err)
		}
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Printf("无法获取文件信息 %s: %v", w.filePath, err)
		return
	}

	fileSize := fileInfo.Size()
	if fileSize <= offset {
		return
	}

	_, err = file.Seek(offset, 0)
	if err != nil {
		log.Printf("无法定位到偏移量 %d: %v", offset, err)
		return
	}

	scanner := bufio.NewScanner(file)
	bytesRead := int64(0)

	for scanner.Scan() {
		line := scanner.Text()
		lineBytes := len(line) + 1
		bytesRead += int64(lineBytes)

		if strings.TrimSpace(line) == "" {
			continue
		}

		alert, err := ParseAlert([]byte(line))
		if err != nil {
			log.Printf("警告: 跳过无效的JSON行: %v", err)
			continue
		}

		w.mu.Lock()
		w.alertsParsed++
		w.mu.Unlock()

		w.forwarder.Forward(alert)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("扫描文件错误: %v", err)
	}

	newOffset := offset + bytesRead
	if err := w.offsetStore.SetOffset(w.filePath, newOffset); err != nil {
		log.Printf("无法保存偏移量: %v", err)
	}
}
