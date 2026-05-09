package common

type FingerprintRequest struct {
	Text    string `json:"text"`
	NGram   int    `json:"ngram,omitempty"`
}

type FingerprintResponse struct {
	Fingerprint string `json:"fingerprint"`
}

type HammingDistanceRequest struct {
	FP1 string `json:"fp1"`
	FP2 string `json:"fp2"`
}

type HammingDistanceResponse struct {
	Distance int `json:"distance"`
}

type IsSimilarRequest struct {
	FP1       string `json:"fp1"`
	FP2       string `json:"fp2"`
	Threshold int    `json:"threshold,omitempty"`
}

type IsSimilarResponse struct {
	Similar bool `json:"similar"`
}

type DedupRequest struct {
	Docs      []string `json:"docs"`
	Threshold int      `json:"threshold,omitempty"`
	NGram     int      `json:"ngram,omitempty"`
}

type DedupResponse struct {
	Docs []string `json:"docs"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
