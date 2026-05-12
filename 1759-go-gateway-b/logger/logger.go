package logger

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"gateway/config"
)

type LogEntry struct {
	Time      string `json:"time"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	RemoteAddr string `json:"remote_addr"`
	UserAgent string `json:"user_agent"`
	Status    int    `json:"status"`
	Duration  string `json:"duration"`
}

type Logger struct {
	cfg *config.Manager
	mu  sync.Mutex
}

func NewLogger(cfg *config.Manager) *Logger {
	return &Logger{cfg: cfg}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)

		cfg := l.cfg.Get()
		if !cfg.Log.Enabled {
			return
		}

		entry := LogEntry{
			Time:       start.Format(time.RFC3339),
			Method:     r.Method,
			Path:       r.URL.Path,
			RemoteAddr: r.RemoteAddr,
			UserAgent:  r.UserAgent(),
			Status:     rec.status,
			Duration:   duration.String(),
		}

		l.log(cfg.Log.File, entry)
	})
}

func (l *Logger) log(file string, entry LogEntry) {
	line := fmt.Sprintf("%s %s %s %s %d %s\n",
		entry.Time,
		entry.Method,
		entry.Path,
		entry.RemoteAddr,
		entry.Status,
		entry.Duration,
	)

	l.mu.Lock()
	defer l.mu.Unlock()

	if file != "" {
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			f.WriteString(line)
			return
		}
	}

	fmt.Print(line)
}
