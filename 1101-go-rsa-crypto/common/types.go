package common

type PaddingMode string

const (
	PaddingPKCS1v15 PaddingMode = "pkcs1v15"
	PaddingOAEP     PaddingMode = "oaep"
)

type HashAlgorithm string

const (
	HashSHA1   HashAlgorithm = "sha1"
	HashSHA224 HashAlgorithm = "sha224"
	HashSHA256 HashAlgorithm = "sha256"
	HashSHA384 HashAlgorithm = "sha384"
	HashSHA512 HashAlgorithm = "sha512"
)

type KeyPairRequest struct {
	KeySize int `json:"key_size"`
}

type KeyPairResponse struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type EncryptRequest struct {
	Plaintext string        `json:"plaintext"`
	Padding   PaddingMode   `json:"padding"`
	HashAlgo  HashAlgorithm `json:"hash_algo,omitempty"`
	Label     string        `json:"label,omitempty"`
}

type EncryptResponse struct {
	Ciphertext string `json:"ciphertext"`
}

type DecryptRequest struct {
	Ciphertext string        `json:"ciphertext"`
	Padding    PaddingMode   `json:"padding"`
	HashAlgo   HashAlgorithm `json:"hash_algo,omitempty"`
	Label      string        `json:"label,omitempty"`
}

type DecryptResponse struct {
	Plaintext string `json:"plaintext"`
}

type KeyInfoResponse struct {
	KeySize       int           `json:"key_size"`
	Padding       PaddingMode   `json:"padding"`
	HasPublicKey  bool          `json:"has_public_key"`
	HasPrivateKey bool          `json:"has_private_key"`
}

type AsyncKeyGenRequest struct {
	KeySize int `json:"key_size"`
}

type AsyncKeyGenResponse struct {
	TaskID string `json:"task_id"`
}

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type KeyGenTask struct {
	TaskID     string       `json:"task_id"`
	Status     TaskStatus   `json:"status"`
	KeySize    int          `json:"key_size"`
	PublicKey  string       `json:"public_key,omitempty"`
	PrivateKey string       `json:"private_key,omitempty"`
	Error      string       `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
