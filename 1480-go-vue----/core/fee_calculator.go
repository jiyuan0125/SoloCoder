package core

import (
	"errors"

	"vehicle-inspection/common"
)

const (
	FeeSedan     = 18000
	FeeSUV       = 20000
	FeeMPV       = 20000
	FeeTruck     = 35000
	FeePassenger = 40000
)

func GetBaseFee(vehicleType common.VehicleType) (int, error) {
	switch vehicleType {
	case common.VehicleTypeSedan:
		return FeeSedan, nil
	case common.VehicleTypeSUV:
		return FeeSUV, nil
	case common.VehicleTypeMPV:
		return FeeMPV, nil
	case common.VehicleTypeTruck:
		return FeeTruck, nil
	case common.VehicleTypePassenger:
		return FeePassenger, nil
	default:
		return 0, errors.New("未知车辆类型")
	}
}

func CalculateRecheckFee(baseFee int, recheckCount int) int {
	if recheckCount <= 0 {
		return 0
	}
	if recheckCount == 1 {
		return baseFee / 2
	}
	return 0
}

func CalculateTotalFee(vehicleType common.VehicleType, recheckCount int) (int, error) {
	baseFee, err := GetBaseFee(vehicleType)
	if err != nil {
		return 0, err
	}
	
	totalFee := baseFee
	
	if recheckCount > 0 {
		recheckFee := CalculateRecheckFee(baseFee, recheckCount)
		totalFee += recheckFee
	}
	
	return totalFee, nil
}
