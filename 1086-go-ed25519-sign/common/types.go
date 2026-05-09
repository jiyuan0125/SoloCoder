package common

type GenerateKeyResponse struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type SignRequest struct {
	PrivateKey string `json:"private_key"`
	Message    string `json:"message"`
}

type SignResponse struct {
	Signature string `json:"signature"`
}

type VerifyRequest struct {
	PublicKey string `json:"public_key"`
	Message   string `json:"message"`
	Signature string `json:"signature"`
}

type VerifyResponse struct {
	Valid bool `json:"valid"`
}

type ConvertPublicKeyRequest struct {
	PublicKey string `json:"public_key"`
}

type ConvertPublicKeyResponse struct {
	X25519PublicKey string `json:"x25519_public_key"`
}

type ConvertPrivateKeyRequest struct {
	PrivateKey string `json:"private_key"`
}

type ConvertPrivateKeyResponse struct {
	X25519PrivateKey string `json:"x25519_private_key"`
}

type InspectKeyRequest struct {
	Key string `json:"key"`
}

type InspectKeyResponse struct {
	Type        string `json:"type"`
	EncodedLen  int    `json:"encoded_len"`
	DecodedLen  int    `json:"decoded_len"`
	Fingerprint string `json:"fingerprint"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
