package license

import (
	"unicode"
)

const (
	minLenRegular   = 7
	maxLenRegular   = 7
	minLenNewEnergy = 8
	maxLenNewEnergy = 8
	minLenMilitary  = 7
	maxLenMilitary  = 8
	minLenEmbassy   = 7
	maxLenEmbassy   = 7
	minLenArmedPolice = 7
	maxLenArmedPolice = 8
)

func isValidCityLetter(c rune) bool {
	if c < 'A' || c > 'Z' {
		return false
	}
	return c != 'I' && c != 'O'
}

func isValidSuffixLetter(c rune) bool {
	if c < 'A' || c > 'Z' {
		return false
	}
	return c != 'I' && c != 'O'
}

func isAlphanumeric(c rune) bool {
	return unicode.IsUpper(c) || unicode.IsDigit(c)
}

func normalizeInput(input string) string {
	result := ""
	for _, r := range input {
		if !unicode.IsSpace(r) {
			result += string(r)
		}
	}
	return result
}

func countLettersAndDigits(s string) (int, int) {
	letters := 0
	digits := 0
	for _, r := range s {
		if unicode.IsUpper(r) {
			letters++
		} else if unicode.IsDigit(r) {
			digits++
		}
	}
	return letters, digits
}

func allDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func allLetters(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

func validateRegularPlate(chars []rune) (bool, string) {
	if len(chars) != minLenRegular {
		return false, "长度错误，应为7位"
	}

	province := chars[0]
	if !provinces[province] {
		return false, "省份简称错误"
	}

	city := chars[1]
	if !isValidCityLetter(city) {
		return false, "地市字母非法"
	}

	suffix := string(chars[2:7])
	letters, digits := countLettersAndDigits(suffix)

	if letters == 0 || digits == 0 {
		return false, "后缀格式不对，需要字母数字混合"
	}

	lastChar := chars[6]
	if unicode.IsUpper(lastChar) && !isValidSuffixLetter(lastChar) {
		return false, "后缀格式不对，最后一位字母非法"
	}

	for _, c := range suffix {
		if !isAlphanumeric(c) {
			return false, "后缀格式不对"
		}
		if unicode.IsUpper(c) && !isValidSuffixLetter(c) {
			return false, "后缀格式不对"
		}
	}

	return true, ""
}

func validateNewEnergyPlate(chars []rune) (bool, string, PlateType) {
	if len(chars) != minLenNewEnergy {
		return false, "长度错误，应为8位", PlateTypeInvalid
	}

	province := chars[0]
	if !provinces[province] {
		return false, "省份简称错误", PlateTypeInvalid
	}

	city := chars[1]
	if !isValidCityLetter(city) {
		return false, "地市字母非法", PlateTypeInvalid
	}

	suffix := string(chars[2:8])
	for _, c := range suffix {
		if !isAlphanumeric(c) {
			return false, "后缀格式不对", PlateTypeInvalid
		}
		if unicode.IsUpper(c) && !isValidSuffixLetter(c) {
			return false, "后缀格式不对", PlateTypeInvalid
		}
	}

	firstSuffix := chars[2]
	lastSuffix := chars[7]

	var plateType PlateType
	if firstSuffix == 'D' {
		plateType = PlateTypeNewEnergyPure
	} else if firstSuffix == 'F' {
		plateType = PlateTypeNewEnergyHybrid
	} else if lastSuffix == 'D' {
		plateType = PlateTypeNewEnergyPure
	} else if lastSuffix == 'F' {
		plateType = PlateTypeNewEnergyHybrid
	} else {
		return false, "后缀格式不对，新能源车牌需要以D或F开头或结尾", PlateTypeInvalid
	}

	return true, "", plateType
}

func validateMilitaryPlate(chars []rune) (bool, string) {
	if len(chars) < minLenMilitary || len(chars) > maxLenMilitary {
		return false, "长度错误，应为7或8位"
	}

	firstChar := chars[0]
	isValidPrefix := false

	if firstChar == '第' {
		isValidPrefix = true
	} else {
		validPrefixes := map[rune]bool{
			'军': true, '海': true, '空': true, '北': true, '沈': true,
			'兰': true, '济': true, '南': true, '广': true, '成': true,
		}
		if validPrefixes[firstChar] {
			isValidPrefix = true
		}
	}

	if !isValidPrefix {
		return false, "军车前缀格式错误"
	}

	secondChar := chars[1]
	if !unicode.IsUpper(secondChar) {
		return false, "军车字母格式错误"
	}

	for i := 2; i < len(chars); i++ {
		c := chars[i]
		if !unicode.IsUpper(c) && !unicode.IsDigit(c) {
			return false, "军车后缀格式错误"
		}
	}

	return true, ""
}

func validateEmbassyPlate(chars []rune) (bool, string) {
	if len(chars) != minLenEmbassy {
		return false, "长度错误，应为7位"
	}

	firstChar := chars[0]
	if firstChar != '使' && firstChar != '领' {
		return false, "使领馆车牌前缀错误"
	}

	suffix := string(chars[1:7])
	if !allDigits(suffix) {
		return false, "使领馆车牌后缀应为6位数字"
	}

	return true, ""
}

func validateArmedPolicePlate(chars []rune) (bool, string) {
	if len(chars) < minLenArmedPolice || len(chars) > maxLenArmedPolice {
		return false, "长度错误，应为7或8位"
	}

	firstChar := chars[0]
	secondChar := chars[1]

	if firstChar == 'W' && secondChar == 'J' {
		for i := 2; i < len(chars); i++ {
			c := chars[i]
			if !unicode.IsUpper(c) && !unicode.IsDigit(c) {
				return false, "武警车牌格式错误"
			}
		}
		return true, ""
	}

	return false, "武警车牌前缀应为WJ"
}
