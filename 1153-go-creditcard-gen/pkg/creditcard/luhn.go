package creditcard

func ComputeLuhnCheckDigit(digits []int) int {
	n := len(digits)
	sum := 0
	double := true

	for i := n - 1; i >= 0; i-- {
		d := digits[i]
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}

	checkDigit := (10 - (sum % 10)) % 10
	return checkDigit
}

func ValidateLuhn(cardNumber string) bool {
	if len(cardNumber) == 0 {
		return false
	}

	digits := make([]int, 0, len(cardNumber))
	for i := 0; i < len(cardNumber); i++ {
		c := cardNumber[i]
		if c < '0' || c > '9' {
			return false
		}
		digits = append(digits, int(c-'0'))
	}

	n := len(digits)
	sum := 0
	double := false

	for i := n - 1; i >= 0; i-- {
		d := digits[i]
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}

	return sum%10 == 0
}
