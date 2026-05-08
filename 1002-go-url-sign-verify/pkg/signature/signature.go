package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

func BuildSignString(params map[string]string, signatureKey string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		if k == signatureKey {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var builder strings.Builder
	for i, k := range keys {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(params[k])
	}
	return builder.String()
}

func CalculateHMACSHA256(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateSignature(params map[string]string, secret string, signatureKey string) string {
	signString := BuildSignString(params, signatureKey)
	return CalculateHMACSHA256(signString, secret)
}
