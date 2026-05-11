package core

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getPhoneLast4(phone string) string {
	if len(phone) < 4 {
		return phone
	}
	return phone[len(phone)-4:]
}

func is24HoursBefore(t time.Time) bool {
	now := time.Now()
	return now.Before(t.Add(-24 * time.Hour))
}

func isWithinCheckInWindow(startTime time.Time) bool {
	now := time.Now()
	windowStart := startTime.Add(-30 * time.Minute)
	windowEnd := startTime.Add(15 * time.Minute)
	return now.After(windowStart) && now.Before(windowEnd)
}

func isLate(startTime time.Time) bool {
	now := time.Now()
	return now.After(startTime.Add(15 * time.Minute))
}

func isFeedbackPeriodEnded(endTime time.Time) bool {
	now := time.Now()
	return now.After(endTime.Add(72 * time.Hour))
}

func isActivityEnded(endTime time.Time) bool {
	now := time.Now()
	return now.After(endTime)
}
