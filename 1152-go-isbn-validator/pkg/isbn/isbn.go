package isbn

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ISBNType string

const (
	ISBN10 ISBNType = "ISBN-10"
	ISBN13 ISBNType = "ISBN-13"
)

type ValidationResult struct {
	IsValid       bool
	ISBNType      ISBNType
	CheckDigit    string
	CalculatedCD  string
	Reason        string
}

func CleanInput(input string) string {
	input = strings.ToUpper(strings.TrimSpace(input))
	cleaned := make([]rune, 0, len(input))
	for _, r := range input {
		if (r >= '0' && r <= '9') || r == 'X' {
			cleaned = append(cleaned, r)
		}
	}
	return string(cleaned)
}

func DetectType(cleaned string) (ISBNType, error) {
	switch len(cleaned) {
	case 10:
		return ISBN10, nil
	case 13:
		return ISBN13, nil
	default:
		return "", fmt.Errorf("invalid length: %d", len(cleaned))
	}
}

func CalculateCheckDigit10(first9 string) (string, error) {
	if len(first9) != 9 {
		return "", fmt.Errorf("expected 9 digits, got %d", len(first9))
	}
	
	var sum int
	for i := 0; i < 9; i++ {
		digit, err := strconv.Atoi(string(first9[i]))
		if err != nil {
			return "", fmt.Errorf("invalid digit at position %d", i+1)
		}
		sum += digit * (10 - i)
	}
	
	remainder := sum % 11
	cd := (11 - remainder) % 11
	
	if cd == 10 {
		return "X", nil
	}
	return strconv.Itoa(cd), nil
}

func CalculateCheckDigit13(first12 string) (string, error) {
	if len(first12) != 12 {
		return "", fmt.Errorf("expected 12 digits, got %d", len(first12))
	}
	
	var sum int
	for i := 0; i < 12; i++ {
		digit, err := strconv.Atoi(string(first12[i]))
		if err != nil {
			return "", fmt.Errorf("invalid digit at position %d", i+1)
		}
		weight := 1
		if i%2 == 1 {
			weight = 3
		}
		sum += digit * weight
	}
	
	remainder := sum % 10
	cd := (10 - remainder) % 10
	return strconv.Itoa(cd), nil
}

func Validate(input string) ValidationResult {
	cleaned := CleanInput(input)
	
	if len(cleaned) == 0 {
		return ValidationResult{
			IsValid: false,
			Reason:  "empty ISBN",
		}
	}
	
	isbnType, err := DetectType(cleaned)
	if err != nil {
		return ValidationResult{
			IsValid: false,
			Reason:  fmt.Sprintf("invalid ISBN length: %d characters (must be 10 or 13)", len(cleaned)),
		}
	}
	
	if isbnType == ISBN13 {
		if !strings.HasPrefix(cleaned, "978") && !strings.HasPrefix(cleaned, "979") {
			return ValidationResult{
				IsValid:  false,
				ISBNType: ISBN13,
				Reason:   "ISBN-13 must start with 978 or 979",
			}
		}
		
		for i := 0; i < 13; i++ {
			if !strings.Contains("0123456789", string(cleaned[i])) {
				return ValidationResult{
					IsValid:  false,
					ISBNType: ISBN13,
					Reason:   fmt.Sprintf("invalid character at position %d: %c (ISBN-13 can only contain digits)", i+1, cleaned[i]),
				}
			}
		}
		
		first12 := cleaned[:12]
		providedCD := string(cleaned[12])
		
		calculatedCD, err := CalculateCheckDigit13(first12)
		if err != nil {
			return ValidationResult{
				IsValid:  false,
				ISBNType: ISBN13,
				Reason:   err.Error(),
			}
		}
		
		return ValidationResult{
			IsValid:      providedCD == calculatedCD,
			ISBNType:     ISBN13,
			CheckDigit:   providedCD,
			CalculatedCD: calculatedCD,
			Reason:       func() string {
				if providedCD == calculatedCD {
					return "valid"
				}
				return fmt.Sprintf("check digit mismatch: provided %s, calculated %s", providedCD, calculatedCD)
			}(),
		}
	}
	
	if isbnType == ISBN10 {
		for i := 0; i < 9; i++ {
			if !strings.Contains("0123456789", string(cleaned[i])) {
				return ValidationResult{
					IsValid:  false,
					ISBNType: ISBN10,
					Reason:   fmt.Sprintf("invalid character at position %d: %c (first 9 positions can only contain digits)", i+1, cleaned[i]),
				}
			}
		}
		
		lastChar := cleaned[9]
		if !strings.Contains("0123456789X", string(lastChar)) {
			return ValidationResult{
				IsValid:  false,
				ISBNType: ISBN10,
				Reason:   fmt.Sprintf("invalid character at position 10: %c (can only be digit or X)", lastChar),
			}
		}
		
		first9 := cleaned[:9]
		providedCD := string(cleaned[9])
		
		calculatedCD, err := CalculateCheckDigit10(first9)
		if err != nil {
			return ValidationResult{
				IsValid:  false,
				ISBNType: ISBN10,
				Reason:   err.Error(),
			}
		}
		
		return ValidationResult{
			IsValid:      providedCD == calculatedCD,
			ISBNType:     ISBN10,
			CheckDigit:   providedCD,
			CalculatedCD: calculatedCD,
			Reason:       func() string {
				if providedCD == calculatedCD {
					return "valid"
				}
				return fmt.Sprintf("check digit mismatch: provided %s, calculated %s", providedCD, calculatedCD)
			}(),
		}
	}
	
	return ValidationResult{
		IsValid: false,
		Reason:  "unknown error",
	}
}

func Convert10To13(isbn10 string) (string, error) {
	cleaned := CleanInput(isbn10)
	if len(cleaned) != 10 {
		return "", fmt.Errorf("not a valid ISBN-10")
	}
	
	first9 := cleaned[:9]
	base13 := "978" + first9
	
	cd, err := CalculateCheckDigit13(base13)
	if err != nil {
		return "", err
	}
	
	return base13 + cd, nil
}

func Convert13To10(isbn13 string) (string, error) {
	cleaned := CleanInput(isbn13)
	if len(cleaned) != 13 {
		return "", fmt.Errorf("not a valid ISBN-13")
	}
	
	if !strings.HasPrefix(cleaned, "978") {
		return "", fmt.Errorf("ISBN-13 must start with 978 to convert to ISBN-10")
	}
	
	first9 := cleaned[3:12]
	
	cd, err := CalculateCheckDigit10(first9)
	if err != nil {
		return "", err
	}
	
	return first9 + cd, nil
}

func Format(input string) (string, error) {
	cleaned := CleanInput(input)
	validation := Validate(cleaned)
	
	if !validation.IsValid {
		return "", fmt.Errorf("cannot format invalid ISBN: %s", validation.Reason)
	}
	
	if validation.ISBNType == ISBN10 {
		return formatISBN10(cleaned)
	}
	
	return formatISBN13(cleaned)
}

func formatISBN10(cleaned string) (string, error) {
	if len(cleaned) != 10 {
		return "", fmt.Errorf("not a valid ISBN-10")
	}
	
	formatted, err := formatBasic(cleaned, 10)
	if err != nil {
		return "", err
	}
	
	return formatted, nil
}

func formatISBN13(cleaned string) (string, error) {
	if len(cleaned) != 13 {
		return "", fmt.Errorf("not a valid ISBN-13")
	}
	
	formatted, err := formatBasic(cleaned, 13)
	if err != nil {
		return "", err
	}
	
	return formatted, nil
}

func formatBasic(cleaned string, length int) (string, error) {
	re := regexp.MustCompile(`[^\dX]`)
	normalized := re.ReplaceAllString(cleaned, "")
	
	if len(normalized) != length {
		return "", fmt.Errorf("expected %d characters, got %d", length, len(normalized))
	}
	
	return normalized, nil
}

func FormatWithHyphens(input string) (string, error) {
	cleaned := CleanInput(input)
	validation := Validate(cleaned)
	
	if !validation.IsValid {
		return "", fmt.Errorf("cannot format invalid ISBN: %s", validation.Reason)
	}
	
	if validation.ISBNType == ISBN10 {
		return formatISBN10WithHyphens(cleaned)
	}
	
	return formatISBN13WithHyphens(cleaned)
}

func formatISBN10WithHyphens(cleaned string) (string, error) {
	if len(cleaned) != 10 {
		return "", fmt.Errorf("not a valid ISBN-10")
	}
	
	group := getGroup(cleaned)
	if group == "" {
		return cleaned, nil
	}
	
	return insertHyphensISBN10(cleaned, group)
}

func formatISBN13WithHyphens(cleaned string) (string, error) {
	if len(cleaned) != 13 {
		return "", fmt.Errorf("not a valid ISBN-13")
	}
	
	prefix := cleaned[:3]
	rest := cleaned[3:12]
	cd := cleaned[12:]
	
	group := getGroup(rest)
	if group == "" {
		return fmt.Sprintf("%s-%s-%s-%s", prefix, rest[:1], rest[1:], cd), nil
	}
	
	base, err := insertHyphensISBN10(rest+cd, group)
	if err != nil {
		return "", err
	}
	
	return prefix + "-" + base, nil
}

func getGroup(isbn string) string {
	isbn = strings.ToUpper(isbn)
	
	groups := []struct {
		prefix string
		length int
	}{
		{"0", 1},
		{"1", 1},
		{"2", 1},
		{"3", 1},
		{"4", 1},
		{"5", 1},
		{"6", 1},
		{"7", 1},
		{"80", 2},
		{"81", 2},
		{"82", 2},
		{"83", 2},
		{"84", 2},
		{"85", 2},
		{"86", 2},
		{"87", 2},
		{"88", 2},
		{"89", 2},
		{"90", 2},
		{"91", 2},
		{"92", 2},
		{"93", 2},
		{"94", 2},
		{"95", 2},
		{"96", 2},
		{"97", 2},
		{"98", 2},
		{"99", 2},
	}
	
	for _, g := range groups {
		if strings.HasPrefix(isbn, g.prefix) {
			return g.prefix
		}
	}
	
	return ""
}

func insertHyphensISBN10(isbn, group string) (string, error) {
	if len(isbn) != 10 {
		return "", fmt.Errorf("not a valid ISBN-10")
	}
	
	groupLen := len(group)
	remainder := isbn[groupLen : len(isbn)-1]
	
	publisherLen := 2
	if len(remainder) >= 6 {
		publisherLen = 2
	} else if len(remainder) >= 5 {
		publisherLen = len(remainder) - 3
	} else {
		publisherLen = 1
	}
	
	if publisherLen > len(remainder) {
		publisherLen = len(remainder) - 1
	}
	if publisherLen < 1 {
		publisherLen = 1
	}
	
	cd := isbn[len(isbn)-1:]
	
	return fmt.Sprintf("%s-%s-%s-%s",
		isbn[:groupLen],
		isbn[groupLen:groupLen+publisherLen],
		isbn[groupLen+publisherLen:len(isbn)-1],
		cd,
	), nil
}
