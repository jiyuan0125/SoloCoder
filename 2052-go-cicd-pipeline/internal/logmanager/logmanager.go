package logmanager

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LogManager struct {
	logDir string
	mu     sync.Mutex
}

func New(logDir string) (*LogManager, error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	return &LogManager{logDir: logDir}, nil
}

func (lm *LogManager) CreateLog(executionID, phaseResultID, taskResultID string) (string, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	dir := filepath.Join(lm.logDir, executionID, phaseResultID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	logPath := filepath.Join(dir, fmt.Sprintf("%s.log", taskResultID))
	f, err := os.Create(logPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return logPath, nil
}

func (lm *LogManager) AppendLog(logPath string, content string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().UTC().Format(time.RFC3339), content))
	return err
}

func (lm *LogManager) AppendLogRaw(logPath string, content string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}

func (lm *LogManager) ReadLog(logPath string) (string, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	data, err := os.ReadFile(logPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (lm *LogManager) StreamLog(logPath string) (io.ReadCloser, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	f, err := os.Open(logPath)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (lm *LogManager) StreamLogLines(logPath string) (*bufio.Scanner, io.Closer, error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	f, err := os.Open(logPath)
	if err != nil {
		return nil, nil, err
	}
	scanner := bufio.NewScanner(f)
	return scanner, f, nil
}
