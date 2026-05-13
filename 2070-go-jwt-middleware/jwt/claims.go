package jwt

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

type Header struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
	Kid string `json:"kid,omitempty"`
}

type Claims struct {
	Sub  string   `json:"sub"`
	UserID string `json:"user_id"`
	Role string   `json:"role"`
	Jti  string   `json:"jti"`
	Iss  string   `json:"iss"`
	Aud  string   `json:"aud"`
	Iat  int64    `json:"iat"`
	Exp  int64    `json:"exp"`
	Type string   `json:"type"`
}

func Base64URLDecode(s string) ([]byte, error) {
	if l := len(s) % 4; l != 0 {
		s += strings.Repeat("=", 4-l)
	}
	return base64.URLEncoding.DecodeString(s)
}

func Base64URLEncode(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

func (h *Header) Encode() (string, error) {
	b, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	return Base64URLEncode(b), nil
}

func (h *Header) Decode(s string) error {
	b, err := Base64URLDecode(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, h)
}

func (c *Claims) Encode() (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return Base64URLEncode(b), nil
}

func (c *Claims) Decode(s string) error {
	b, err := Base64URLDecode(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, c)
}

func (c *Claims) IsExpired() bool {
	return time.Now().Unix() > c.Exp
}

func (c *Claims) ExpiresAt() time.Time {
	return time.Unix(c.Exp, 0)
}
