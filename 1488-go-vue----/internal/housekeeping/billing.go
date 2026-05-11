package housekeeping

import (
	"math"
)

const (
	BasePriceDailyCleaning   = 45
	BasePriceDeepCleaning    = 65
	BasePriceNanny           = 200
	BasePriceChildcare       = 180
	BasePriceElderlyCare     = 150
)

func GetBasePrice(category ServiceCategory) int {
	switch category {
	case ServiceDailyCleaning:
		return BasePriceDailyCleaning
	case ServiceDeepCleaning:
		return BasePriceDeepCleaning
	case ServiceNanny:
		return BasePriceNanny
	case ServiceChildcare:
		return BasePriceChildcare
	case ServiceElderlyCare:
		return BasePriceElderlyCare
	default:
		return 0
	}
}

func IsDailyService(category ServiceCategory) bool {
	return category == ServiceNanny || category == ServiceChildcare || category == ServiceElderlyCare
}

func CalculateBillingDuration(estimated, actual float64) float64 {
	if actual <= estimated+0.5 {
		return estimated
	}

	excess := actual - estimated
	excessSlots := math.Ceil((excess - 0.5) / 0.5)
	return estimated + excessSlots*0.5
}

func CalculatePrice(category ServiceCategory, billingDuration float64) int {
	basePrice := GetBasePrice(category)
	if basePrice == 0 {
		return 0
	}

	var price float64
	if IsDailyService(category) {
		days := math.Ceil(billingDuration)
		price = float64(basePrice) * days
	} else {
		price = float64(basePrice) * billingDuration
	}

	return int(math.Round(price))
}

func CalculateBookingPrice(category ServiceCategory, estimated, actual float64) int {
	billingDuration := CalculateBillingDuration(estimated, actual)
	return CalculatePrice(category, billingDuration)
}
