package utils

import (
	"github.com/google/uuid"
)

func NewUUID() string {
	return uuid.New().String()
}

func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
