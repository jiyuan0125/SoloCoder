package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

func ParseLogLevel(s string) (LogLevel, error) {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARN":
		return LevelWarn, nil
	case "ERROR":
		return LevelError, nil
	default:
		return LevelInfo, errors.New("invalid log level: " + s)
	}
}

func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "INFO"
	}
}

type Logger struct {
	mu           sync.RWMutex
	level        LogLevel
	filePath     string
	file         *os.File
	fileWriter   *bufio.Writer
	bufferSize   int
	buffer       []map[string]interface{}
	flushTimeout time.Duration
	totalWritten int64
	lastFlush    time.Time
	ticker       *time.Ticker
	done         chan struct{}
}

func NewLogger(filePath string, bufferSize int, flushTimeout time.Duration, level LogLevel) (*Logger, error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	l := &Logger{
		level:        level,
		filePath:     filePath,
		file:         f,
		fileWriter:   bufio.NewWriter(f),
		bufferSize:   bufferSize,
		buffer:       make([]map[string]interface{}, 0, bufferSize),
		flushTimeout: flushTimeout,
		lastFlush:    time.Now(),
		done:         make(chan struct{}),
	}

	l.ticker = time.NewTicker(flushTimeout / 2)
	go l.flushLoop()

	return l, nil
}

func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

func (l *Logger) GetStats() (bufferCount int, totalWritten int64) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.buffer), l.totalWritten
}

func (l *Logger) ShouldLog(level LogLevel) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return level >= l.level
}

func (l *Logger) Log(level LogLevel, source string, traceID string, message string, customFields map[string]interface{}) {
	if !l.ShouldLog(level) {
		return
	}

	entry := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"level":     level.String(),
		"source":    source,
		"message":   message,
	}

	if traceID != "" {
		entry["trace_id"] = traceID
	}

	for k, v := range customFields {
		if _, reserved := entry[k]; !reserved {
			entry[k] = v
		}
	}

	l.mu.Lock()
	l.buffer = append(l.buffer, entry)
	shouldFlush := len(l.buffer) >= l.bufferSize
	l.mu.Unlock()

	if shouldFlush {
		l.Flush()
	}
}

func (l *Logger) flushLoop() {
	for {
		select {
		case <-l.ticker.C:
			l.mu.RLock()
			needFlush := len(l.buffer) > 0 && time.Since(l.lastFlush) >= l.flushTimeout
			l.mu.RUnlock()
			if needFlush {
				l.Flush()
			}
		case <-l.done:
			return
		}
	}
}

func (l *Logger) Flush() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.buffer) == 0 {
		return
	}

	data, err := json.Marshal(l.buffer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal log buffer: %v\n", err)
		l.buffer = l.buffer[:0]
		return
	}

	if _, err := l.fileWriter.Write(append(data, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write log buffer: %v\n", err)
		l.buffer = l.buffer[:0]
		return
	}

	if err := l.fileWriter.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to flush log file: %v\n", err)
	}

	written := len(l.buffer)
	l.totalWritten += int64(written)
	l.buffer = l.buffer[:0]
	l.lastFlush = time.Now()
}

func (l *Logger) Close() {
	select {
	case <-l.done:
		return
	default:
		close(l.done)
	}
	if l.ticker != nil {
		l.ticker.Stop()
	}
	l.Flush()
	if l.fileWriter != nil {
		l.fileWriter.Flush()
	}
	if l.file != nil {
		l.file.Sync()
		l.file.Close()
	}
}

type Config struct {
	Port         int
	LogFilePath  string
	BufferSize   int
	FlushTimeout time.Duration
	LogLevel     LogLevel
}

func getEnvInt(key string, defaultValue int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}

func LoadConfig() *Config {
	level, _ := ParseLogLevel(os.Getenv("LOG_LEVEL"))
	if level == LevelInfo && os.Getenv("LOG_LEVEL") == "" {
		level = LevelInfo
	}

	return &Config{
		Port:         getEnvInt("PORT", 8080),
		LogFilePath:  os.Getenv("LOG_FILE"),
		BufferSize:   getEnvInt("BUFFER_SIZE", 500),
		FlushTimeout: getEnvDuration("FLUSH_TIMEOUT", 5*time.Second),
		LogLevel:     level,
	}
}

type LogRequest struct {
	Level   string                 `json:"level"`
	Source  string                 `json:"source"`
	TraceID string                 `json:"trace_id,omitempty"`
	Message string                 `json:"message"`
	Custom  map[string]interface{} `json:"-"`
}

func (r *LogRequest) UnmarshalJSON(data []byte) error {
	raw := map[string]interface{}{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if l, ok := raw["level"].(string); ok {
		r.Level = l
	}
	if s, ok := raw["source"].(string); ok {
		r.Source = s
	}
	if t, ok := raw["trace_id"].(string); ok {
		r.TraceID = t
	}
	if m, ok := raw["message"].(string); ok {
		r.Message = m
	}

	custom := map[string]interface{}{}
	for k, v := range raw {
		if k != "level" && k != "source" && k != "trace_id" && k != "message" {
			custom[k] = v
		}
	}
	r.Custom = custom

	return nil
}

type Server struct {
	logger *Logger
	cfg    *Config
}

func NewServer(logger *Logger, cfg *Config) *Server {
	return &Server{logger: logger, cfg: cfg}
}

func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}

func (s *Server) handlePostLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req LogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if req.Level == "" || req.Source == "" || req.Message == "" {
		s.writeError(w, http.StatusBadRequest, "missing required fields: level, source, message")
		return
	}

	level, err := ParseLogLevel(req.Level)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid log level: "+req.Level)
		return
	}

	s.logger.Log(level, req.Source, req.TraceID, req.Message, req.Custom)
	s.writeJSON(w, http.StatusAccepted, map[string]string{"status": "accepted"})
}

func (s *Server) handleGetLevel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"level": s.logger.GetLevel().String()})
}

func (s *Server) handlePutLevel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	levelStr, ok := body["level"]
	if !ok {
		s.writeError(w, http.StatusBadRequest, "missing 'level' field")
		return
	}

	level, err := ParseLogLevel(levelStr)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid log level: "+levelStr)
		return
	}

	s.logger.SetLevel(level)
	s.writeJSON(w, http.StatusOK, map[string]string{"level": level.String()})
}

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	bufferCount, totalWritten := s.logger.GetStats()
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"buffer_size":       s.cfg.BufferSize,
		"buffer_count":      bufferCount,
		"total_flushed":     totalWritten,
		"flush_timeout_sec": s.cfg.FlushTimeout.Seconds(),
	})
}

func (s *Server) handleLevel(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetLevel(w, r)
	case http.MethodPut:
		s.handlePutLevel(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func main() {
	cfg := LoadConfig()

	if cfg.LogFilePath == "" {
		cfg.LogFilePath = "app.log"
	}

	logger, err := NewLogger(cfg.LogFilePath, cfg.BufferSize, cfg.FlushTimeout, cfg.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	server := NewServer(logger, cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/logs", server.handlePostLogs)
	mux.HandleFunc("/logs/level", server.handleLevel)
	mux.HandleFunc("/logs/stats", server.handleGetStats)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	serverDone := make(chan struct{})
	go func() {
		fmt.Printf("server listening on port %d\n", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
		close(serverDone)
	}()

	select {
	case sig := <-sigChan:
		fmt.Printf("received signal %s, shutting down gracefully...\n", sig)
		httpServer.Shutdown(nil)
		<-serverDone
	case <-serverDone:
	}
}
