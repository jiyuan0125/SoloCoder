package creditcard

import (
	"math/rand"
)

func GenerateCardNumber(ct CardType) (string, error) {
	if !IsSupportedCardType(ct) {
		return "", ErrUnsupportedCardType
	}

	prefix, totalLength := selectIINAndLength(ct)

	prefixLen := len(prefix)
	remainingDigits := totalLength - prefixLen - 1

	digits := make([]int, 0, totalLength)
	for _, c := range prefix {
		digits = append(digits, int(c-'0'))
	}

	for i := 0; i < remainingDigits; i++ {
		digits = append(digits, rand.Intn(10))
	}

	checkDigit := ComputeLuhnCheckDigit(digits)
	digits = append(digits, checkDigit)

	result := make([]byte, totalLength)
	for i, d := range digits {
		result[i] = byte(d) + '0'
	}

	return string(result), nil
}

func GenerateCardNumbers(ct CardType, count int) ([]string, error) {
	if count <= 0 {
		return nil, ErrInvalidCount
	}
	if count > MaxGenerateCount {
		return nil, ErrCountExceedsLimit
	}
	if !IsSupportedCardType(ct) {
		return nil, ErrUnsupportedCardType
	}

	cards := make([]string, count)
	for i := 0; i < count; i++ {
		card, err := GenerateCardNumber(ct)
		if err != nil {
			return nil, err
		}
		cards[i] = card
	}

	return cards, nil
}
