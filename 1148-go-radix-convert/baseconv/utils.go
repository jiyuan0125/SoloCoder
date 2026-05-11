package baseconv

import (
	"errors"
	"math/big"
	"strings"
)

const (
	charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	MinBase = 2
	MaxBase = 62
)

func charToValue(c byte) (int, error) {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0'), nil
	case c >= 'A' && c <= 'Z':
		return int(c - 'A' + 10), nil
	case c >= 'a' && c <= 'z':
		return int(c - 'a' + 36), nil
	default:
		return 0, errors.New("invalid character: " + string(c))
	}
}

func valueToChar(v int) (byte, error) {
	if v < 0 || v >= len(charset) {
		return 0, errors.New("invalid value")
	}
	return charset[v], nil
}

func isValidForBase(c byte, base int) bool {
	val, err := charToValue(c)
	if err != nil {
		return false
	}
	return val < base
}

func isSign(c byte) bool {
	return c == '+' || c == '-'
}

func parseNumber(s string) (isNegative bool, intPart, fracPart string, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return false, "", "", errors.New("empty string")
	}

	idx := 0
	if isSign(s[0]) {
		if s[0] == '-' {
			isNegative = true
		}
		idx++
	}

	parts := strings.SplitN(s[idx:], ".", 2)
	intPart = parts[0]
	if len(parts) > 1 {
		fracPart = parts[1]
	}

	if intPart == "" && fracPart == "" {
		return false, "", "", errors.New("no digits")
	}

	return isNegative, intPart, fracPart, nil
}

func intPartToBigInt(s string, base int) (*big.Int, error) {
	result := big.NewInt(0)
	baseBig := big.NewInt(int64(base))

	for i := 0; i < len(s); i++ {
		c := s[i]
		if !isValidForBase(c, base) {
			return nil, errors.New("invalid character for base " + string(rune(base+'0')))
		}
		val, _ := charToValue(c)
		result.Mul(result, baseBig)
		result.Add(result, big.NewInt(int64(val)))
	}

	return result, nil
}

func bigIntToIntPart(n *big.Int, base int) (string, error) {
	if n.Sign() == 0 {
		return "0", nil
	}

	baseBig := big.NewInt(int64(base))
	var chars []byte
	zero := big.NewInt(0)

	for n.Cmp(zero) > 0 {
		rem := new(big.Int)
		n.DivMod(n, baseBig, rem)
		c, err := valueToChar(int(rem.Int64()))
		if err != nil {
			return "", err
		}
		chars = append(chars, c)
	}

	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}

	return string(chars), nil
}

func fracPartToFloat64(s string, base int) (float64, error) {
	result := 0.0
	baseFloat := float64(base)
	divisor := baseFloat

	for i := 0; i < len(s); i++ {
		c := s[i]
		if !isValidForBase(c, base) {
			return 0, errors.New("invalid character for base " + string(rune(base+'0')))
		}
		val, _ := charToValue(c)
		result += float64(val) / divisor
		divisor *= baseFloat
	}

	return result, nil
}

func float64ToFracPart(f float64, base int, precision int) (string, error) {
	if precision <= 0 {
		precision = 32
	}

	var chars []byte
	baseFloat := float64(base)
	current := f

	for i := 0; i < precision && current > 0; i++ {
		current *= baseFloat
		val := int(current)
		c, err := valueToChar(val)
		if err != nil {
			return "", err
		}
		chars = append(chars, c)
		current -= float64(val)
	}

	return string(chars), nil
}
