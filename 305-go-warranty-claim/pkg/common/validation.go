package common

import (
	"regexp"
	"time"
)

func IsValidSerialNumber(serial string) bool {
	if len(serial) != SerialNumberLength {
		return false
	}
	match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", serial)
	return match
}

func IsValidDescription(desc string) bool {
	if len(desc) == 0 {
		return false
	}
	if len(desc) > MaxDescriptionLength {
		return false
	}
	return true
}

func IsValidPurchaseDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

func IsPurchaseDateNotFuture(dateStr string) bool {
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return !date.After(today)
}

func CalculateWarrantyExpiry(purchaseDateStr string) string {
	purchaseDate, err := time.Parse("2006-01-02", purchaseDateStr)
	if err != nil {
		return ""
	}
	
	isLeapDay := purchaseDate.Month() == time.February && purchaseDate.Day() == 29
	
	expiryYear := purchaseDate.Year() + 1
	
	if isLeapDay {
		return time.Date(expiryYear, time.February, 28, 0, 0, 0, 0, purchaseDate.Location()).Format("2006-01-02")
	}
	
	expiryDate := time.Date(expiryYear, purchaseDate.Month(), purchaseDate.Day(), 0, 0, 0, 0, purchaseDate.Location())
	return expiryDate.Format("2006-01-02")
}

func IsWithinWarranty(purchaseDateStr string) bool {
	expiryStr := CalculateWarrantyExpiry(purchaseDateStr)
	if expiryStr == "" {
		return false
	}
	
	expiryDate, _ := time.Parse("2006-01-02", expiryStr)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	
	return !today.After(expiryDate)
}
