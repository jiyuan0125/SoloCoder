package logger

import (
	"sync"
	"time"
)

const MaxLogs = 50000

type AccessLog struct {
	ID           int
	Time         time.Time
	ClientIP     string
	RequestPath  string
	BackendHost  string
	BackendPort  int
	StatusCode   int
	Duration     time.Duration
}

type AccessLogger struct {
	logs     []*AccessLog
	counter  int
	mu       sync.RWMutex
}

func NewAccessLogger() *AccessLogger {
	return &AccessLogger{
		logs:    make([]*AccessLog, 0, MaxLogs),
		counter: 0,
	}
}

func (al *AccessLogger) Record(clientIP, requestPath string, backendHost string, backendPort int, statusCode int, duration time.Duration) {
	al.mu.Lock()
	defer al.mu.Unlock()

	al.counter++

	logEntry := &AccessLog{
		ID:           al.counter,
		Time:         time.Now(),
		ClientIP:     clientIP,
		RequestPath:  requestPath,
		BackendHost:  backendHost,
		BackendPort:  backendPort,
		StatusCode:   statusCode,
		Duration:     duration,
	}

	if len(al.logs) >= MaxLogs {
		al.logs = al.logs[1:]
	}

	al.logs = append(al.logs, logEntry)
}

func (al *AccessLogger) List() []*AccessLog {
	al.mu.RLock()
	defer al.mu.RUnlock()

	logs := make([]*AccessLog, len(al.logs))
	copy(logs, al.logs)
	return logs
}

func (al *AccessLogger) Recent(count int) []*AccessLog {
	al.mu.RLock()
	defer al.mu.RUnlock()

	if count <= 0 {
		return []*AccessLog{}
	}

	start := len(al.logs) - count
	if start < 0 {
		start = 0
	}

	logs := make([]*AccessLog, len(al.logs)-start)
	copy(logs, al.logs[start:])
	return logs
}
