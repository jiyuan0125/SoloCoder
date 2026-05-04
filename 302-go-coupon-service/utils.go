package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func GenerateID() string {
	timestamp := time.Now().UnixNano()
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	return fmt.Sprintf("%d%s", timestamp, hex.EncodeToString(randomBytes))
}

func GenerateCouponCode() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func IsValidNow(validFrom, validTo time.Time) bool {
	now := time.Now()
	return now.After(validFrom) && now.Before(validTo.Add(time.Second))
}

func IsThresholdMet(orderAmount, thresholdAmount float64) bool {
	return orderAmount >= thresholdAmount
}
