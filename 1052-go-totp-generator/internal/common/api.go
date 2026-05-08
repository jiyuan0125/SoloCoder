package common

type RegisterRequest struct {
	Account string `json:"account"`
	Issuer  string `json:"issuer,omitempty"`
}

type RegisterResponse struct {
	Secret    string `json:"secret"`
	QRCodeURL string `json:"qrcode_url"`
}

type GenerateRequest struct {
	Secret string `json:"secret"`
}

type GenerateResponse struct {
	TOTP string `json:"totp"`
}

type ValidateRequest struct {
	Secret string `json:"secret"`
	TOTP   string `json:"totp"`
	Window int    `json:"window,omitempty"`
}

type ValidateResponse struct {
	Valid bool `json:"valid"`
}
