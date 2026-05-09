package types

type SignRequest struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Timestamp int64  `json:"timestamp"`
	Body      string `json:"body"`
}

type SignResponse struct {
	Signature string `json:"signature"`
	Version   int    `json:"version"`
}

type VerifyRequest struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Timestamp int64  `json:"timestamp"`
	Body      string `json:"body"`
	Signature string `json:"signature"`
	Version   *int   `json:"version,omitempty"`
}

type VerifyResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
}

type RotateKeyResponse struct {
	Success      bool `json:"success"`
	NewVersion   int  `json:"new_version"`
	PreviousVersion int `json:"previous_version"`
}

type KeyListResponse struct {
	CurrentVersion int   `json:"current_version"`
	Versions       []int `json:"versions"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
