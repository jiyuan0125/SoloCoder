package statemachine

import (
	"order-state-machine/models"
	"strings"
)

type StateTransitionError struct {
	FromStatus   string
	FromName     string
	ToStatus     string
	ToName       string
	ErrorMessage string
}

func (e *StateTransitionError) Error() string {
	if e.ErrorMessage != "" {
		return e.ErrorMessage
	}
	return "非法状态流转: 当前状态为[" + e.FromName + "]，无法跳转到[" + e.ToName + "]"
}

func getNextStatus(flow []string, current string) (string, bool) {
	for i, s := range flow {
		if s == current {
			if i < len(flow)-1 {
				return flow[i+1], true
			}
			return "", false
		}
	}
	return "", false
}

func getStatusIndex(flow []string, status string) int {
	for i, s := range flow {
		if s == status {
			return i
		}
	}
	return -1
}

func ValidateOrderStatusTransition(currentStatus, toStatus string) error {
	if currentStatus == models.OrderStatusCancelled {
		return &StateTransitionError{
			FromStatus:   currentStatus,
			FromName:     models.GetStatusName(currentStatus),
			ToStatus:     toStatus,
			ToName:       models.GetStatusName(toStatus),
			ErrorMessage: "订单已取消，无法进行状态变更",
		}
	}

	currentIdx := getStatusIndex(models.OrderStatusFlow, currentStatus)
	toIdx := getStatusIndex(models.OrderStatusFlow, toStatus)

	if currentIdx == -1 || toIdx == -1 {
		return &StateTransitionError{
			FromStatus:   currentStatus,
			FromName:     models.GetStatusName(currentStatus),
			ToStatus:     toStatus,
			ToName:       models.GetStatusName(toStatus),
		}
	}

	if toIdx != currentIdx+1 {
		return &StateTransitionError{
			FromStatus:   currentStatus,
			FromName:     models.GetStatusName(currentStatus),
			ToStatus:     toStatus,
			ToName:       models.GetStatusName(toStatus),
		}
	}

	return nil
}

func GetNextOrderStatus(currentStatus string) (string, bool) {
	return getNextStatus(models.OrderStatusFlow, currentStatus)
}

func ValidateRepairStatusTransition(currentStatus, toStatus string) error {
	currentIdx := getStatusIndex(models.RepairStatusFlow, currentStatus)
	toIdx := getStatusIndex(models.RepairStatusFlow, toStatus)

	if currentIdx == -1 || toIdx == -1 {
		return &StateTransitionError{
			FromStatus:   currentStatus,
			FromName:     models.GetStatusName(currentStatus),
			ToStatus:     toStatus,
			ToName:       models.GetStatusName(toStatus),
		}
	}

	if toIdx != currentIdx+1 {
		return &StateTransitionError{
			FromStatus:   currentStatus,
			FromName:     models.GetStatusName(currentStatus),
			ToStatus:     toStatus,
			ToName:       models.GetStatusName(toStatus),
		}
	}

	return nil
}

func GetNextRepairStatus(currentStatus string) (string, bool) {
	return getNextStatus(models.RepairStatusFlow, currentStatus)
}

func ValidateExchangeStatusTransition(currentStatus, toStatus string) error {
	currentIdx := getStatusIndex(models.ExchangeStatusFlow, currentStatus)
	toIdx := getStatusIndex(models.ExchangeStatusFlow, toStatus)

	if currentIdx == -1 || toIdx == -1 {
		return &StateTransitionError{
			FromStatus:   currentStatus,
			FromName:     models.GetStatusName(currentStatus),
			ToStatus:     toStatus,
			ToName:       models.GetStatusName(toStatus),
		}
	}

	if toIdx != currentIdx+1 {
		return &StateTransitionError{
			FromStatus:   currentStatus,
			FromName:     models.GetStatusName(currentStatus),
			ToStatus:     toStatus,
			ToName:       models.GetStatusName(toStatus),
		}
	}

	return nil
}

func GetNextExchangeStatus(currentStatus string) (string, bool) {
	return getNextStatus(models.ExchangeStatusFlow, currentStatus)
}

func ValidateRefundStatusTransition(currentStatus, toStatus string) error {
	currentIdx := getStatusIndex(models.RefundStatusFlow, currentStatus)
	toIdx := getStatusIndex(models.RefundStatusFlow, toStatus)

	if currentIdx == -1 || toIdx == -1 {
		return &StateTransitionError{
			FromStatus:   currentStatus,
			FromName:     models.GetStatusName(currentStatus),
			ToStatus:     toStatus,
			ToName:       models.GetStatusName(toStatus),
		}
	}

	if toIdx != currentIdx+1 {
		return &StateTransitionError{
			FromStatus:   currentStatus,
			FromName:     models.GetStatusName(currentStatus),
			ToStatus:     toStatus,
			ToName:       models.GetStatusName(toStatus),
		}
	}

	return nil
}

func GetNextRefundStatus(currentStatus string) (string, bool) {
	return getNextStatus(models.RefundStatusFlow, currentStatus)
}

func ValidateAfterSalesStatusTransition(afterSalesType, currentStatus, toStatus string) error {
	afterSalesType = strings.ToLower(afterSalesType)
	switch afterSalesType {
	case models.AfterSalesTypeRepair:
		return ValidateRepairStatusTransition(currentStatus, toStatus)
	case models.AfterSalesTypeExchange:
		return ValidateExchangeStatusTransition(currentStatus, toStatus)
	case models.AfterSalesTypeRefund:
		return ValidateRefundStatusTransition(currentStatus, toStatus)
	default:
		return &StateTransitionError{
			ErrorMessage: "未知的售后类型: " + afterSalesType,
		}
	}
}

func GetAfterSalesInitialStatus(afterSalesType string) (string, error) {
	afterSalesType = strings.ToLower(afterSalesType)
	switch afterSalesType {
	case models.AfterSalesTypeRepair:
		return models.RepairStatusApplied, nil
	case models.AfterSalesTypeExchange:
		return models.ExchangeStatusApplied, nil
	case models.AfterSalesTypeRefund:
		return models.RefundStatusApplied, nil
	default:
		return "", &StateTransitionError{
			ErrorMessage: "未知的售后类型: " + afterSalesType,
		}
	}
}
