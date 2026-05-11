package cardverify

import (
	"strings"
	"unicode"
)

const validSeparators = " -_"

func CleanCardNumber(raw string) (string, error) {
	if raw == "" {
		return "", ErrEmptyCardNumber
	}

	var builder strings.Builder
	for _, r := range raw {
		if unicode.IsDigit(r) {
			builder.WriteRune(r)
		} else if strings.ContainsRune(validSeparators, r) {
			continue
		} else {
			return "", ErrInvalidCharacter
		}
	}

	cleaned := builder.String()
	if cleaned == "" {
		return "", ErrEmptyCardNumber
	}

	return cleaned, nil
}

func ValidateCardNumber(raw string) (*ValidateResult, error) {
	result := &ValidateResult{}
	result.RawNumber = raw

	cleaned, err := CleanCardNumber(raw)
	if err != nil {
		result.IsValid = false
		result.ErrorReason = err.Error()
		return result, err
	}
	result.CleanNumber = cleaned

	if len(cleaned) < 13 || len(cleaned) > 19 {
		result.IsValid = false
		result.ErrorReason = ErrInvalidLength.Error()
		return result, ErrInvalidLength
	}

	org := DetectCardOrganization(cleaned)
	result.CardOrganization = org

	binInfo, binFound := LookupBIN(cleaned)
	if binFound {
		result.BIN = binInfo.BIN
		result.BankName = binInfo.BankName
		result.CardType = binInfo.CardType
	} else {
		result.BIN = getBINRange(cleaned)
		result.BankName = "未知"
		result.CardType = CardTypeUnknown
	}

	if !LuhnCheck(cleaned) {
		result.IsValid = false
		result.ErrorReason = ErrLuhnCheckFailed.Error()
		return result, ErrLuhnCheckFailed
	}

	result.IsValid = true
	return result, nil
}

func getBINRange(number string) string {
	if len(number) >= 8 {
		return number[:8]
	}
	if len(number) >= 6 {
		return number[:6]
	}
	return number
}
