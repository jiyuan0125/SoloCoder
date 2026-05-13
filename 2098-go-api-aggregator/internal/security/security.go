package security

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

var localHostnames = map[string]bool{
	"localhost": true,
	"::1":       true,
	"0.0.0.0":   true,
}

func ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("url is empty")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url format: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme: %s", u.Scheme)
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("missing host in url")
	}

	return nil
}

func IsInternalAddress(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return true
	}

	host := u.Hostname()
	port := u.Port()

	if port == "8080" {
		return true
	}

	if localHostnames[strings.ToLower(host)] {
		return true
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() {
			return true
		}
		if ip.IsPrivate() {
			return true
		}
		if ip.IsLinkLocalUnicast() {
			return true
		}
		if ip.IsLinkLocalMulticast() {
			return true
		}
		if ip.IsMulticast() {
			return true
		}
		if ip.IsUnspecified() {
			return true
		}
	}

	lowerHost := strings.ToLower(host)
	for local := range localHostnames {
		if lowerHost == local {
			return true
		}
	}

	if strings.HasSuffix(lowerHost, ".internal") {
		return true
	}
	if strings.HasSuffix(lowerHost, ".local") {
		return true
	}

	return false
}

func ValidateEndpointURL(rawURL string) error {
	if err := ValidateURL(rawURL); err != nil {
		return err
	}

	if IsInternalAddress(rawURL) {
		return fmt.Errorf("internal addresses are not allowed (SSRF protection)")
	}

	return nil
}
