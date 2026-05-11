package oauth2

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"math/big"
	"strings"
)

const pkceCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~"

func GenerateCodeVerifier(length int) string {
	if length < 43 {
		length = 43
	}
	if length > 128 {
		length = 128
	}
	b := make([]byte, length)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(pkceCharset))))
		b[i] = pkceCharset[idx.Int64()]
	}
	return string(b)
}

func GenerateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func VerifyCodeChallenge(verifier, challenge string) bool {
	expected := GenerateCodeChallenge(verifier)
	return constantTimeCompare(expected, challenge)
}

func constantTimeCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

func isValidCodeVerifier(v string) bool {
	if len(v) < 43 || len(v) > 128 {
		return false
	}
	for _, c := range v {
		if !strings.ContainsRune(pkceCharset, c) {
			return false
		}
	}
	return true
}
