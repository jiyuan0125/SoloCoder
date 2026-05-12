package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type JWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid"`
}

type JWTClaims struct {
	Sub  string `json:"sub"`
	Role string `json:"role"`
	Iat  int64  `json:"iat"`
	Exp  int64  `json:"exp"`
}

type TokenData struct {
	Header *JWTHeader
	Claims *JWTClaims
}

func SignJWT(header JWTHeader, claims JWTClaims, secret []byte) (string, error) {
	header.Alg = "HS256"
	header.Typ = "JWT"

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}

	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsBytes)

	signingInput := headerB64 + "." + claimsB64

	h := hmac.New(sha256.New, secret)
	h.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return signingInput + "." + signature, nil
}

func ParseJWT(token string) (*TokenData, string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, "", fmt.Errorf("invalid token format")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, "", fmt.Errorf("decode header: %w", err)
	}

	claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", fmt.Errorf("decode claims: %w", err)
	}

	var header JWTHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, "", fmt.Errorf("unmarshal header: %w", err)
	}

	var claims JWTClaims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, "", fmt.Errorf("unmarshal claims: %w", err)
	}

	return &TokenData{Header: &header, Claims: &claims}, parts[0] + "." + parts[1], nil
}

func VerifySignature(signingInput string, signature string, secret []byte) bool {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(signingInput))
	expected := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func GenerateRandomKey() []byte {
	key := make([]byte, 32)
	rand.Read(key)
	return key
}

func GenerateKeyID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func NowUnix() int64 {
	return time.Now().Unix()
}
