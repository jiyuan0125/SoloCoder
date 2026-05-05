package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"go-report-gen/protocol"
)

const (
	historyDir = ".report_history"
	maxHistory = 100
)

type HistoryManager struct {
	reports    map[string]*protocol.Report
	mu         sync.RWMutex
	storageDir string
}

func NewHistoryManager() (*HistoryManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	storageDir := filepath.Join(homeDir, historyDir)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, fmt.Errorf("创建历史记录目录失败: %w", err)
	}

	hm := &HistoryManager{
		reports:    make(map[string]*protocol.Report),
		storageDir: storageDir,
	}

	if err := hm.loadFromDisk(); err != nil {
		fmt.Printf("警告: 加载历史记录失败: %v\n", err)
	}

	return hm, nil
}

func (hm *HistoryManager) AddReport(report *protocol.Report) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.reports[report.ID] = report

	if err := hm.saveToDisk(report); err != nil {
		return err
	}

	hm.cleanupOldReports()

	return nil
}

func (hm *HistoryManager) GetReport(reportID string) (*protocol.Report, bool) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	report, exists := hm.reports[reportID]
	return report, exists
}

func (hm *HistoryManager) ListHistory(limit int) []protocol.HistoryEntry {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	var entries []protocol.HistoryEntry
	for _, report := range hm.reports {
		entries = append(entries, protocol.HistoryEntry{
			ID:           report.ID,
			RepoPath:     report.RepoPath,
			GeneratedAt:  report.GeneratedAt,
			CommitCount:  report.Stats.TotalCommits,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].GeneratedAt.After(entries[j].GeneratedAt)
	})

	if limit > 0 && limit < len(entries) {
		entries = entries[:limit]
	}

	return entries
}

func (hm *HistoryManager) saveToDisk(report *protocol.Report) error {
	filePath := filepath.Join(hm.storageDir, fmt.Sprintf("%s.json", report.ID))

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化报告失败: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("保存报告到文件失败: %w", err)
	}

	return nil
}

func (hm *HistoryManager) loadFromDisk() error {
	files, err := filepath.Glob(filepath.Join(hm.storageDir, "*.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var report protocol.Report
		if err := json.Unmarshal(data, &report); err != nil {
			continue
		}

		hm.reports[report.ID] = &report
	}

	return nil
}

func (hm *HistoryManager) cleanupOldReports() {
	if len(hm.reports) <= maxHistory {
		return
	}

	var entries []protocol.HistoryEntry
	for _, report := range hm.reports {
		entries = append(entries, protocol.HistoryEntry{
			ID:          report.ID,
			GeneratedAt: report.GeneratedAt,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].GeneratedAt.After(entries[j].GeneratedAt)
	})

	for i := maxHistory; i < len(entries); i++ {
		reportID := entries[i].ID
		delete(hm.reports, reportID)

		filePath := filepath.Join(hm.storageDir, fmt.Sprintf("%s.json", reportID))
		os.Remove(filePath)
	}
}

func SaveReportContentToFile(content string, baseDir string) (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("report_%s.txt", timestamp)

	var outputDir string
	if baseDir != "" {
		outputDir = baseDir
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "."
		}
		outputDir = filepath.Join(homeDir, historyDir)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("创建输出目录失败: %w", err)
	}

	filePath := filepath.Join(outputDir, fileName)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("保存报告到文件失败: %w", err)
	}

	return filePath, nil
}
