package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"health-archive/internal/models"
)

var idCardRegex = regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`)

func ValidateIDCard(idCard string) bool {
	if len(idCard) != 18 {
		return false
	}
	
	if !idCardRegex.MatchString(idCard) {
		return false
	}

	return true
}

func ExtractGenderFromIDCard(idCard string) (models.Gender, error) {
	if len(idCard) < 17 {
		return "", fmt.Errorf("invalid ID card length")
	}

	genderCode := idCard[16:17]
	num, err := strconv.Atoi(genderCode)
	if err != nil {
		return "", fmt.Errorf("failed to parse gender code: %w", err)
	}

	if num%2 == 1 {
		return models.GenderMale, nil
	}
	return models.GenderFemale, nil
}

func ExtractBirthDateFromIDCard(idCard string) (time.Time, error) {
	if len(idCard) < 14 {
		return time.Time{}, fmt.Errorf("invalid ID card length")
	}

	birthDateStr := idCard[6:14]
	birthDate, err := time.Parse("20060102", birthDateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse birth date: %w", err)
	}

	return birthDate, nil
}

func IsUnderage(birthDate time.Time) bool {
	age := CalculateAge(birthDate)
	return age < 18
}

func CalculateAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	if now.Month() < birthDate.Month() || 
	   (now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}

	return age
}

func NormalizeIDCard(idCard string) string {
	return strings.ToUpper(strings.TrimSpace(idCard))
}
