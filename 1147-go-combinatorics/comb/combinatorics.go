package comb

import (
	"errors"
	"math/big"
)

var (
	ErrNegativeInput       = errors.New("n and k must be non-negative integers")
	ErrKGreaterThanN       = errors.New("k cannot be greater than n")
	ErrInvalidModulus      = errors.New("modulus must be greater than 1")
	ErrDuplicateCountZero  = errors.New("duplicate count must be positive")
)

func Permutation(n, k int64) (*big.Int, error) {
	if n < 0 || k < 0 {
		return nil, ErrNegativeInput
	}
	if k > n {
		return nil, ErrKGreaterThanN
	}
	if n == 0 {
		return big.NewInt(1), nil
	}
	if k == 0 {
		return big.NewInt(1), nil
	}

	result := big.NewInt(1)
	for i := n - k + 1; i <= n; i++ {
		result.Mul(result, big.NewInt(i))
	}
	return result, nil
}

func PermutationMod(n, k, mod int64) (int64, error) {
	if n < 0 || k < 0 {
		return 0, ErrNegativeInput
	}
	if k > n {
		return 0, ErrKGreaterThanN
	}
	if mod <= 1 {
		return 0, ErrInvalidModulus
	}
	if n == 0 {
		return 1 % mod, nil
	}
	if k == 0 {
		return 1 % mod, nil
	}

	result := int64(1)
	for i := n - k + 1; i <= n; i++ {
		result = (result * (i % mod)) % mod
	}
	return result, nil
}

func Combination(n, k int64) (*big.Int, error) {
	if n < 0 || k < 0 {
		return nil, ErrNegativeInput
	}
	if k > n {
		return nil, ErrKGreaterThanN
	}
	if n == 0 {
		return big.NewInt(1), nil
	}
	if k == 0 || k == n {
		return big.NewInt(1), nil
	}
	if k == 1 {
		return big.NewInt(n), nil
	}

	if k > n/2 {
		k = n - k
	}

	numerator := big.NewInt(1)
	for i := int64(1); i <= k; i++ {
		numerator.Mul(numerator, big.NewInt(n-k+i))
	}

	denominator := big.NewInt(1)
	for i := int64(2); i <= k; i++ {
		denominator.Mul(denominator, big.NewInt(i))
	}

	result := new(big.Int).Div(numerator, denominator)
	return result, nil
}

func CombinationMod(n, k, mod int64) (int64, error) {
	if n < 0 || k < 0 {
		return 0, ErrNegativeInput
	}
	if k > n {
		return 0, ErrKGreaterThanN
	}
	if mod <= 1 {
		return 0, ErrInvalidModulus
	}
	if n == 0 {
		return 1 % mod, nil
	}
	if k == 0 || k == n {
		return 1 % mod, nil
	}
	if k == 1 {
		return n % mod, nil
	}

	if k > n/2 {
		k = n - k
	}

	numerator := int64(1)
	for i := int64(1); i <= k; i++ {
		numerator = (numerator * ((n - k + i) % mod)) % mod
	}

	invDenominator := int64(1)
	for i := int64(2); i <= k; i++ {
		invDenominator = (invDenominator * modInverse(i, mod)) % mod
	}

	result := (numerator * invDenominator) % mod
	return result, nil
}

func modInverse(a, mod int64) int64 {
	return powMod(a, mod-2, mod)
}

func powMod(base, exp, mod int64) int64 {
	result := int64(1)
	base = base % mod
	for exp > 0 {
		if exp%2 == 1 {
			result = (result * base) % mod
		}
		base = (base * base) % mod
		exp = exp / 2
	}
	return result
}

type DuplicateCount struct {
	Count int64
}

func PermutationWithDuplicates(counts []int64) (*big.Int, error) {
	if len(counts) == 0 {
		return big.NewInt(1), nil
	}

	var total int64
	for _, c := range counts {
		if c <= 0 {
			return nil, ErrDuplicateCountZero
		}
		total += c
	}

	if total == 0 {
		return big.NewInt(1), nil
	}

	numerator := big.NewInt(1)
	for i := int64(2); i <= total; i++ {
		numerator.Mul(numerator, big.NewInt(i))
	}

	denominator := big.NewInt(1)
	for _, c := range counts {
		fact := big.NewInt(1)
		for i := int64(2); i <= c; i++ {
			fact.Mul(fact, big.NewInt(i))
		}
		denominator.Mul(denominator, fact)
	}

	result := new(big.Int).Div(numerator, denominator)
	return result, nil
}

func PermutationWithDuplicatesMod(counts []int64, mod int64) (int64, error) {
	if mod <= 1 {
		return 0, ErrInvalidModulus
	}
	if len(counts) == 0 {
		return 1 % mod, nil
	}

	var total int64
	for _, c := range counts {
		if c <= 0 {
			return 0, ErrDuplicateCountZero
		}
		total += c
	}

	if total == 0 {
		return 1 % mod, nil
	}

	numerator := int64(1)
	for i := int64(2); i <= total; i++ {
		numerator = (numerator * (i % mod)) % mod
	}

	invDenominator := int64(1)
	for _, c := range counts {
		fact := int64(1)
		for i := int64(2); i <= c; i++ {
			fact = (fact * (i % mod)) % mod
		}
		invDenominator = (invDenominator * modInverse(fact, mod)) % mod
	}

	result := (numerator * invDenominator) % mod
	return result, nil
}
