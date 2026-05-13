package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const MaxFileSize = 100 * 1024 * 1024

type ExecutionLog struct {
	TaskName    string
	StartTime   time.Time
	EndTime     time.Time
	ExitCode    int
	Output      string
	IsManual    bool
	RetryCount  int
	Error       string
}

type Logger struct {
	logDir string
	mu     sync.Mutex
}

func New(logDir string) (*Logger, error) {
	if logDir == "" {
		logDir = "logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}
	return &Logger{logDir: logDir}, nil
}

func (l *Logger) LogFilePath(taskName string) string {
	return filepath.Join(l.logDir, fmt.Sprintf("%s.log", taskName))
}

func (l *Logger) getCurrentLogFile(taskName string) (*os.File, error) {
	path := l.LogFilePath(taskName)

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		}
		return nil, err
	}

	if info.Size() >= MaxFileSize {
		rotatedPath := path + "." + time.Now().Format("20060102-150405")
		if err := os.Rename(path, rotatedPath); err != nil {
			return nil, err
		}
	}

	return os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
}

func (l *Logger) Write(log *ExecutionLog) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := l.getCurrentLogFile(log.TaskName)
	if err != nil {
		return err
	}
	defer f.Close()

	output := bytes.TrimSpace([]byte(log.Output))
	errorMsg := ""
	if log.Error != "" {
		errorMsg = log.Error
	}

	_, err = fmt.Fprintf(f,
		"[%s] [%s] START=%s END=%s DURATION=%v EXIT=%d RETRY=%d OUTPUT=%q ERROR=%q\n",
		time.Now().Format(time.RFC3339),
		log.TaskName,
		log.StartTime.Format(time.RFC3339),
		log.EndTime.Format(time.RFC3339),
		log.EndTime.Sub(log.StartTime),
		log.ExitCode,
		log.RetryCount,
		string(output),
		errorMsg,
	)
	return err
}

func (l *Logger) GetLogs(taskName string, limit int) ([]string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	path := l.LogFilePath(taskName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	lines := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
	start := 0
	if len(lines) > limit {
		start = len(lines) - limit
	}

	result := make([]string, 0, limit)
	for i := start; i < len(lines); i++ {
		if len(lines[i]) > 0 {
			result = append(result, string(lines[i]))
		}
	}
	return result, nil
}

func (l *Logger) GetLogReader(taskName string) (io.ReadCloser, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return os.Open(l.LogFilePath(taskName))
}
