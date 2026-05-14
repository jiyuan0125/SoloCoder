package main

import (
	"testing"
	"time"
)

func TestCalculateStatus(t *testing.T) {
	tests := []struct {
		name      string
		hour      int
		minute    int
		punchType string
		expected  string
	}{
		{"上班9:00", 9, 0, "in", "normal"},
		{"上班9:05", 9, 5, "in", "late_minor"},
		{"上班9:16", 9, 16, "in", "late"},
		{"上班10:01", 10, 1, "in", "absent_half"},
		{"下班17:05", 17, 5, "out", "early_minor"},
		{"下班17:06", 17, 6, "out", "early_minor"},
		{"下班17:29", 17, 29, "out", "early_minor"},
		{"下班17:30", 17, 30, "out", "normal"},
		{"下班17:35", 17, 35, "out", "normal"},
		{"下班18:00", 18, 0, "out", "normal"},
		{"下班16:00", 16, 0, "out", "early"},
		{"下班16:59", 16, 59, "out", "early"},
		{"下班15:59", 15, 59, "out", "absent_half"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTime := time.Date(2026, 5, 13, tt.hour, tt.minute, 0, 0, time.Local)
			result := calculateStatus(testTime, tt.punchType)
			if result != tt.expected {
				t.Errorf("%s: expected %s, got %s", tt.name, tt.expected, result)
			}
		})
	}
}
