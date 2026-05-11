package core

import (
	"usedcar/api"
)

func CalculateFinalPrice(referencePriceFen int64, mileage float64, condition api.Condition) int64 {
	price := referencePriceFen

	if mileage > 10 {
		excessKms := mileage - 10
		excessUnits := int64(excessKms)
		deductionPercent := excessUnits * 1
		price = price * (100 - deductionPercent) / 100
	} else if mileage < 5 {
		price = price * 102 / 100
	}

	avgScore := float64(condition.Exterior+condition.Interior+condition.Engine+condition.Chassis) / 4.0
	if avgScore > 4 {
		price = price * 105 / 100
	} else if avgScore < 3 {
		price = price * 90 / 100
	}

	return roundToHundredYuan(price)
}

func roundToHundredYuan(priceFen int64) int64 {
	hundredFen := int64(10000)
	remainder := priceFen % hundredFen
	if remainder >= hundredFen/2 {
		return priceFen + (hundredFen - remainder)
	}
	return priceFen - remainder
}

func CalculateDepositAmount(evaluationPriceFen int64) int64 {
	deposit := evaluationPriceFen * 5 / 100
	minDeposit := int64(50000)
	maxDeposit := int64(1000000)

	if deposit < minDeposit {
		return minDeposit
	} else if deposit > maxDeposit {
		return maxDeposit
	}
	return deposit
}
