package api

type CheckRequest struct {
	ISBN string `json:"isbn"`
}

type CheckResponse struct {
	Success        bool   `json:"success"`
	IsValid        bool   `json:"is_valid"`
	ISBNType       string `json:"isbn_type,omitempty"`
	CheckDigit     string `json:"check_digit,omitempty"`
	CalculatedCD   string `json:"calculated_check_digit,omitempty"`
	Message        string `json:"message"`
}

type ConvertRequest struct {
	ISBN string `json:"isbn"`
}

type ConvertResponse struct {
	Success        bool   `json:"success"`
	Original       string `json:"original,omitempty"`
	OriginalType   string `json:"original_type,omitempty"`
	Converted      string `json:"converted,omitempty"`
	ConvertedType  string `json:"converted_type,omitempty"`
	CheckDigit     string `json:"check_digit,omitempty"`
	Message        string `json:"message"`
}

type FormatRequest struct {
	ISBN string `json:"isbn"`
}

type FormatResponse struct {
	Success        bool   `json:"success"`
	Original       string `json:"original,omitempty"`
	Formatted      string `json:"formatted,omitempty"`
	Message        string `json:"message"`
}
