package core

import (
	"strconv"
	"strings"
)

func ValidateISBN(isbn string) bool {
	isbn = strings.TrimSpace(isbn)
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")
	if len(isbn) != 10 && len(isbn) != 13 {
		return false
	}
	if len(isbn) == 10 {
		return validateISBN10(isbn)
	}
	return validateISBN13(isbn)
}

func validateISBN10(isbn string) bool {
	sum := 0
	for i := 0; i < 9; i++ {
		d, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return false
		}
		sum += d * (10 - i)
	}
	check := isbn[9]
	if check == 'X' || check == 'x' {
		sum += 10
	} else {
		d, err := strconv.Atoi(string(check))
		if err != nil {
			return false
		}
		sum += d
	}
	return sum%11 == 0
}

func validateISBN13(isbn string) bool {
	sum := 0
	for i := 0; i < 13; i++ {
		d, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return false
		}
		if i%2 == 0 {
			sum += d
		} else {
			sum += d * 3
		}
	}
	return sum%10 == 0
}
