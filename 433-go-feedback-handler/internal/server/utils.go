package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"

	"go-feedback-handler/pkg/protocol"
)

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateShortID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func calculateSimilarity(s1, s2 string) float64 {
	s1 = strings.ToLower(strings.TrimSpace(s1))
	s2 = strings.ToLower(strings.TrimSpace(s2))

	if s1 == s2 {
		return 1.0
	}

	len1 := len(s1)
	len2 := len(s2)

	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	set1 := make(map[string]int)
	set2 := make(map[string]int)

	for i := 0; i < len1-1; i++ {
		bigram := s1[i : i+2]
		set1[bigram]++
	}

	for i := 0; i < len2-1; i++ {
		bigram := s2[i : i+2]
		set2[bigram]++
	}

	intersection := 0
	union := 0

	for k, v1 := range set1 {
		if v2, exists := set2[k]; exists {
			intersection += min(v1, v2)
		}
	}

	for _, v := range set1 {
		union += v
	}
	for _, v := range set2 {
		union += v
	}

	if union == 0 {
		return 0.0
	}

	return float64(2*intersection) / float64(union)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func hoursSince(t time.Time) float64 {
	return time.Since(t).Hours()
}

func daysSince(t time.Time) float64 {
	return time.Since(t).Hours() / 24
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	days := hours / 24
	remainingHours := hours % 24

	if days > 0 {
		return fmt.Sprintf("%d天%d小时", days, remainingHours)
	}
	return fmt.Sprintf("%d小时", hours)
}

func getMonthKey(t time.Time) string {
	return fmt.Sprintf("%d-%02d", t.Year(), t.Month())
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func roundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func feedbackToTagIDs(fb *protocol.Feedback) []string {
	return fb.TagIDs
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	result := make([]string, 0, len(slice))
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}
