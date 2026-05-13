package models

import (
	"errors"
	"time"
)

type Product struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	Category           string    `json:"category"`
	CurrentStock       int       `json:"current_stock"`
	SafetyStock        int       `json:"safety_stock"`
	PurchaseLeadTime   int       `json:"purchase_lead_time"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type AlertStatus string

const (
	AlertStatusPending     AlertStatus = "pending"
	AlertStatusProcessing  AlertStatus = "processing"
	AlertStatusAwaiting    AlertStatus = "awaiting"
	AlertStatusClosed      AlertStatus = "closed"
)

type AlertLevel string

const (
	AlertLevelYellow AlertLevel = "yellow"
	AlertLevelOrange AlertLevel = "orange"
	AlertLevelRed    AlertLevel = "red"
)

type Alert struct {
	ID              int64       `json:"id"`
	ProductID       int64       `json:"product_id"`
	ProductName     string      `json:"product_name"`
	Category        string      `json:"category"`
	CurrentStock    int         `json:"current_stock"`
	SafetyStock     int         `json:"safety_stock"`
	Level           AlertLevel  `json:"level"`
	Status          AlertStatus `json:"status"`
	AssignedTo      string      `json:"assigned_to"`
	SuggestedQty    int         `json:"suggested_qty"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

type CategoryStats struct {
	Category string `json:"category"`
	Count    int    `json:"count"`
}

func ValidateAlertStatus(status AlertStatus) error {
	switch status {
	case AlertStatusPending, AlertStatusProcessing, AlertStatusAwaiting, AlertStatusClosed:
		return nil
	default:
		return errors.New("invalid alert status")
	}
}

func CalculateAlertLevel(currentStock, safetyStock int) AlertLevel {
	if currentStock < 5 {
		return AlertLevelRed
	}
	if currentStock < safetyStock/2 {
		return AlertLevelOrange
	}
	return AlertLevelYellow
}

func CalculateSuggestedQty(currentStock, safetyStock int) int {
	suggested := int(float64(safetyStock)*1.5) - currentStock
	if suggested < 0 {
		return 0
	}
	return suggested
}
