package relay

import (
	"encoding/json"
	"strings"
)

type AlertLevel string

const (
	AlertLevelInfo  AlertLevel = "info"
	AlertLevelWarn  AlertLevel = "warn"
	AlertLevelError AlertLevel = "error"
	AlertLevelFatal AlertLevel = "fatal"
)

type Alert struct {
	Level   AlertLevel `json:"level"`
	Message string     `json:"message"`
}

func ParseAlert(raw []byte) (*Alert, error) {
	var rawAlert map[string]interface{}
	if err := json.Unmarshal(raw, &rawAlert); err != nil {
		return nil, err
	}

	alert := &Alert{
		Level: AlertLevelInfo,
	}

	if levelVal, ok := rawAlert["level"]; ok {
		if levelStr, ok := levelVal.(string); ok {
			alert.Level = NormalizeLevel(levelStr)
		}
	}

	if messageVal, ok := rawAlert["message"]; ok {
		if messageStr, ok := messageVal.(string); ok {
			alert.Message = messageStr
		}
	}

	return alert, nil
}

func NormalizeLevel(level string) AlertLevel {
	lower := strings.ToLower(level)
	switch lower {
	case "warn", "warning":
		return AlertLevelWarn
	case "error":
		return AlertLevelError
	case "fatal", "critical":
		return AlertLevelFatal
	case "info", "information", "notice":
		return AlertLevelInfo
	default:
		return AlertLevelInfo
	}
}

func GetChannelsForLevel(level AlertLevel) []string {
	switch level {
	case AlertLevelInfo:
		return []string{"email"}
	case AlertLevelWarn:
		return []string{"email", "dingtalk"}
	case AlertLevelError, AlertLevelFatal:
		return []string{"email", "sms", "dingtalk"}
	default:
		return []string{"email"}
	}
}

func GetLevelEmoji(level AlertLevel) string {
	switch level {
	case AlertLevelInfo:
		return "ℹ️"
	case AlertLevelWarn:
		return "⚠️"
	case AlertLevelError:
		return "❌"
	case AlertLevelFatal:
		return "🔥"
	default:
		return "ℹ️"
	}
}
