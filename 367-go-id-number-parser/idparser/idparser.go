package idparser

import (
	"strconv"
	"strings"
	"time"
)

var weights = []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
var checkCodes = []string{"1", "0", "X", "9", "8", "7", "6", "5", "4", "3", "2"}

type IDInfo struct {
	OriginalID     string
	StandardizedID string
	BirthDate      time.Time
	Gender         string
	ProvinceCode   string
	ProvinceName   string
	Age            int
	IsValid        bool
}

type ParseError struct {
	Message string
}

func (e *ParseError) Error() string {
	return e.Message
}

func NewParseError(message string) *ParseError {
	return &ParseError{Message: message}
}

func CleanInput(input string) string {
	cleaned := strings.TrimSpace(input)
	cleaned = fullWidthToHalfWidth(cleaned)
	return cleaned
}

func fullWidthToHalfWidth(s string) string {
	var result strings.Builder
	for _, r := range s {
		if r >= '０' && r <= '９' {
			result.WriteRune(r - '０' + '0')
		} else if r == 'Ｘ' || r == 'ｘ' {
			result.WriteRune('X')
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func Convert15To18(id string) (string, error) {
	if len(id) != 15 {
		return "", NewParseError("身份证号长度不是15位")
	}

	for _, c := range id {
		if c < '0' || c > '9' {
			return "", NewParseError("15位身份证号只能包含数字")
		}
	}

	year := id[6:8]
	fullID := id[:6] + "19" + year + id[8:]

	checkCode, err := CalculateCheckCode(fullID)
	if err != nil {
		return "", err
	}

	return fullID + checkCode, nil
}

func CalculateCheckCode(id string) (string, error) {
	if len(id) != 17 {
		return "", NewParseError("计算校验码需要17位数字")
	}

	sum := 0
	for i := 0; i < 17; i++ {
		digit, err := strconv.Atoi(string(id[i]))
		if err != nil {
			return "", NewParseError("身份证号包含非数字字符")
		}
		sum += digit * weights[i]
	}

	mod := sum % 11
	return checkCodes[mod], nil
}

func ValidateCheckCode(id string) (bool, error) {
	if len(id) != 18 {
		return false, NewParseError("身份证号长度不是18位")
	}

	actualCheckCode := strings.ToUpper(string(id[17]))
	expectedCheckCode, err := CalculateCheckCode(id[:17])
	if err != nil {
		return false, err
	}

	return actualCheckCode == expectedCheckCode, nil
}

func ParseBirthDate(id string) (time.Time, error) {
	if len(id) != 18 {
		return time.Time{}, NewParseError("解析出生日期需要18位身份证号")
	}

	birthDateStr := id[6:14]
	year, _ := strconv.Atoi(birthDateStr[0:4])
	month, _ := strconv.Atoi(birthDateStr[4:6])
	day, _ := strconv.Atoi(birthDateStr[6:8])

	if month < 1 || month > 12 {
		return time.Time{}, NewParseError("无效的月份")
	}

	if day < 1 {
		return time.Time{}, NewParseError("无效的日期")
	}

	maxDay := getMaxDay(year, month)
	if day > maxDay {
		return time.Time{}, NewParseError("无效的日期")
	}

	loc, _ := time.LoadLocation("Local")
	birthDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, loc)
	return birthDate, nil
}

func getMaxDay(year int, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if isLeapYear(year) {
			return 29
		}
		return 28
	default:
		return 0
	}
}

func isLeapYear(year int) bool {
	if year%400 == 0 {
		return true
	}
	if year%100 == 0 {
		return false
	}
	if year%4 == 0 {
		return true
	}
	return false
}

func ParseGender(id string) (string, error) {
	if len(id) != 18 {
		return "", NewParseError("解析性别需要18位身份证号")
	}

	genderDigit, err := strconv.Atoi(string(id[16]))
	if err != nil {
		return "", NewParseError("性别位不是有效数字")
	}

	if genderDigit%2 == 1 {
		return "男", nil
	}
	return "女", nil
}

func CalculateAge(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	if now.YearDay() < birthDate.YearDay() {
		age--
	}

	return age
}

func Parse(id string) (*IDInfo, error) {
	cleanedID := CleanInput(id)

	var standardizedID string
	var originalIs15 bool

	switch len(cleanedID) {
	case 15:
		var err error
		standardizedID, err = Convert15To18(cleanedID)
		if err != nil {
			return nil, err
		}
		originalIs15 = true
	case 18:
		standardizedID = cleanedID
		originalIs15 = false
	default:
		return nil, NewParseError("身份证号长度必须是15位或18位")
	}

	isValid, err := ValidateCheckCode(standardizedID)
	if err != nil {
		return nil, err
	}
	if !isValid {
		return nil, NewParseError("身份证号校验码无效")
	}

	birthDate, err := ParseBirthDate(standardizedID)
	if err != nil {
		return nil, err
	}

	gender, err := ParseGender(standardizedID)
	if err != nil {
		return nil, err
	}

	provinceCode := standardizedID[0:2]
	provinceName := GetProvinceName(provinceCode)
	if provinceName == "" {
		return nil, NewParseError("无效的省份代码")
	}

	age := CalculateAge(birthDate)

	info := &IDInfo{
		OriginalID:     id,
		StandardizedID: standardizedID,
		BirthDate:      birthDate,
		Gender:         gender,
		ProvinceCode:   provinceCode,
		ProvinceName:   provinceName,
		Age:            age,
		IsValid:        true,
	}

	if originalIs15 {
		info.StandardizedID = standardizedID
	}

	return info, nil
}
