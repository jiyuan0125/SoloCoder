package license

import (
	"unicode"
)

func Validate(plate string) ValidationResult {
	normalized := normalizeInput(plate)
	if normalized == "" {
		return ValidationResult{
			Valid:  false,
			Type:   PlateTypeInvalid,
			Reason: "输入为空",
		}
	}

	chars := []rune(normalized)
	length := len(chars)

	if length < 7 || length > 8 {
		return ValidationResult{
			Valid:  false,
			Type:   PlateTypeInvalid,
			Reason: "长度错误，应为7或8位",
		}
	}

	firstChar := chars[0]

	if firstChar == '使' || firstChar == '领' {
		valid, reason := validateEmbassyPlate(chars)
		if valid {
			return ValidationResult{
				Valid:  true,
				Type:   PlateTypeEmbassy,
				Reason: "",
			}
		}
		return ValidationResult{
			Valid:  false,
			Type:   PlateTypeInvalid,
			Reason: reason,
		}
	}

	if firstChar == 'W' && length >= 2 && chars[1] == 'J' {
		valid, reason := validateArmedPolicePlate(chars)
		if valid {
			return ValidationResult{
				Valid:  true,
				Type:   PlateTypeArmedPolice,
				Reason: "",
			}
		}
		return ValidationResult{
			Valid:  false,
			Type:   PlateTypeInvalid,
			Reason: reason,
		}
	}

	if !provinces[firstChar] {
		militaryValid, _ := validateMilitaryPlate(chars)
		if militaryValid {
			return ValidationResult{
				Valid:  true,
				Type:   PlateTypeMilitary,
				Reason: "",
			}
		}

		return ValidationResult{
			Valid:  false,
			Type:   PlateTypeInvalid,
			Reason: "省份简称错误或车牌格式不识别",
		}
	}

	if length == 8 {
		valid, reason, plateType := validateNewEnergyPlate(chars)
		if valid {
			return ValidationResult{
				Valid:  true,
				Type:   plateType,
				Reason: "",
			}
		}
		return ValidationResult{
			Valid:  false,
			Type:   PlateTypeInvalid,
			Reason: reason,
		}
	}

	if length == 7 {
		militaryValid, _ := validateMilitaryPlate(chars)
		if militaryValid {
			return ValidationResult{
				Valid:  true,
				Type:   PlateTypeMilitary,
				Reason: "",
			}
		}

		valid, reason := validateRegularPlate(chars)
		if valid {
			return ValidationResult{
				Valid:  true,
				Type:   PlateTypeRegularSmall,
				Reason: "",
			}
		}
		return ValidationResult{
			Valid:  false,
			Type:   PlateTypeInvalid,
			Reason: reason,
		}
	}

	return ValidationResult{
		Valid:  false,
		Type:   PlateTypeInvalid,
		Reason: "未知错误",
	}
}

func IsValidProvince(r rune) bool {
	return provinces[r]
}

func IsValidCityLetter(r rune) bool {
	return isValidCityLetter(r)
}

func GetProvinces() []rune {
	result := make([]rune, 0, len(provinces))
	for p := range provinces {
		result = append(result, p)
	}
	return result
}

func Normalize(plate string) string {
	return normalizeInput(plate)
}

func ContainsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}
