package creditcard

import (
	"strconv"
	"strings"
)

type ValidationResult struct {
	Valid       bool
	CardType    CardType
	CardTypeStr string
	Message     string
}

func ValidateCardNumber(cardNumber string) ValidationResult {
	cleaned := strings.ReplaceAll(cardNumber, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	if len(cleaned) < 13 || len(cleaned) > 19 {
		return ValidationResult{
			Valid:   false,
			Message: "card number has invalid length",
		}
	}

	if !ValidateLuhn(cleaned) {
		return ValidationResult{
			Valid:   false,
			Message: "Luhn check failed",
		}
	}

	ct := identifyCardType(cleaned)
	if ct == "" {
		return ValidationResult{
			Valid:   false,
			Message: "unrecognized card type",
		}
	}

	return ValidationResult{
		Valid:       true,
		CardType:    ct,
		CardTypeStr: CardTypeDisplayName(ct),
		Message:     "card number is valid",
	}
}

func identifyCardType(cardNumber string) CardType {
	if len(cardNumber) < 2 {
		return ""
	}

	if strings.HasPrefix(cardNumber, "4") {
		if len(cardNumber) == 16 {
			return CardTypeVisa
		}
	}

	if len(cardNumber) >= 4 {
		firstFour, _ := strconv.Atoi(cardNumber[0:4])
		if firstFour >= 2221 && firstFour <= 2720 {
			if len(cardNumber) == 16 {
				return CardTypeMasterCard
			}
		}
	}

	if len(cardNumber) >= 2 {
		firstTwo, _ := strconv.Atoi(cardNumber[0:2])
		if firstTwo >= 51 && firstTwo <= 55 {
			if len(cardNumber) == 16 {
				return CardTypeMasterCard
			}
		}
	}

	if strings.HasPrefix(cardNumber, "34") || strings.HasPrefix(cardNumber, "37") {
		if len(cardNumber) == 15 {
			return CardTypeAmericanExpress
		}
	}

	if strings.HasPrefix(cardNumber, "35") {
		if len(cardNumber) == 16 {
			return CardTypeJCB
		}
	}

	return ""
}
