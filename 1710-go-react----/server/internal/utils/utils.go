package utils

import (
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func ParseRatio(ratioStr string) []string {
	parts := strings.Split(ratioStr, ":")
	if len(parts) != 2 {
		return []string{"试验组", "对照组"}
	}
	return parts
}

func GenerateNextGroup(ratioStr string, currentCount int) string {
	parts := strings.Split(ratioStr, ":")
	if len(parts) != 2 {
		return "试验组"
	}
	ratio1, _ := strconv.Atoi(parts[0])
	ratio2, _ := strconv.Atoi(parts[1])
	totalRatio := ratio1 + ratio2
	if totalRatio == 0 {
		return "试验组"
	}
	pos := currentCount % totalRatio
	if pos < ratio1 {
		return "试验组"
	}
	return "对照组"
}

func GenerateRandomizationID(siteCode string, seq int) string {
	return fmt.Sprintf("%s-%03d", siteCode, seq)
}

func IsOutOfWindow(actualDate time.Time, windowDays, windowTolerance int, enrollmentDate time.Time) bool {
	if enrollmentDate.IsZero() {
		return false
	}
	expectedDate := enrollmentDate.AddDate(0, 0, windowDays)
	minDate := expectedDate.AddDate(0, 0, -windowTolerance)
	maxDate := expectedDate.AddDate(0, 0, windowTolerance)
	return actualDate.Before(minDate) || actualDate.After(maxDate)
}

func CanTransitionStatus(from, to string) bool {
	transitions := map[string][]string{
		"筛选中": {"入组"},
		"入组":   {"治疗中", "退出"},
		"治疗中": {"随访中", "退出"},
		"随访中": {"已完成", "退出"},
	}
	allowed, ok := transitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func WriteCSVResponse(c *gin.Context, filename string, rows [][]string) {
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "text/csv; charset=utf-8")

	writer := csv.NewWriter(c.Writer)
	writer.WriteAll(rows)
	writer.Flush()
}

func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}
