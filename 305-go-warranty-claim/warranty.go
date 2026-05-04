package main

import "time"

func CalculateWarrantyExpiry(purchaseDate time.Time) time.Time {
	if purchaseDate.Month() == time.February && purchaseDate.Day() == 29 {
		nextYear := purchaseDate.Year() + 1
		return time.Date(nextYear, time.February, 28, 0, 0, 0, 0, purchaseDate.Location())
	}
	return purchaseDate.AddDate(1, 0, 0)
}

func IsUnderWarranty(purchaseDate time.Time, currentDate time.Time) bool {
	expiryDate := CalculateWarrantyExpiry(purchaseDate)
	return currentDate.Before(expiryDate) || currentDate.Equal(expiryDate)
}

func ValidatePurchaseDate(purchaseDate time.Time, currentDate time.Time) error {
	if purchaseDate.After(currentDate) {
		return &ValidationError{
			Field:   "purchase_date",
			Message: "购买日期不能晚于当前日期",
		}
	}
	return nil
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
