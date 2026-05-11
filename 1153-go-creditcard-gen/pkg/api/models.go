package api

const (
	HeaderWarning       = "X-Test-Card-Warning"
	HeaderWarningValue  = "For testing purposes only. Not valid for real transactions."
)

type GenerateRequest struct {
	CardType string `json:"card_type"`
	Count    int    `json:"count"`
}

type CardInfo struct {
	Number      string `json:"number"`
	CardType    string `json:"card_type"`
	DisplayName string `json:"display_name"`
}

type GenerateResponse struct {
	Success    bool       `json:"success"`
	CardType   string     `json:"card_type"`
	Count      int        `json:"count"`
	Cards      []CardInfo `json:"cards"`
	Warning    string     `json:"warning"`
	Error      string     `json:"error,omitempty"`
}

type ValidateRequest struct {
	CardNumber string `json:"card_number"`
}

type ValidateResponse struct {
	Success    bool   `json:"success"`
	Valid      bool   `json:"valid"`
	CardNumber string `json:"card_number"`
	CardType   string `json:"card_type,omitempty"`
	Message    string `json:"message"`
	Warning    string `json:"warning"`
	Error      string `json:"error,omitempty"`
}
