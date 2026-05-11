package bigmath

import (
	"errors"
	"math/big"
	"strings"
)

var (
	ErrInvalidInput = errors.New("invalid input: empty or contains non-digit characters")
	ErrDivideByZero = errors.New("division by zero")
)

func normalize(s string) string {
	if len(s) == 0 {
		return ""
	}

	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	}

	s = strings.TrimLeft(s, "0")

	if len(s) == 0 {
		return "0"
	}

	if negative {
		return "-" + s
	}

	return s
}

func validate(s string) bool {
	if len(s) == 0 {
		return false
	}

	start := 0
	if s[0] == '-' {
		if len(s) == 1 {
			return false
		}
		start = 1
	}

	for i := start; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}

	return true
}

func parse(s string) (*big.Int, error) {
	if !validate(s) {
		return nil, ErrInvalidInput
	}

	n := new(big.Int)
	n.SetString(normalize(s), 10)
	return n, nil
}

func Add(a, b string) (string, error) {
	na, err := parse(a)
	if err != nil {
		return "", err
	}

	nb, err := parse(b)
	if err != nil {
		return "", err
	}

	result := new(big.Int).Add(na, nb)
	return normalize(result.String()), nil
}

func Subtract(a, b string) (string, error) {
	na, err := parse(a)
	if err != nil {
		return "", err
	}

	nb, err := parse(b)
	if err != nil {
		return "", err
	}

	result := new(big.Int).Sub(na, nb)
	return normalize(result.String()), nil
}

func Multiply(a, b string) (string, error) {
	na, err := parse(a)
	if err != nil {
		return "", err
	}

	nb, err := parse(b)
	if err != nil {
		return "", err
	}

	result := new(big.Int).Mul(na, nb)
	return normalize(result.String()), nil
}

func Divide(a, b string) (quotient string, remainder string, err error) {
	na, err := parse(a)
	if err != nil {
		return "", "", err
	}

	nb, err := parse(b)
	if err != nil {
		return "", "", err
	}

	if nb.Sign() == 0 {
		return "", "", ErrDivideByZero
	}

	q := new(big.Int).Quo(na, nb)
	r := new(big.Int).Rem(na, nb)

	return normalize(q.String()), normalize(r.String()), nil
}
