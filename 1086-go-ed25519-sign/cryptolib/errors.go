package cryptolib

import "errors"

var (
	ErrInvalidBase64       = errors.New("invalid base64 encoding")
	ErrInvalidKeyLength    = errors.New("invalid key length")
	ErrInvalidKeyFormat    = errors.New("invalid key format")
	ErrInvalidSignature    = errors.New("signature verification failed")
	ErrKeyConversionFailed = errors.New("key conversion failed")
)
