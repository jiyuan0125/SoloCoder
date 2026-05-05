package common

type CardType string

const (
	CardTypeDebit   CardType = "借记卡"
	CardTypeCredit  CardType = "信用卡"
	CardTypeUnknown CardType = "未知"
)

type ParseRequest struct {
	CardNumber string `json:"card_number"`
}

type ParseResponse struct {
	CardNumber       string   `json:"card_number"`
	BankName         string   `json:"bank_name"`
	CardType         CardType `json:"card_type"`
	IsValid          bool     `json:"is_valid"`
	Error            string   `json:"error,omitempty"`
	MaskedNumber     string   `json:"masked_number"`
	FormattedNumber  string   `json:"formatted_number"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
