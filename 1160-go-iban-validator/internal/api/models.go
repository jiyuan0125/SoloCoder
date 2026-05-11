package api

type ValidateRequest struct {
	IBAN string `json:"iban"`
}

type ValidateResponse struct {
	Valid         bool   `json:"valid"`
	CountryCode   string `json:"country_code"`
	CountryName   string `json:"country_name"`
	BBAN          string `json:"bban"`
	FormatValid   bool   `json:"format_valid"`
	ChecksumValid bool   `json:"checksum_valid"`
	Formatted     string `json:"formatted"`
}

type FormatRequest struct {
	IBAN string `json:"iban"`
}

type FormatResponse struct {
	Formatted string `json:"formatted"`
}

type BatchValidateRequest struct {
	IBANs []string `json:"ibans"`
}

type BatchValidateResponse struct {
	Results []ValidateResponse `json:"results"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
