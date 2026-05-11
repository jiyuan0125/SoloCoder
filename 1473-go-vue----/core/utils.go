package core

import (
	"fmt"
	"strings"
	"time"
)

func ParseTime(timeStr string) (int, int, error) {
	var hour, minute int
	_, err := fmt.Sscanf(timeStr, "%02d:%02d", &hour, &minute)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid time format: %s, expected HH:MM", timeStr)
	}
	if hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("invalid hour: %d", hour)
	}
	if minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid minute: %d", minute)
	}
	return hour, minute, nil
}

func FormatTime(hour, minute int) string {
	return fmt.Sprintf("%02d:%02d", hour, minute)
}

func IsTimeInRange(hour, minute, startHour, startMinute, endHour, endMinute int) bool {
	current := hour*60 + minute
	start := startHour*60 + startMinute
	end := endHour*60 + endMinute
	if start <= end {
		return current >= start && current <= end
	}
	return current >= start || current <= end
}

func NormalizeStationName(name string) string {
	return strings.ReplaceAll(name, " ", "")
}

func GetCurrentTime() (int, int) {
	now := time.Now()
	return now.Hour(), now.Minute()
}
