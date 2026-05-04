package common

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"
	"strings"
	"time"
)

func ValidateIDCard(idCard string) bool {
	idCard = strings.TrimSpace(strings.ToUpper(idCard))
	if len(idCard) != 18 {
		return false
	}
	for i, c := range idCard {
		if i == 17 {
			if c < '0' || c > '9' {
				if c != 'X' {
					return false
				}
			}
		} else {
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return true
}

func ValidatePhone(phone string) bool {
	phone = strings.TrimSpace(phone)
	if len(phone) != 11 {
		return false
	}
	if phone[0] != '1' {
		return false
	}
	matched, _ := regexp.MatchString(`^\d{11}$`, phone)
	return matched
}

func ValidateStudentName(name string) bool {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return false
	}
	runes := []rune(name)
	return len(runes) <= 20
}

func ValidatePlanName(name string) bool {
	name = strings.TrimSpace(name)
	if len(name) == 0 {
		return false
	}
	runes := []rune(name)
	return len(runes) <= 50
}

func ValidateAddress(address string) bool {
	address = strings.TrimSpace(address)
	return len(address) > 0
}

func GenerateID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes) + hex.EncodeToString([]byte{
		byte(timestamp >> 56),
		byte(timestamp >> 48),
		byte(timestamp >> 40),
		byte(timestamp >> 32),
		byte(timestamp >> 24),
		byte(timestamp >> 16),
		byte(timestamp >> 8),
		byte(timestamp),
	})
}
