package utils

import (
	"regexp"
	"strings"
)

func TrimIDCard(input string) string {
	return strings.TrimSpace(input)
}

func ValidateIDCard(idCard string) error {
	trimmed := strings.TrimSpace(idCard)
	
	if len(trimmed) != 18 {
		return &IDCardError{
			Field:   "id_card",
			Message: "身份证号必须为18位",
		}
	}
	
	pattern := `^[0-9]{17}[0-9X]$`
	matched, _ := regexp.MatchString(pattern, trimmed)
	
	if !matched {
		return &IDCardError{
			Field:   "id_card",
			Message: "身份证号格式不正确：前17位必须为数字，最后一位可以是数字或X",
		}
	}
	
	return nil
}

type IDCardError struct {
	Field   string
	Message string
}

func (e *IDCardError) Error() string {
	return e.Message
}
