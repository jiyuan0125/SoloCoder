package core

import "math"

const (
	PassRateThreshold = 0.95
)

type TemperatureLimits struct {
	Min float64
	Max float64
}

type HumidityLimits struct {
	Min float64
	Max float64
}

var cargoConfig = map[CargoType]struct {
	tempLimits     TemperatureLimits
	humidityLimits HumidityLimits
	monitor        bool
}{
	CargoTypeFrozen: {
		tempLimits:     TemperatureLimits{Min: math.Inf(-1), Max: -18.0},
		humidityLimits: HumidityLimits{Min: 0, Max: 100},
		monitor:        true,
	},
	CargoTypeRefrigerated: {
		tempLimits:     TemperatureLimits{Min: 0.0, Max: 4.0},
		humidityLimits: HumidityLimits{Min: 0, Max: 100},
		monitor:        true,
	},
	CargoTypeMedicine: {
		tempLimits:     TemperatureLimits{Min: 15.0, Max: 25.0},
		humidityLimits: HumidityLimits{Min: 0, Max: 100},
		monitor:        true,
	},
	CargoTypeNormal: {
		tempLimits:     TemperatureLimits{Min: math.Inf(-1), Max: math.Inf(1)},
		humidityLimits: HumidityLimits{Min: 0, Max: 100},
		monitor:        false,
	},
}

func GetTemperatureLimits(cargoType CargoType) (TemperatureLimits, error) {
	config, exists := cargoConfig[cargoType]
	if !exists {
		return TemperatureLimits{}, &InvalidCargoTypeError{CargoType: string(cargoType)}
	}
	return config.tempLimits, nil
}

func GetHumidityLimits(cargoType CargoType) (HumidityLimits, error) {
	config, exists := cargoConfig[cargoType]
	if !exists {
		return HumidityLimits{}, &InvalidCargoTypeError{CargoType: string(cargoType)}
	}
	return config.humidityLimits, nil
}

func ShouldMonitor(cargoType CargoType) (bool, error) {
	config, exists := cargoConfig[cargoType]
	if !exists {
		return false, &InvalidCargoTypeError{CargoType: string(cargoType)}
	}
	return config.monitor, nil
}

func IsTemperatureWithinLimits(temp float64, limits TemperatureLimits) bool {
	return temp >= limits.Min && temp <= limits.Max
}

func IsHumidityWithinLimits(humidity float64, limits HumidityLimits) bool {
	return humidity >= limits.Min && humidity <= limits.Max
}

func ValidateCargoType(cargoType string) (CargoType, error) {
	switch CargoType(cargoType) {
	case CargoTypeFrozen, CargoTypeRefrigerated, CargoTypeMedicine, CargoTypeNormal:
		return CargoType(cargoType), nil
	default:
		return "", &InvalidCargoTypeError{CargoType: cargoType}
	}
}
