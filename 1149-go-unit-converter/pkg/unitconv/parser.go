package unitconv

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type ParsedUnit struct {
	Original             string
	SimpleUnit           string
	Numerator            []string
	NumeratorExponents   []int
	Denominator          []string
	DenominatorExponents []int
}

func (p *ParsedUnit) IsCompound() bool {
	return len(p.Numerator) > 1 || len(p.Denominator) > 0
}

func ParseUnit(unitStr string) (*ParsedUnit, error) {
	unitStr = strings.TrimSpace(unitStr)

	if _, ok := GetUnitInfo(unitStr); ok {
		return &ParsedUnit{
			Original:           unitStr,
			SimpleUnit:         unitStr,
			Numerator:          []string{unitStr},
			NumeratorExponents: []int{1},
		}, nil
	}

	parts := strings.SplitN(unitStr, "/", 2)
	numeratorPart := parts[0]
	denominatorPart := ""
	if len(parts) == 2 {
		denominatorPart = parts[1]
	}

	numerator, numExps, err := parseUnitSegment(numeratorPart)
	if err != nil {
		return nil, err
	}

	denominator := []string{}
	denExps := []int{}
	if denominatorPart != "" {
		denominator, denExps, err = parseUnitSegment(denominatorPart)
		if err != nil {
			return nil, err
		}
	}

	simpleUnit := ""
	if len(numerator) == 1 && len(denominator) == 0 {
		simpleUnit = numerator[0]
	}

	return &ParsedUnit{
		Original:             unitStr,
		SimpleUnit:           simpleUnit,
		Numerator:            numerator,
		NumeratorExponents:   numExps,
		Denominator:          denominator,
		DenominatorExponents: denExps,
	}, nil
}

func parseUnitSegment(segment string) ([]string, []int, error) {
	segment = strings.ReplaceAll(segment, " ", "")
	segment = normalizeExponents(segment)

	result := []string{}
	exponents := []int{}

	i := 0
	for i < len(segment) {
		if segment[i] == '*' {
			i++
			continue
		}

		unitStart := i
		for i < len(segment) && segment[i] != '*' && segment[i] != '^' {
			i++
		}
		unitPart := segment[unitStart:i]

		if len(unitPart) == 0 {
			return nil, nil, fmt.Errorf("invalid unit format: %s", segment)
		}

		exponent := 1
		if i < len(segment) && segment[i] == '^' {
			i++
			expStart := i
			for i < len(segment) && (unicode.IsDigit(rune(segment[i])) || segment[i] == '-') {
				i++
			}
			if expStart == i {
				return nil, nil, fmt.Errorf("invalid exponent in: %s", segment)
			}
			expStr := segment[expStart:i]
			expVal, err := strconv.Atoi(expStr)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid exponent %s: %v", expStr, err)
			}
			exponent = expVal
		}

		unitPart = normalizeUnitPart(unitPart)

		if _, ok := GetUnitInfo(unitPart); !ok {
			return nil, nil, fmt.Errorf("unsupported unit: %s. %s", unitPart, GetAllSupportedUnits())
		}

		result = append(result, unitPart)
		exponents = append(exponents, exponent)
	}

	return result, exponents, nil
}

func normalizeExponents(s string) string {
	s = strings.ReplaceAll(s, "²", "^2")
	s = strings.ReplaceAll(s, "³", "^3")
	s = strings.ReplaceAll(s, "⁴", "^4")
	s = strings.ReplaceAll(s, "⁵", "^5")
	s = strings.ReplaceAll(s, "⁶", "^6")
	s = strings.ReplaceAll(s, "⁷", "^7")
	s = strings.ReplaceAll(s, "⁸", "^8")
	s = strings.ReplaceAll(s, "⁹", "^9")
	return s
}

func normalizeUnitPart(part string) string {
	part = strings.ReplaceAll(part, "2", "²")
	part = strings.ReplaceAll(part, "3", "³")
	return part
}

var validUnitPattern = regexp.MustCompile(`^[a-zA-Z0-9µµ°²³^*/\-_]+$`)

func ValidateUnit(unit string) error {
	unit = strings.TrimSpace(unit)
	if unit == "" {
		return fmt.Errorf("unit cannot be empty")
	}

	if !validUnitPattern.MatchString(unit) {
		return fmt.Errorf("invalid unit format: %s", unit)
	}

	parsed, err := ParseUnit(unit)
	if err != nil {
		return err
	}

	for _, u := range parsed.Numerator {
		if _, ok := GetUnitInfo(u); !ok {
			return fmt.Errorf("unsupported unit: %s. %s", u, GetAllSupportedUnits())
		}
	}
	for _, u := range parsed.Denominator {
		if _, ok := GetUnitInfo(u); !ok {
			return fmt.Errorf("unsupported unit: %s. %s", u, GetAllSupportedUnits())
		}
	}

	return nil
}
