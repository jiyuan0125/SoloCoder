package statemachine

import (
	"fmt"

	"qc-process/internal/models"
)

func GetStatusIndex(status models.QCStatus) int {
	for i, s := range models.StatusOrder {
		if s == status {
			return i
		}
	}
	return -1
}

func CanTransition(current, next models.QCStatus) bool {
	currentIdx := GetStatusIndex(current)
	nextIdx := GetStatusIndex(next)

	if currentIdx == -1 || nextIdx == -1 {
		return false
	}

	return nextIdx == currentIdx+1
}

func ValidateTransition(current, next models.QCStatus) error {
	if !CanTransition(current, next) {
		return fmt.Errorf("错误：无法从状态 %q 跳转到 %q，跳步操作不允许。当前状态: %s", current, next, current)
	}
	return nil
}

func ValidateAndTransition(current, next models.QCStatus) (models.QCStatus, error) {
	if err := ValidateTransition(current, next); err != nil {
		return current, err
	}
	return next, nil
}
