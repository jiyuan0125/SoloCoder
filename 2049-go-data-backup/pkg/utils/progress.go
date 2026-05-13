package utils

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type ProgressBar struct {
	total      int64
	written    int64
	lastUpdate time.Time
	label      string
	mu         sync.Mutex
}

func NewProgressBar(label string, total int64) *ProgressBar {
	return &ProgressBar{
		label:      label,
		total:      total,
		lastUpdate: time.Now(),
	}
}

func (pb *ProgressBar) Update(written int64, total int64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.written = written
	pb.total = total

	if time.Since(pb.lastUpdate) < 100*time.Millisecond && written != total {
		return
	}

	pb.lastUpdate = time.Now()
	pb.render()
}

func (pb *ProgressBar) render() {
	percent := float64(pb.written) / float64(pb.total) * 100
	if pb.total == 0 {
		percent = 100
	}

	barWidth := 30
	filled := int(float64(barWidth) * percent / 100)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	writtenMB := float64(pb.written) / 1024 / 1024
	totalMB := float64(pb.total) / 1024 / 1024

	fmt.Printf("\r%s [%s] %.1f%% (%.2f MB / %.2f MB)",
		pb.label, bar, percent, writtenMB, totalMB)

	if pb.written >= pb.total {
		fmt.Println()
	}
}

func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB",
		float64(bytes)/float64(div), "KMGTPE"[exp])
}
