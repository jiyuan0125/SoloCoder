package core

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"sync/atomic"
)

func roundToTwoDecimals(value float64) float64 {
	return math.Round(value*100) / 100
}

func roundPrice(price float64) float64 {
	return roundToTwoDecimals(price)
}

func calculateTotalPrice(materials []Material, items []QuotationItem) float64 {
	materialMap := make(map[string]Material)
	for _, m := range materials {
		materialMap[m.ID] = m
	}

	var total float64
	for _, item := range items {
		if material, ok := materialMap[item.MaterialID]; ok {
			amount := item.UnitPrice * material.Quantity
			total += amount
		}
	}
	return roundToTwoDecimals(total)
}

var counter int64

func generateID(prefix string) string {
	num := atomic.AddInt64(&counter, 1)
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	return fmt.Sprintf("%s-%d-%s", prefix, num, hex.EncodeToString(randomBytes))
}
