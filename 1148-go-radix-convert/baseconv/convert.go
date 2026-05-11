package baseconv

import (
	"math/big"
)

func Convert(s string, fromBase, toBase int, precision int) (string, error) {
	if err := ValidateNumber(s, fromBase); err != nil {
		return "", err
	}
	if err := ValidateBase(toBase); err != nil {
		return "", err
	}

	isNegative, intPart, fracPart, err := parseNumber(s)
	if err != nil {
		return "", err
	}

	intValue := big.NewInt(0)
	if intPart != "" {
		intValue, err = intPartToBigInt(intPart, fromBase)
		if err != nil {
			return "", err
		}
	}

	var fracValue float64
	if fracPart != "" {
		fracValue, err = fracPartToFloat64(fracPart, fromBase)
		if err != nil {
			return "", err
		}
	}

	if intValue.Sign() == 0 && fracValue == 0 {
		return "0", nil
	}

	var resultInt string
	if intValue.Sign() > 0 {
		resultInt, err = bigIntToIntPart(new(big.Int).Set(intValue), toBase)
		if err != nil {
			return "", err
		}
	} else {
		resultInt = "0"
	}

	var resultFrac string
	if fracValue > 0 {
		resultFrac, err = float64ToFracPart(fracValue, toBase, precision)
		if err != nil {
			return "", err
		}
	}

	var result string
	if resultFrac == "" {
		result = resultInt
	} else {
		result = resultInt + "." + resultFrac
	}

	if isNegative && (intValue.Sign() > 0 || fracValue > 0) {
		result = "-" + result
	}

	return result, nil
}

func ConvertDefault(s string, fromBase, toBase int) (string, error) {
	return Convert(s, fromBase, toBase, 32)
}
