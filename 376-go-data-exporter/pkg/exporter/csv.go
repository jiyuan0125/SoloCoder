package exporter

import (
	"io"
	"strings"
)

func needsCSVQuoting(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			return true
		}
	}
	return false
}

func escapeCSVValue(s string) string {
	if !needsCSVQuoting(s) {
		return s
	}
	var builder strings.Builder
	builder.WriteByte('"')
	for _, c := range s {
		if c == '"' {
			builder.WriteString(`""`)
		} else {
			builder.WriteRune(c)
		}
	}
	builder.WriteByte('"')
	return builder.String()
}

func writeCSVRow(w io.Writer, values []string) error {
	for i, val := range values {
		if i > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}
		escaped := escapeCSVValue(val)
		if _, err := w.Write([]byte(escaped)); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte{'\n'}); err != nil {
		return err
	}
	return nil
}
