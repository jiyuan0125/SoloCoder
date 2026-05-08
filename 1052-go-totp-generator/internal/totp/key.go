package totp

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"net/url"
	"strings"
)

func GenerateSecret() (string, error) {
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret), nil
}

func DecodeSecret(secret string) ([]byte, error) {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	padding := (8 - len(secret)%8) % 8
	if padding > 0 {
		secret += strings.Repeat("=", padding)
	}
	return base32.StdEncoding.DecodeString(secret)
}

func GenerateOTPURL(secret, account, issuer string) string {
	params := url.Values{}
	params.Set("secret", secret)
	if issuer != "" {
		params.Set("issuer", issuer)
	}

	var label string
	if issuer != "" {
		label = fmt.Sprintf("%s:%s", issuer, account)
	} else {
		label = account
	}

	u := &url.URL{
		Scheme:   "otpauth",
		Host:     "totp",
		Path:     "/" + label,
		RawQuery: params.Encode(),
	}

	return u.String()
}
