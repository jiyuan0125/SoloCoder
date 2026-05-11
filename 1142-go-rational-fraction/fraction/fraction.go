package fraction

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Fraction struct {
	Numerator   int64
	Denominator int64
}

func (f *Fraction) normalize() {
	if f.Denominator < 0 {
		f.Numerator = -f.Numerator
		f.Denominator = -f.Denominator
	}
	
	if f.Numerator == 0 {
		f.Denominator = 1
		return
	}
	
	g := gcd(abs(f.Numerator), abs(f.Denominator))
	f.Numerator /= g
	f.Denominator /= g
}

func gcd(a, b int64) int64 {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func lcm(a, b int64) int64 {
	return a / gcd(a, b) * b
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func New(numerator, denominator int64) (*Fraction, error) {
	if denominator == 0 {
		return nil, errors.New("denominator cannot be zero")
	}
	
	f := &Fraction{
		Numerator:   numerator,
		Denominator: denominator,
	}
	f.normalize()
	return f, nil
}

func Parse(s string) (*Fraction, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty input")
	}
	
	if strings.Contains(s, "/") {
		return parseFraction(s)
	}
	
	if strings.Contains(s, ".") {
		return parseDecimal(s)
	}
	
	return parseInteger(s)
}

func parseFraction(s string) (*Fraction, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid fraction format: %s", s)
	}
	
	numStr := strings.TrimSpace(parts[0])
	denStr := strings.TrimSpace(parts[1])
	
	if numStr == "" || denStr == "" {
		return nil, fmt.Errorf("invalid fraction format: %s", s)
	}
	
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid numerator: %s", numStr)
	}
	
	den, err := strconv.ParseInt(denStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid denominator: %s", denStr)
	}
	
	if den == 0 {
		return nil, errors.New("denominator cannot be zero")
	}
	
	return New(num, den)
}

func parseDecimal(s string) (*Fraction, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid decimal format: %s", s)
	}
	
	intPart := strings.TrimSpace(parts[0])
	decPart := strings.TrimSpace(parts[1])
	
	if intPart == "" && decPart == "" {
		return nil, fmt.Errorf("invalid decimal format: %s", s)
	}
	
	if decPart == "" {
		return parseInteger(intPart)
	}
	
	var negative bool
	if strings.HasPrefix(intPart, "-") {
		negative = true
		intPart = strings.TrimPrefix(intPart, "-")
	} else if strings.HasPrefix(intPart, "+") {
		intPart = strings.TrimPrefix(intPart, "+")
	}
	
	decPart = strings.TrimRight(decPart, "0")
	if decPart == "" {
		return parseInteger(fmt.Sprintf("%s%s", map[bool]string{true: "-", false: ""}[negative], intPart))
	}
	
	numStr := intPart + decPart
	denomPower := len(decPart)
	
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid decimal number: %s", s)
	}
	
	denom := int64(1)
	for i := 0; i < denomPower; i++ {
		denom *= 10
	}
	
	if negative {
		num = -num
	}
	
	return New(num, denom)
}

func parseInteger(s string) (*Fraction, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, errors.New("empty integer")
	}
	
	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid integer: %s", s)
	}
	
	return New(num, 1)
}

func (f *Fraction) String() string {
	if f.Numerator == 0 {
		return "0"
	}
	if f.Denominator == 1 {
		return strconv.FormatInt(f.Numerator, 10)
	}
	return fmt.Sprintf("%d/%d", f.Numerator, f.Denominator)
}

func (f *Fraction) MixedString() (string, bool) {
	if f.Numerator == 0 {
		return "0", true
	}
	
	if f.Denominator == 1 {
		return strconv.FormatInt(f.Numerator, 10), true
	}
	
	absNum := abs(f.Numerator)
	if absNum < f.Denominator {
		return "", false
	}
	
	integerPart := absNum / f.Denominator
	remainder := absNum % f.Denominator
	
	var result string
	if f.Numerator < 0 {
		result = fmt.Sprintf("-%d又%d/%d", integerPart, remainder, f.Denominator)
	} else {
		result = fmt.Sprintf("%d又%d/%d", integerPart, remainder, f.Denominator)
	}
	
	return result, true
}

func (f *Fraction) IsInteger() bool {
	return f.Denominator == 1
}

func (f *Fraction) ToDecimal(precision int) (string, error) {
	if precision < 0 {
		return "", errors.New("precision cannot be negative")
	}
	
	if f.Numerator == 0 {
		return "0", nil
	}
	
	sign := ""
	num := f.Numerator
	if num < 0 {
		sign = "-"
		num = -num
	}
	
	integerPart := num / f.Denominator
	remainder := num % f.Denominator
	
	if remainder == 0 {
		return fmt.Sprintf("%s%d", sign, integerPart), nil
	}
	
	if precision == 0 {
		return fmt.Sprintf("%s%d", sign, integerPart), nil
	}
	
	decimalDigits := make([]byte, 0, precision)
	for i := 0; i < precision; i++ {
		remainder *= 10
		digit := remainder / f.Denominator
		decimalDigits = append(decimalDigits, byte('0'+digit))
		remainder = remainder % f.Denominator
	}
	
	return fmt.Sprintf("%s%d.%s", sign, integerPart, string(decimalDigits)), nil
}

func (f *Fraction) Add(other *Fraction) (*Fraction, error) {
	if other == nil {
		return nil, errors.New("other fraction is nil")
	}
	
	commonDenom := lcm(f.Denominator, other.Denominator)
	num1 := f.Numerator * (commonDenom / f.Denominator)
	num2 := other.Numerator * (commonDenom / other.Denominator)
	
	return New(num1+num2, commonDenom)
}

func (f *Fraction) Sub(other *Fraction) (*Fraction, error) {
	if other == nil {
		return nil, errors.New("other fraction is nil")
	}
	
	commonDenom := lcm(f.Denominator, other.Denominator)
	num1 := f.Numerator * (commonDenom / f.Denominator)
	num2 := other.Numerator * (commonDenom / other.Denominator)
	
	return New(num1-num2, commonDenom)
}

func (f *Fraction) Mul(other *Fraction) (*Fraction, error) {
	if other == nil {
		return nil, errors.New("other fraction is nil")
	}
	
	return New(f.Numerator*other.Numerator, f.Denominator*other.Denominator)
}

func (f *Fraction) Div(other *Fraction) (*Fraction, error) {
	if other == nil {
		return nil, errors.New("other fraction is nil")
	}
	
	if other.Numerator == 0 {
		return nil, errors.New("division by zero")
	}
	
	return New(f.Numerator*other.Denominator, f.Denominator*other.Numerator)
}

func Add(f1, f2 *Fraction) (*Fraction, error) {
	if f1 == nil || f2 == nil {
		return nil, errors.New("fraction cannot be nil")
	}
	return f1.Add(f2)
}

func Sub(f1, f2 *Fraction) (*Fraction, error) {
	if f1 == nil || f2 == nil {
		return nil, errors.New("fraction cannot be nil")
	}
	return f1.Sub(f2)
}

func Mul(f1, f2 *Fraction) (*Fraction, error) {
	if f1 == nil || f2 == nil {
		return nil, errors.New("fraction cannot be nil")
	}
	return f1.Mul(f2)
}

func Div(f1, f2 *Fraction) (*Fraction, error) {
	if f1 == nil || f2 == nil {
		return nil, errors.New("fraction cannot be nil")
	}
	return f1.Div(f2)
}

var (
	ErrOverflow = errors.New("integer overflow")
)

func SafeAdd(a, b int64) (int64, error) {
	if a > 0 && b > math.MaxInt64-a {
		return 0, ErrOverflow
	}
	if a < 0 && b < math.MinInt64-a {
		return 0, ErrOverflow
	}
	return a + b, nil
}

func SafeMul(a, b int64) (int64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a == math.MinInt64 || b == math.MinInt64 {
		return 0, ErrOverflow
	}
	if a > math.MaxInt64/b || a < math.MinInt64/b {
		return 0, ErrOverflow
	}
	return a * b, nil
}

func SafeSub(a, b int64) (int64, error) {
	if b == 0 {
		return a, nil
	}
	if b == math.MinInt64 {
		if a > 0 {
			return 0, ErrOverflow
		}
	} else {
		if a > math.MaxInt64+b || a < math.MinInt64+b {
			return 0, ErrOverflow
		}
	}
	return a - b, nil
}
