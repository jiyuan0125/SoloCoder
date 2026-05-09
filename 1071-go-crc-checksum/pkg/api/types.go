package api

type CRCType string

const (
	CRCType32 CRCType = "crc32"
	CRCType64 CRCType = "crc64"
)

type CustomCRC32Config struct {
	Polynomial string `json:"polynomial"`
	Init       string `json:"init"`
	RefIn      bool   `json:"ref_in"`
	RefOut     bool   `json:"ref_out"`
	XorOut     string `json:"xor_out"`
}

type CustomCRC64Config struct {
	Polynomial string `json:"polynomial"`
	Init       string `json:"init"`
	RefIn      bool   `json:"ref_in"`
	RefOut     bool   `json:"ref_out"`
	XorOut     string `json:"xor_out"`
}

type CalculateRequest struct {
	Type          CRCType           `json:"type"`
	Variant       string            `json:"variant,omitempty"`
	Data          string            `json:"data,omitempty"`
	CustomCRC32   *CustomCRC32Config `json:"custom_crc32,omitempty"`
	CustomCRC64   *CustomCRC64Config `json:"custom_crc64,omitempty"`
	IsBase64      bool              `json:"is_base64,omitempty"`
}

type CalculateResponse struct {
	Success bool   `json:"success"`
	CRC     string `json:"crc,omitempty"`
	Error   string `json:"error,omitempty"`
}

type VerifyRequest struct {
	Type          CRCType           `json:"type"`
	Variant       string            `json:"variant,omitempty"`
	Data          string            `json:"data,omitempty"`
	ExpectedCRC   string            `json:"expected_crc"`
	CustomCRC32   *CustomCRC32Config `json:"custom_crc32,omitempty"`
	CustomCRC64   *CustomCRC64Config `json:"custom_crc64,omitempty"`
	IsBase64      bool              `json:"is_base64,omitempty"`
}

type VerifyResponse struct {
	Success bool   `json:"success"`
	Match   bool   `json:"match,omitempty"`
	Actual  string `json:"actual,omitempty"`
	Error   string `json:"error,omitempty"`
}

type VariantsResponse struct {
	Success bool                `json:"success"`
	CRC32   map[string]string   `json:"crc32_variants,omitempty"`
	CRC64   map[string]string   `json:"crc64_variants,omitempty"`
	Error   string              `json:"error,omitempty"`
}
