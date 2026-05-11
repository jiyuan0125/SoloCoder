package idtoken

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
)

func SplitToken(token string) ([]string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidFormat
	}
	return parts, nil
}

func Base64URLDecode(s string) ([]byte, error) {
	if l := len(s) % 4; l != 0 {
		s += strings.Repeat("=", 4-l)
	}
	return base64.URLEncoding.DecodeString(s)
}

func ParseHeader(headerPart string) (*Header, error) {
	headerBytes, err := Base64URLDecode(headerPart)
	if err != nil {
		return nil, ErrInvalidHeader
	}
	var header Header
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, ErrInvalidHeader
	}
	return &header, nil
}

func ParseClaims(claimsPart string) (*Claims, error) {
	claimsBytes, err := Base64URLDecode(claimsPart)
	if err != nil {
		return nil, ErrInvalidPayload
	}
	var claims Claims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, ErrInvalidPayload
	}
	var rawMap map[string]interface{}
	if err := json.Unmarshal(claimsBytes, &rawMap); err != nil {
		return nil, ErrInvalidPayload
	}
	claims.Raw = rawMap
	return &claims, nil
}

func ParseSignature(signaturePart string) ([]byte, error) {
	return Base64URLDecode(signaturePart)
}

func Parse(token string) (*Header, *Claims, []byte, string, error) {
	parts, err := SplitToken(token)
	if err != nil {
		return nil, nil, nil, "", err
	}
	signingInput := parts[0] + "." + parts[1]
	header, err := ParseHeader(parts[0])
	if err != nil {
		return nil, nil, nil, "", err
	}
	claims, err := ParseClaims(parts[1])
	if err != nil {
		return nil, nil, nil, "", err
	}
	signature, err := ParseSignature(parts[2])
	if err != nil {
		return nil, nil, nil, "", err
	}
	return header, claims, signature, signingInput, nil
}

func (c *Claims) GetAudience() ([]string, error) {
	if len(c.Aud) == 0 {
		return nil, nil
	}
	var audString string
	if err := json.Unmarshal(c.Aud, &audString); err == nil {
		return []string{audString}, nil
	}
	var audArray []string
	if err := json.Unmarshal(c.Aud, &audArray); err == nil {
		return audArray, nil
	}
	return nil, ErrInvalidAudience
}

func (c *Claims) ToJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(c.Raw); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
