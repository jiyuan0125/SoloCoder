package csvparser

import (
	"strings"
)

func Serialize(data [][]string) string {
	if len(data) == 0 {
		return ""
	}

	var sb strings.Builder

	for i, record := range data {
		for j, field := range record {
			if j > 0 {
				sb.WriteByte(',')
			}
			sb.WriteString(serializeField(field))
		}
		if i < len(data)-1 {
			sb.WriteString("\r\n")
		}
	}

	return sb.String()
}

func serializeField(field string) string {
	needsQuotes := strings.ContainsAny(field, ",\"\r\n")

	if !needsQuotes {
		return field
	}

	escaped := strings.ReplaceAll(field, "\"", "\"\"")
	return "\"" + escaped + "\""
}
