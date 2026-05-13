package utils

import (
	"regexp"
	"strings"
)

func MaskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	reg := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if !reg.MatchString(phone) {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

func MaskEmail(email string) string {
	email = strings.TrimSpace(email)
	atIndex := strings.Index(email, "@")
	if atIndex <= 1 {
		return email
	}

	local := email[:atIndex]
	domain := email[atIndex:]

	if len(local) == 1 {
		return email
	}

	masked := string(local[0])
	for i := 0; i < len(local)-2; i++ {
		masked += "*"
	}
	masked += string(local[len(local)-1])

	return masked + domain
}

func MaskValue(fieldName, value string) string {
	fieldLower := strings.ToLower(fieldName)
	if strings.Contains(fieldLower, "phone") || strings.Contains(fieldLower, "mobile") {
		return MaskPhone(value)
	}
	if strings.Contains(fieldLower, "email") || strings.Contains(fieldLower, "mail") {
		return MaskEmail(value)
	}
	return value
}
