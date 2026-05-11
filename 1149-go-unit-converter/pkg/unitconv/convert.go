package unitconv

import (
	"fmt"
	"math"
)

func Convert(value float64, from, to string, tempDelta bool) (float64, error) {
	parsedFrom, err := ParseUnit(from)
	if err != nil {
		return 0, err
	}

	parsedTo, err := ParseUnit(to)
	if err != nil {
		return 0, err
	}

	if parsedFrom.IsCompound() || parsedTo.IsCompound() {
		if !parsedFrom.IsCompound() || !parsedTo.IsCompound() {
			return 0, fmt.Errorf("cannot convert between compound and non-compound units: %s -> %s", from, to)
		}
		return convertCompound(value, parsedFrom, parsedTo, tempDelta)
	}

	return convertSimple(value, parsedFrom.SimpleUnit, parsedTo.SimpleUnit, tempDelta)
}

func convertSimple(value float64, from, to string, tempDelta bool) (float64, error) {
	fromInfo, fromOK := GetUnitInfo(from)
	toInfo, toOK := GetUnitInfo(to)

	if !fromOK {
		return 0, fmt.Errorf("unsupported unit: %s. %s", from, GetAllSupportedUnits())
	}
	if !toOK {
		return 0, fmt.Errorf("unsupported unit: %s. %s", to, GetAllSupportedUnits())
	}

	if fromInfo.UnitType != toInfo.UnitType {
		return 0, fmt.Errorf("unit type mismatch: %s is %s, %s is %s", from, fromInfo.UnitType, to, toInfo.UnitType)
	}

	if fromInfo.UnitType == UnitTypeTemp {
		if tempDelta {
			return convertTempDelta(value, fromInfo, toInfo), nil
		}
		return convertTempPoint(value, fromInfo, toInfo), nil
	}

	baseValue := value * fromInfo.BaseValue
	result := baseValue / toInfo.BaseValue
	return result, nil
}

func convertTempPoint(value float64, from, to *UnitInfo) float64 {
	kelvin := (value - from.Offset) * from.BaseValue + 273.15
	result := (kelvin - 273.15)/to.BaseValue + to.Offset
	return result
}

func convertTempDelta(value float64, from, to *UnitInfo) float64 {
	return value * from.BaseValue / to.BaseValue
}

func convertCompound(value float64, from, to *ParsedUnit, tempDelta bool) (float64, error) {
	if len(from.Numerator) != len(to.Numerator) || len(from.Denominator) != len(to.Denominator) {
		return 0, fmt.Errorf("compound unit structure mismatch: %s -> %s", from.Original, to.Original)
	}

	factor := 1.0

	for i := range from.Numerator {
		if from.NumeratorExponents[i] != to.NumeratorExponents[i] {
			return 0, fmt.Errorf("compound unit structure mismatch: %s -> %s", from.Original, to.Original)
		}
		partFactor, err := getConversionFactor(from.Numerator[i], to.Numerator[i], tempDelta)
		if err != nil {
			return 0, err
		}
		factor *= math.Pow(partFactor, float64(from.NumeratorExponents[i]))
	}

	for i := range from.Denominator {
		if from.DenominatorExponents[i] != to.DenominatorExponents[i] {
			return 0, fmt.Errorf("compound unit structure mismatch: %s -> %s", from.Original, to.Original)
		}
		partFactor, err := getConversionFactor(from.Denominator[i], to.Denominator[i], tempDelta)
		if err != nil {
			return 0, err
		}
		factor /= math.Pow(partFactor, float64(from.DenominatorExponents[i]))
	}

	return value * factor, nil
}

func getConversionFactor(from, to string, tempDelta bool) (float64, error) {
	fromInfo, fromOK := GetUnitInfo(from)
	toInfo, toOK := GetUnitInfo(to)

	if !fromOK {
		return 0, fmt.Errorf("unsupported unit: %s. %s", from, GetAllSupportedUnits())
	}
	if !toOK {
		return 0, fmt.Errorf("unsupported unit: %s. %s", to, GetAllSupportedUnits())
	}

	if fromInfo.UnitType != toInfo.UnitType {
		return 0, fmt.Errorf("unit type mismatch: %s is %s, %s is %s", from, fromInfo.UnitType, to, toInfo.UnitType)
	}

	if fromInfo.UnitType == UnitTypeTemp {
		if !tempDelta {
			return 0, fmt.Errorf("temperature units in compound conversion must use delta mode (tempDelta=true)")
		}
		return fromInfo.BaseValue / toInfo.BaseValue, nil
	}

	return fromInfo.BaseValue / toInfo.BaseValue, nil
}
