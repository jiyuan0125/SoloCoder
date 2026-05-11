package idtoken

import (
	"encoding/json"
	"errors"
	"time"
)

type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type Claims struct {
	Iss string `json:"iss"`
	Sub string `json:"sub"`
	Aud json.RawMessage `json:"aud"`
	Exp int64 `json:"exp"`
	Iat int64 `json:"iat"`
	Nonce string `json:"nonce,omitempty"`
	Acr string `json:"acr,omitempty"`
	Email string `json:"email,omitempty"`
	Name string `json:"name,omitempty"`
	Picture string `json:"picture,omitempty"`
	Raw map[string]interface{}
}

type Config struct {
	Issuer string
	ClientID string
	Secret []byte
}

type ValidationResult struct {
	Valid bool
	Claims *Claims
	Errors []error
}

var (
	ErrInvalidFormat = errors.New("invalid token format")
	ErrInvalidHeader = errors.New("invalid header")
	ErrInvalidPayload = errors.New("invalid payload")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrInvalidAlgorithm = errors.New("invalid algorithm, only HS256 is supported")
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenUsedInFuture = errors.New("token was issued in the future")
	ErrInvalidIssuer = errors.New("invalid issuer")
	ErrInvalidAudience = errors.New("invalid audience")
	ErrInvalidNonce = errors.New("invalid nonce")
	ErrInvalidAcr = errors.New("invalid acr value, must be urn:mfa or urn:password")
)

const (
	ClockSkew = 60 * time.Second
	NonceTTL = 5 * time.Minute
)
