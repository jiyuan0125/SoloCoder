package common

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

var (
	idCounter int64
	idMutex   sync.Mutex
)

func GenerateID() string {
	idMutex.Lock()
	defer idMutex.Unlock()
	
	idCounter++
	
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	
	timestamp := time.Now().UnixNano()
	return hex.EncodeToString([]byte{
		byte(timestamp >> 56), byte(timestamp >> 48),
		byte(timestamp >> 40), byte(timestamp >> 32),
		byte(timestamp >> 24), byte(timestamp >> 16),
		byte(timestamp >> 8), byte(timestamp),
		byte(idCounter >> 24), byte(idCounter >> 16),
		byte(idCounter >> 8), byte(idCounter),
	})[:16]
}

func GenerateTaskID() string {
	return "task_" + GenerateID()
}

func GenerateExecutionID() string {
	return "exec_" + GenerateID()
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

func DurationMs(start, end time.Time) int64 {
	return end.Sub(start).Milliseconds()
}

func IsSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func IsWithinMonths(t time.Time, months int) bool {
	cutoff := time.Now().AddDate(0, -months, 0)
	return t.After(cutoff)
}
