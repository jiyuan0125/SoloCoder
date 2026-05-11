package iban

import (
	"fmt"
	"math/big"
	"strings"
	"unicode"
)

type CountryRule struct {
	CountryCode string
	CountryName string
	Length      int
	BBANPattern string
}

type ValidationResult struct {
	Valid          bool
	CountryCode    string
	CountryName    string
	BBAN           string
	FormatValid    bool
	ChecksumValid  bool
}

var countryRules = map[string]CountryRule{
	"AL": {CountryCode: "AL", CountryName: "Albania", Length: 28, BBANPattern: "^[0-9]{8}[A-Z0-9]{16}$"},
	"AD": {CountryCode: "AD", CountryName: "Andorra", Length: 24, BBANPattern: "^[0-9]{8}[A-Z0-9]{12}$"},
	"AT": {CountryCode: "AT", CountryName: "Austria", Length: 20, BBANPattern: "^[0-9]{16}$"},
	"BE": {CountryCode: "BE", CountryName: "Belgium", Length: 16, BBANPattern: "^[0-9]{12}$"},
	"BA": {CountryCode: "BA", CountryName: "Bosnia and Herzegovina", Length: 20, BBANPattern: "^[0-9]{16}$"},
	"BG": {CountryCode: "BG", CountryName: "Bulgaria", Length: 22, BBANPattern: "^[A-Z]{4}[0-9]{14}$"},
	"HR": {CountryCode: "HR", CountryName: "Croatia", Length: 21, BBANPattern: "^[0-9]{17}$"},
	"CY": {CountryCode: "CY", CountryName: "Cyprus", Length: 28, BBANPattern: "^[0-9]{8}[A-Z0-9]{16}$"},
	"CZ": {CountryCode: "CZ", CountryName: "Czech Republic", Length: 24, BBANPattern: "^[0-9]{20}$"},
	"DK": {CountryCode: "DK", CountryName: "Denmark", Length: 18, BBANPattern: "^[0-9]{14}$"},
	"EE": {CountryCode: "EE", CountryName: "Estonia", Length: 20, BBANPattern: "^[0-9]{16}$"},
	"FI": {CountryCode: "FI", CountryName: "Finland", Length: 18, BBANPattern: "^[0-9]{14}$"},
	"FR": {CountryCode: "FR", CountryName: "France", Length: 27, BBANPattern: "^[0-9]{10}[A-Z0-9]{11}[0-9]{2}$"},
	"DE": {CountryCode: "DE", CountryName: "Germany", Length: 22, BBANPattern: "^[0-9]{18}$"},
	"GR": {CountryCode: "GR", CountryName: "Greece", Length: 27, BBANPattern: "^[0-9]{7}[A-Z0-9]{16}$"},
	"HU": {CountryCode: "HU", CountryName: "Hungary", Length: 28, BBANPattern: "^[0-9]{24}$"},
	"IS": {CountryCode: "IS", CountryName: "Iceland", Length: 26, BBANPattern: "^[0-9]{22}$"},
	"IE": {CountryCode: "IE", CountryName: "Ireland", Length: 22, BBANPattern: "^[A-Z0-9]{18}$"},
	"IT": {CountryCode: "IT", CountryName: "Italy", Length: 27, BBANPattern: "^[A-Z]{1}[0-9]{10}[A-Z0-9]{12}$"},
	"LV": {CountryCode: "LV", CountryName: "Latvia", Length: 21, BBANPattern: "^[A-Z]{4}[A-Z0-9]{13}$"},
	"LI": {CountryCode: "LI", CountryName: "Liechtenstein", Length: 21, BBANPattern: "^[0-9]{5}[A-Z0-9]{12}$"},
	"LT": {CountryCode: "LT", CountryName: "Lithuania", Length: 20, BBANPattern: "^[0-9]{16}$"},
	"LU": {CountryCode: "LU", CountryName: "Luxembourg", Length: 20, BBANPattern: "^[0-9]{3}[A-Z0-9]{13}$"},
	"MT": {CountryCode: "MT", CountryName: "Malta", Length: 31, BBANPattern: "^[A-Z]{4}[0-9]{5}[A-Z0-9]{18}$"},
	"MC": {CountryCode: "MC", CountryName: "Monaco", Length: 27, BBANPattern: "^[0-9]{10}[A-Z0-9]{11}[0-9]{2}$"},
	"ME": {CountryCode: "ME", CountryName: "Montenegro", Length: 22, BBANPattern: "^[0-9]{18}$"},
	"NL": {CountryCode: "NL", CountryName: "Netherlands", Length: 18, BBANPattern: "^[A-Z]{4}[0-9]{10}$"},
	"NO": {CountryCode: "NO", CountryName: "Norway", Length: 15, BBANPattern: "^[0-9]{11}$"},
	"PL": {CountryCode: "PL", CountryName: "Poland", Length: 28, BBANPattern: "^[0-9]{24}$"},
	"PT": {CountryCode: "PT", CountryName: "Portugal", Length: 25, BBANPattern: "^[0-9]{21}$"},
	"RO": {CountryCode: "RO", CountryName: "Romania", Length: 24, BBANPattern: "^[A-Z]{4}[A-Z0-9]{16}$"},
	"SM": {CountryCode: "SM", CountryName: "San Marino", Length: 27, BBANPattern: "^[A-Z]{1}[0-9]{10}[A-Z0-9]{12}$"},
	"RS": {CountryCode: "RS", CountryName: "Serbia", Length: 22, BBANPattern: "^[0-9]{18}$"},
	"SK": {CountryCode: "SK", CountryName: "Slovakia", Length: 24, BBANPattern: "^[0-9]{20}$"},
	"SI": {CountryCode: "SI", CountryName: "Slovenia", Length: 19, BBANPattern: "^[0-9]{15}$"},
	"ES": {CountryCode: "ES", CountryName: "Spain", Length: 24, BBANPattern: "^[0-9]{20}$"},
	"SE": {CountryCode: "SE", CountryName: "Sweden", Length: 24, BBANPattern: "^[0-9]{20}$"},
	"CH": {CountryCode: "CH", CountryName: "Switzerland", Length: 21, BBANPattern: "^[0-9]{5}[A-Z0-9]{12}$"},
	"GB": {CountryCode: "GB", CountryName: "United Kingdom", Length: 22, BBANPattern: "^[A-Z]{4}[0-9]{14}$"},
}

func SanitizeInput(input string) string {
	return strings.ReplaceAll(input, " ", "")
}

func IsASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func Validate(iban string) *ValidationResult {
	if !IsASCII(iban) {
		return &ValidationResult{Valid: false}
	}

	sanitized := strings.ToUpper(SanitizeInput(iban))

	if len(sanitized) < 4 {
		return &ValidationResult{Valid: false}
	}

	countryCode := sanitized[:2]
	for _, r := range countryCode {
		if !unicode.IsLetter(r) {
			return &ValidationResult{Valid: false}
		}
	}

	rule, exists := countryRules[countryCode]
	formatValid := false
	checksumValid := false
	var bban string
	var countryName string

	if exists {
		countryName = rule.CountryName
		if len(sanitized) == rule.Length {
			bban = sanitized[4:]
			if MatchPattern(bban, rule.BBANPattern) {
				formatValid = true
			}
		}
	}

	if formatValid {
		checksumValid = ValidateChecksum(sanitized)
	}

	return &ValidationResult{
		Valid:         formatValid && checksumValid,
		CountryCode:   countryCode,
		CountryName:   countryName,
		BBAN:          bban,
		FormatValid:   formatValid,
		ChecksumValid: checksumValid,
	}
}

func ValidateChecksum(iban string) bool {
	if len(iban) < 4 {
		return false
	}

	rearranged := iban[4:] + iban[:4]

	var numericStr strings.Builder
	for _, r := range rearranged {
		if unicode.IsLetter(r) {
			numericStr.WriteString(fmt.Sprintf("%d", int(r-'A')+10))
		} else {
			numericStr.WriteRune(r)
		}
	}

	numeric := numericStr.String()
	if len(numeric) == 0 {
		return false
	}

	bigNum := new(big.Int)
	_, ok := bigNum.SetString(numeric, 10)
	if !ok {
		return false
	}

	modulus := big.NewInt(97)
	result := new(big.Int).Mod(bigNum, modulus)

	return result.Cmp(big.NewInt(1)) == 0
}

func MatchPattern(input, pattern string) bool {
	if !strings.HasPrefix(pattern, "^") {
		pattern = "^" + pattern
	}
	if !strings.HasSuffix(pattern, "$") {
		pattern = pattern + "$"
	}

	length := len(input)
	patternLen := len(pattern)
	i := 0
	j := 0

	for i < length && j < patternLen {
		if pattern[j] == '^' {
			j++
			continue
		}
		if pattern[j] == '$' {
			return i == length
		}

		if j+1 < patternLen && pattern[j+1] == '{' {
			endBrace := strings.Index(pattern[j:], "}")
			if endBrace == -1 {
				return false
			}

			var charClass func(rune) bool
			if pattern[j] == '[' {
				endBracket := strings.Index(pattern[j:], "]")
				if endBracket == -1 {
					return false
				}
				charSet := pattern[j+1 : j+endBracket]
				charClass = func(c rune) bool {
					if strings.Contains(charSet, "A-Z") && c >= 'A' && c <= 'Z' {
						return true
					}
					if strings.Contains(charSet, "0-9") && c >= '0' && c <= '9' {
						return true
					}
					return false
				}
				j = j + endBracket + 1
			} else {
				char := pattern[j]
				charClass = func(c rune) bool {
					return c == rune(char)
				}
				j++
			}

			minCount := 0
			maxCount := 0
			countSpec := pattern[j+1 : j+endBrace-j]
			fmt.Sscanf(countSpec, "%d,%d", &minCount, &maxCount)
			j = j + endBrace + 1

			count := 0
			for i < length && count < maxCount {
				if charClass(rune(input[i])) {
					count++
					i++
				} else {
					break
				}
			}

			if count < minCount {
				return false
			}
		} else {
			if pattern[j] == '[' {
				endBracket := strings.Index(pattern[j:], "]")
				if endBracket == -1 {
					return false
				}
				charSet := pattern[j+1 : j+endBracket]
				charClass := func(c rune) bool {
					if strings.Contains(charSet, "A-Z") && c >= 'A' && c <= 'Z' {
						return true
					}
					if strings.Contains(charSet, "0-9") && c >= '0' && c <= '9' {
						return true
					}
					return false
				}
				if i < length && charClass(rune(input[i])) {
					i++
				} else {
					return false
				}
				j = j + endBracket + 1
			} else {
				if i >= length || rune(input[i]) != rune(pattern[j]) {
					return false
				}
				i++
				j++
			}
		}
	}

	for j < patternLen {
		if pattern[j] == '$' {
			return i == length
		}
		j++
	}

	return i == length
}

func Format(iban string) string {
	sanitized := strings.ToUpper(SanitizeInput(iban))

	var builder strings.Builder
	for i, r := range sanitized {
		if i > 0 && i%4 == 0 {
			builder.WriteRune(' ')
		}
		builder.WriteRune(r)
	}

	return builder.String()
}

func GetCountryRule(countryCode string) (CountryRule, bool) {
	rule, exists := countryRules[countryCode]
	return rule, exists
}

func GetAllCountries() []CountryRule {
	rules := make([]CountryRule, 0, len(countryRules))
	for _, rule := range countryRules {
		rules = append(rules, rule)
	}
	return rules
}
