package creditcard

import "errors"

const MaxGenerateCount = 1000

var (
	ErrUnsupportedCardType = errors.New("unsupported card type")
	ErrInvalidCount        = errors.New("count must be positive")
	ErrCountExceedsLimit   = errors.New("count exceeds maximum limit of 1000")
	ErrInvalidCardNumber   = errors.New("invalid card number")
)
