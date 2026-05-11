package etccore

import (
	"etc-system/common"
	"math"
)

func GetPassengerClass(seats int) common.PassengerClass {
	switch {
	case seats <= 7:
		return common.PassengerClass1
	case seats <= 19:
		return common.PassengerClass2
	case seats <= 39:
		return common.PassengerClass3
	default:
		return common.PassengerClass4
	}
}

func GetTruckClass(loadWeight float64) common.TruckClass {
	switch {
	case loadWeight <= 2:
		return common.TruckClass1
	case loadWeight <= 5:
		return common.TruckClass2
	case loadWeight <= 10:
		return common.TruckClass3
	case loadWeight <= 15:
		return common.TruckClass4
	default:
		return common.TruckClass5
	}
}

func GetPassengerPrice(class common.PassengerClass) float64 {
	switch class {
	case common.PassengerClass1:
		return 0.4
	case common.PassengerClass2:
		return 0.5
	case common.PassengerClass3:
		return 0.6
	case common.PassengerClass4:
		return 0.7
	default:
		return 0.4
	}
}

func GetTruckPrice(class common.TruckClass) float64 {
	switch class {
	case common.TruckClass1:
		return 0.45
	case common.TruckClass2:
		return 0.8
	case common.TruckClass3:
		return 1.1
	case common.TruckClass4:
		return 1.4
	case common.TruckClass5:
		return 1.6
	default:
		return 0.45
	}
}

func roundToCent(value float64) float64 {
	return math.Round(value*100) / 100
}

func CalculateFee(account *common.Account, mileage float64) float64 {
	if account.VehicleType == common.Passenger {
		class := GetPassengerClass(account.Seats)
		price := GetPassengerPrice(class)
		return roundToCent(mileage * price)
	} else if account.VehicleType == common.Truck {
		class := GetTruckClass(account.LoadWeight)
		price := GetTruckPrice(class)
		return roundToCent(mileage * price)
	}
	return 0
}
