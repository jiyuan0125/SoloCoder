package crawler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
)

func ValidateURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: must be http or https")
	}
	if parsed.Host == "" {
		return fmt.Errorf("invalid URL: missing host")
	}
	return nil
}

func NormalizeURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	
	parsed.Fragment = ""
	parsed.RawQuery = url.Values{}.Encode()
	return parsed.String(), nil
}

func GetDomain(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return parsed.Host, nil
}

func HashURL(rawURL string) string {
	hash := sha256.Sum256([]byte(rawURL))
	return hex.EncodeToString(hash[:])
}

func IsSameDomain(url1, url2 string) bool {
	domain1, err1 := GetDomain(url1)
	domain2, err2 := GetDomain(url2)
	if err1 != nil || err2 != nil {
		return false
	}
	return domain1 == domain2
}
