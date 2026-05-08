package logger

import (
	"bytes"
	"encoding/json"
)

const truncationMarker = "[TRUNCATED]"
const markerLen = len(truncationMarker)

func truncateField(value string, maxLen int) string {
	if maxLen <= 0 || len(value) <= maxLen {
		return value
	}
	if maxLen <= markerLen {
		return truncationMarker[:maxLen]
	}
	targetLen := maxLen - markerLen
	if targetLen <= 0 {
		return truncationMarker
	}
	runes := []rune(value)
	if len(runes) <= targetLen {
		return value
	}
	truncated := string(runes[:targetLen])
	return truncated + truncationMarker
}

func truncateFieldValue(value any, maxLen int) any {
	if maxLen <= 0 {
		return value
	}
	switch v := value.(type) {
	case string:
		return truncateField(v, maxLen)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return "[UNMARSHALABLE: " + err.Error() + "]"
		}
		if len(data) <= maxLen {
			return v
		}
		return truncateField(string(data), maxLen)
	}
}

func sanitizeValue(value any) any {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(value); err != nil {
		return "[UNSERIALIZABLE: " + err.Error() + "]"
	}
	return value
}
