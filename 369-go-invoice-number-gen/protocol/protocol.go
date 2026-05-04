package protocol

type GenerateRequest struct {
	Prefix string `json:"prefix"`
}

type GenerateResponse struct {
	InvoiceNumber string `json:"invoice_number"`
	Error         string `json:"error,omitempty"`
}
