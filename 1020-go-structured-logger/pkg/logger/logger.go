package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"structured-logger/pkg/api"
)

const (
	DefaultMaxFileSize = 100 * 1024 * 1024
	DefaultMaxBackups  = 5
)

type Logger struct {
	mu          sync.RWMutex
	level       int32
	levelStr    LogLevel
	writers     []io.Writer
	fieldLimits map[string]int
	recentLogs  *ringBuffer
	options     LoggerOptions
}

type LoggerOptions struct {
	Level        LogLevel
	OutputToStd  bool
	OutputToFile bool
	FilePath     string
	MaxFileSize  int64
	MaxBackups   int
	FieldLimits  map[string]int
	RecentBuffer int
}

type ringBuffer struct {
	mu    sync.Mutex
	logs  []api.LogEntry
	head  int
	tail  int
	count int
	size  int
}

func newRingBuffer(size int) *ringBuffer {
	if size <= 0 {
		size = 1000
	}
	return &ringBuffer{
		logs:  make([]api.LogEntry, size),
		size:  size,
	}
}

func (rb *ringBuffer) Add(entry api.LogEntry) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.logs[rb.tail] = entry
	rb.tail = (rb.tail + 1) % rb.size
	if rb.count < rb.size {
		rb.count++
	} else {
		rb.head = (rb.head + 1) % rb.size
	}
}

func (rb *ringBuffer) GetLast(n int) []api.LogEntry {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	if n <= 0 {
		n = 10
	}
	if n > rb.count {
		n = rb.count
	}
	result := make([]api.LogEntry, 0, n)
	for i := 0; i < n; i++ {
		idx := (rb.tail - 1 - i + rb.size) % rb.size
		result = append(result, rb.logs[idx])
	}
	return result
}

func New(options ...LoggerOptions) (*Logger, error) {
	opts := LoggerOptions{
		Level:        LevelInfo,
		OutputToStd:  true,
		OutputToFile: false,
		MaxFileSize:  DefaultMaxFileSize,
		MaxBackups:   DefaultMaxBackups,
		RecentBuffer: 1000,
	}
	if len(options) > 0 {
		opts = options[0]
	}
	l := &Logger{
		levelStr:    opts.Level,
		fieldLimits: opts.FieldLimits,
		recentLogs:  newRingBuffer(opts.RecentBuffer),
		options:     opts,
	}
	atomic.StoreInt32(&l.level, int32(levelOrder[opts.Level]))
	var writers []io.Writer
	if opts.OutputToStd {
		writers = append(writers, os.Stdout)
	}
	if opts.OutputToFile && opts.FilePath != "" {
		fileWriter, err := NewRotatingFileWriter(opts.FilePath, opts.MaxFileSize, opts.MaxBackups)
		if err != nil {
			return nil, err
		}
		writers = append(writers, fileWriter)
	}
	l.writers = writers
	return l, nil
}

func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.levelStr = level
	atomic.StoreInt32(&l.level, int32(levelOrder[level]))
}

func (l *Logger) GetLevel() LogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.levelStr
}

func (l *Logger) GetConfig() api.GetConfigResponse {
	l.mu.RLock()
	defer l.mu.RUnlock()
	fieldLimits := make(map[string]int)
	for k, v := range l.fieldLimits {
		fieldLimits[k] = v
	}
	return api.GetConfigResponse{
		Level:        l.levelStr,
		MaxFileSize:  l.options.MaxFileSize,
		MaxBackups:   l.options.MaxBackups,
		FieldLimits:  fieldLimits,
		OutputToFile: l.options.OutputToFile,
		OutputToStd:  l.options.OutputToStd,
	}
}

func (l *Logger) GetRecentLogs(n int) []api.LogEntry {
	return l.recentLogs.GetLast(n)
}

func (l *Logger) shouldLog(level LogLevel) bool {
	return ShouldLog(level, l.levelStr)
}

func (l *Logger) Debug(msg string, fields ...map[string]any) {
	l.log(LevelDebug, msg, 2, fields...)
}

func (l *Logger) Info(msg string, fields ...map[string]any) {
	l.log(LevelInfo, msg, 2, fields...)
}

func (l *Logger) Warn(msg string, fields ...map[string]any) {
	l.log(LevelWarn, msg, 2, fields...)
}

func (l *Logger) Error(msg string, fields ...map[string]any) {
	l.log(LevelError, msg, 2, fields...)
}

func (l *Logger) Log(level LogLevel, msg string, fields ...map[string]any) {
	l.log(level, msg, 2, fields...)
}

func (l *Logger) log(level LogLevel, msg string, skip int, fields ...map[string]any) {
	if !l.shouldLog(level) {
		return
	}
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		file = "???"
		line = 0
	}
	if i := lastIndexOfByte(file, '/'); i >= 0 {
		file = file[i+1:]
	}
	var mergedFields map[string]any
	if len(fields) > 0 && len(fields[0]) > 0 {
		mergedFields = make(map[string]any)
		for k, v := range fields[0] {
			mergedFields[k] = v
		}
	}
	entry := api.LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		File:      file,
		Line:      line,
		Fields:    mergedFields,
	}
	l.recentLogs.Add(entry)
	outputEntry := l.prepareEntry(entry)
	data, err := json.Marshal(outputEntry)
	if err != nil {
		return
	}
	var buf bytes.Buffer
	buf.Write(data)
	buf.WriteByte('\n')
	output := buf.Bytes()
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, w := range l.writers {
		w.Write(output)
	}
}

func (l *Logger) prepareEntry(entry api.LogEntry) api.LogEntry {
	result := entry
	if limit, ok := l.fieldLimits["message"]; ok && limit > 0 {
		result.Message = truncateField(entry.Message, limit)
	}
	if entry.Fields != nil {
		result.Fields = make(map[string]any)
		for k, v := range entry.Fields {
			safeV := sanitizeValue(v)
			if limit, ok := l.fieldLimits[k]; ok && limit > 0 {
				result.Fields[k] = truncateFieldValue(safeV, limit)
			} else {
				result.Fields[k] = safeV
			}
		}
	}
	return result
}

func lastIndexOfByte(s string, b byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func truncateJSONNumber(num string, maxLen int) string {
	if len(num) <= maxLen {
		return num
	}
	if maxLen <= 3 {
		return num[:maxLen]
	}
	return num[:maxLen-3] + "..."
}

func validateAndMarshalValue(value any) (json.RawMessage, error) {
	switch v := value.(type) {
	case string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool, nil:
		return json.Marshal(v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return json.Marshal(fmt.Sprintf("[UNSERIALIZABLE: %s]", err.Error()))
		}
		return data, nil
	}
}

func FormatInt(n int64) string {
	return strconv.FormatInt(n, 10)
}

func FormatUint(n uint64) string {
	return strconv.FormatUint(n, 10)
}

func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}
