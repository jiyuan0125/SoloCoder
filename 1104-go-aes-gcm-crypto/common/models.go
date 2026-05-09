package common

type EncryptRequest struct {
	Plaintext string `json:"plaintext"`
	Password  string `json:"password"`
}

type EncryptResponse struct {
	Ciphertext string `json:"ciphertext"`
	Error      string `json:"error,omitempty"`
}

type DecryptRequest struct {
	Ciphertext string `json:"ciphertext"`
	Password   string `json:"password"`
}

type DecryptResponse struct {
	Plaintext string `json:"plaintext"`
	Error     string `json:"error,omitempty"`
}

type ConfigRequest struct {
	Iterations int `json:"iterations"`
}

type ConfigResponse struct {
	Iterations int    `json:"iterations"`
	Error      string `json:"error,omitempty"`
}
