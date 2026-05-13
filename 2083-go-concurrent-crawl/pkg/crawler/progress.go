package crawler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type URLState string

const (
	StatePending   URLState = "pending"
	StateProcessing URLState = "processing"
	StateCompleted URLState = "completed"
	StateFailed    URLState = "failed"
)

type URLInfo struct {
	URL    string   `json:"url"`
	Depth  int      `json:"depth"`
	State  URLState `json:"state"`
	Error  string   `json:"error,omitempty"`
}

type Progress struct {
	Visited   map[string]URLInfo `json:"visited"`
	Completed int                `json:"completed"`
	Failed    int                `json:"failed"`
	QueueSize int                `json:"queue_size"`
}

type ProgressManager struct {
	path     string
	progress *Progress
	mu       sync.RWMutex
}

func NewProgressManager(outputDir string) *ProgressManager {
	return &ProgressManager{
		path: filepath.Join(outputDir, "progress.json"),
		progress: &Progress{
			Visited: make(map[string]URLInfo),
		},
	}
}

func (pm *ProgressManager) Load() bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	data, err := os.ReadFile(pm.path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		fmt.Fprintf(os.Stderr, "Warning: Failed to read progress file, starting fresh: %v\n", err)
		return false
	}

	var p Progress
	if err := json.Unmarshal(data, &p); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Progress file is corrupted, starting fresh: %v\n", err)
		return false
	}

	if p.Visited == nil {
		p.Visited = make(map[string]URLInfo)
	}

	pm.progress = &p
	return true
}

func (pm *ProgressManager) Save() error {
	pm.mu.RLock()
	data, err := json.MarshalIndent(pm.progress, "", "  ")
	pm.mu.RUnlock()
	
	if err != nil {
		return err
	}

	dir := filepath.Dir(pm.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(pm.path, data, 0644)
}

func (pm *ProgressManager) IsVisited(url string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	_, exists := pm.progress.Visited[url]
	return exists
}

func (pm *ProgressManager) AddURL(url string, depth int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if _, exists := pm.progress.Visited[url]; !exists {
		pm.progress.Visited[url] = URLInfo{
			URL:   url,
			Depth: depth,
			State: StatePending,
		}
		pm.progress.QueueSize++
	}
}

func (pm *ProgressManager) MarkProcessing(url string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if info, exists := pm.progress.Visited[url]; exists {
		info.State = StateProcessing
		pm.progress.Visited[url] = info
	}
}

func (pm *ProgressManager) MarkCompleted(url string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if info, exists := pm.progress.Visited[url]; exists {
		info.State = StateCompleted
		pm.progress.Visited[url] = info
		pm.progress.Completed++
		if pm.progress.QueueSize > 0 {
			pm.progress.QueueSize--
		}
	}
}

func (pm *ProgressManager) MarkFailed(url string, err error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	
	if info, exists := pm.progress.Visited[url]; exists {
		info.State = StateFailed
		if err != nil {
			info.Error = err.Error()
		}
		pm.progress.Visited[url] = info
		pm.progress.Failed++
		if pm.progress.QueueSize > 0 {
			pm.progress.QueueSize--
		}
	}
}

func (pm *ProgressManager) GetPendingURLs() []URLInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	
	var pending []URLInfo
	for _, info := range pm.progress.Visited {
		if info.State == StatePending || info.State == StateProcessing {
			pending = append(pending, info)
		}
	}
	return pending
}

func (pm *ProgressManager) GetStats() (completed, failed, queueSize int) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.progress.Completed, pm.progress.Failed, pm.progress.QueueSize
}

func (pm *ProgressManager) GetProgress() *Progress {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.progress
}
