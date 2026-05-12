package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"skill-cert/models"
)

var idCardPattern = regexp.MustCompile(`^[1-9]\d{5}(19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`)

func ValidateIDCard(idCard string) bool {
	if !idCardPattern.MatchString(idCard) {
		return false
	}

	weights := []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	checkCodes := []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}

	var sum int
	for i := 0; i < 17; i++ {
		num, _ := strconv.Atoi(string(idCard[i]))
		sum += num * weights[i]
	}

	mod := sum % 11
	expectedCheck := checkCodes[mod]
	actualCheck := string(idCard[17])
	if actualCheck == "x" {
		actualCheck = "X"
	}

	return expectedCheck == actualCheck
}

func GetLevelCode(level models.SkillLevel) string {
	switch level {
	case models.LevelPrimary:
		return "1"
	case models.LevelIntermediate:
		return "2"
	case models.LevelAdvanced:
		return "3"
	case models.LevelTechnician:
		return "4"
	case models.LevelSeniorTech:
		return "5"
	default:
		return "0"
	}
}

func GenerateCertificateNo(year int, occupationCode string, level models.SkillLevel, serial int) (string, error) {
	if len(occupationCode) < 3 {
		return "", fmt.Errorf("职业代码长度不足3位")
	}
	if serial < 1 || serial > 99999 {
		return "", fmt.Errorf("流水号超出范围")
	}

	levelCode := GetLevelCode(level)
	return fmt.Sprintf("%04d%s%s%05d", year, occupationCode[:3], levelCode, serial), nil
}

func IsExpiringSoon(expiryDate time.Time) bool {
	threeMonthsLater := time.Now().AddDate(0, 3, 0)
	return expiryDate.Before(threeMonthsLater) && expiryDate.After(time.Now())
}

func NextBatchStatus(current models.BatchStatus) (models.BatchStatus, error) {
	switch current {
	case models.BatchNotStarted:
		return models.BatchInProgress, nil
	case models.BatchInProgress:
		return models.BatchCompleted, nil
	case models.BatchCompleted:
		return "", fmt.Errorf("批次已结束，无法继续流转")
	default:
		return "", fmt.Errorf("未知批次状态")
	}
}

func ValidateScoreRange(score float64) bool {
	return score >= 0 && score <= 100
}

func IsAllSubjectsPassed(scores map[models.ExamSubject]float64, subjects []models.ExamSubject) bool {
	for _, subject := range subjects {
		score, exists := scores[subject]
		if !exists || score < 60 {
			return false
		}
	}
	return true
}
