package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math"
	"time"
)

func GenerateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func UniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

func CalculateProgress(completedCount, totalCount int) float64 {
	if totalCount == 0 {
		return 0
	}
	return float64(completedCount) / float64(totalCount) * 100
}

func DaysBetween(t1, t2 time.Time) int {
	diff := t2.Sub(t1)
	return int(diff.Hours() / 24)
}

func RoundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
