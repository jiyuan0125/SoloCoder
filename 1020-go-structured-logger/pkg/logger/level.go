package logger

import "structured-logger/pkg/api"

type LogLevel = api.LogLevel

const (
	LevelDebug = api.LevelDebug
	LevelInfo  = api.LevelInfo
	LevelWarn  = api.LevelWarn
	LevelError = api.LevelError
)

var levelOrder = map[LogLevel]int{
	LevelDebug: 0,
	LevelInfo:  1,
	LevelWarn:  2,
	LevelError: 3,
}

func ParseLevel(level string) LogLevel {
	switch level {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}

func ShouldLog(level, minLevel LogLevel) bool {
	return levelOrder[level] >= levelOrder[minLevel]
}
