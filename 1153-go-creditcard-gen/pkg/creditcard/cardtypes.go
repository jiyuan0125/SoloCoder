package creditcard

import (
	"fmt"
	"math/rand"
	"strings"
)

type CardType string

const (
	CardTypeVisa            CardType = "visa"
	CardTypeMasterCard      CardType = "mastercard"
	CardTypeAmericanExpress CardType = "amex"
	CardTypeJCB             CardType = "jcb"
)

var SupportedCardTypes = []CardType{
	CardTypeVisa,
	CardTypeMasterCard,
	CardTypeAmericanExpress,
	CardTypeJCB,
}

type iinRange struct {
	prefixStart int
	prefixEnd   int
	length      int
}

var cardTypeRanges = map[CardType][]iinRange{
	CardTypeVisa: {
		{4, 4, 16},
	},
	CardTypeMasterCard: {
		{2221, 2720, 16},
		{51, 55, 16},
	},
	CardTypeAmericanExpress: {
		{34, 34, 15},
		{37, 37, 15},
	},
	CardTypeJCB: {
		{35, 35, 16},
	},
}

var cardTypeNames = map[CardType]string{
	CardTypeVisa:            "Visa",
	CardTypeMasterCard:      "MasterCard",
	CardTypeAmericanExpress: "American Express",
	CardTypeJCB:             "JCB",
}

func ParseCardType(s string) (CardType, error) {
	lower := strings.ToLower(strings.TrimSpace(s))
	for _, ct := range SupportedCardTypes {
		if string(ct) == lower {
			return ct, nil
		}
	}
	return "", fmt.Errorf("unsupported card type: %s", s)
}

func CardTypeDisplayName(ct CardType) string {
	if name, ok := cardTypeNames[ct]; ok {
		return name
	}
	return string(ct)
}

func IsSupportedCardType(ct CardType) bool {
	for _, supported := range SupportedCardTypes {
		if supported == ct {
			return true
		}
	}
	return false
}

func selectIINAndLength(ct CardType) (string, int) {
	ranges := cardTypeRanges[ct]
	selectedRange := ranges[rand.Intn(len(ranges))]

	var prefix string
	prefixLen := len(fmt.Sprintf("%d", selectedRange.prefixStart))

	if selectedRange.prefixStart == selectedRange.prefixEnd {
		prefix = fmt.Sprintf("%d", selectedRange.prefixStart)
	} else {
		prefix = fmt.Sprintf("%d", rand.Intn(selectedRange.prefixEnd-selectedRange.prefixStart+1)+selectedRange.prefixStart)
	}

	for len(prefix) < prefixLen {
		prefix = "0" + prefix
	}

	return prefix, selectedRange.length
}
