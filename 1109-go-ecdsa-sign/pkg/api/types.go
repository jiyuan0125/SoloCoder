package api

type GenerateKeyRequest struct {
	Curve string `json:"curve"`
}

type GenerateKeyResponse struct {
	Success   bool   `json:"success"`
	PublicKey string `json:"public_key,omitempty"`
	Curve     string `json:"curve,omitempty"`
	Error     string `json:"error,omitempty"`
}

type PublicKeyResponse struct {
	Success   bool   `json:"success"`
	PublicKey string `json:"public_key,omitempty"`
	Curve     string `json:"curve,omitempty"`
	Error     string `json:"error,omitempty"`
}

type SignRequest struct {
	Message string `json:"message"`
}

type SignResponse struct {
	Success   bool   `json:"success"`
	Signature string `json:"signature,omitempty"`
	Error     string `json:"error,omitempty"`
}

type VerifyRequest struct {
	Message   string `json:"message"`
	Signature string `json:"signature"`
	PublicKey string `json:"public_key"`
}

type VerifyResponse struct {
	Success bool   `json:"success"`
	Valid   bool   `json:"valid,omitempty"`
	Error   string `json:"error,omitempty"`
}

type StatsResponse struct {
	Success    bool  `json:"success"`
	SignCount  int64 `json:"sign_count,omitempty"`
	KCacheSize int   `json:"k_cache_size,omitempty"`
	Error      string `json:"error,omitempty"`
}
