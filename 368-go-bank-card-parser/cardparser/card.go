package cardparser

import (
	"regexp"
	"strings"
)

type CardType string

const (
	CardTypeDebit   CardType = "借记卡"
	CardTypeCredit  CardType = "信用卡"
	CardTypeUnknown CardType = "未知"
)

type CardInfo struct {
	CardNumber    string
	BankName      string
	CardType      CardType
	IsValid       bool
	Error         string
	MaskedNumber  string
	FormattedNumber string
}

func SanitizeCardNumber(input string) string {
	re := regexp.MustCompile(`[\s\-]`)
	return re.ReplaceAllString(input, "")
}

func IsAllSameDigit(number string) bool {
	if len(number) == 0 {
		return false
	}
	first := number[0]
	for i := 1; i < len(number); i++ {
		if number[i] != first {
			return false
		}
	}
	return true
}

func IsAllZero(number string) bool {
	for i := 0; i < len(number); i++ {
		if number[i] != '0' {
			return false
		}
	}
	return true
}

func IsValidLength(number string) bool {
	length := len(number)
	return length >= 16 && length <= 19
}

func LuhnCheck(number string) bool {
	if len(number) < 2 {
		return false
	}

	sum := 0
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')
		if digit < 0 || digit > 9 {
			return false
		}

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		double = !double
	}

	return sum%10 == 0
}

func MaskCardNumber(number string) string {
	if len(number) <= 4 {
		return number
	}
	masked := strings.Repeat("*", len(number)-4) + number[len(number)-4:]
	return masked
}

func FormatCardNumber(number string) string {
	if len(number) == 0 {
		return ""
	}

	var builder strings.Builder
	for i := 0; i < len(number); i++ {
		if i > 0 && i%4 == 0 {
			builder.WriteRune(' ')
		}
		builder.WriteByte(number[i])
	}
	return builder.String()
}

func ParseCardNumber(input string) *CardInfo {
	sanitized := SanitizeCardNumber(input)
	re := regexp.MustCompile(`^\d+$`)
	if !re.MatchString(sanitized) {
		return &CardInfo{
			CardNumber:    sanitized,
			IsValid:       false,
			Error:         "卡号包含非数字字符",
		}
	}

	if IsAllSameDigit(sanitized) || IsAllZero(sanitized) {
		return &CardInfo{
			CardNumber:    sanitized,
			IsValid:       false,
			Error:         "无效的银行卡号",
		}
	}

	if !IsValidLength(sanitized) {
		return &CardInfo{
			CardNumber:    sanitized,
			IsValid:       false,
			Error:         "卡号长度无效，应为16-19位",
		}
	}

	if !LuhnCheck(sanitized) {
		return &CardInfo{
			CardNumber:    sanitized,
			IsValid:       false,
			Error:         "Luhn校验失败",
		}
	}

	bankName, cardType := LookupBankAndType(sanitized)

	return &CardInfo{
		CardNumber:       sanitized,
		BankName:         bankName,
		CardType:         cardType,
		IsValid:          true,
		MaskedNumber:     MaskCardNumber(sanitized),
		FormattedNumber:  FormatCardNumber(sanitized),
	}
}
