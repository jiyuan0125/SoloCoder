package checker

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"health-checker/pkg/models"
)

func CheckHTTP(address string, timeout time.Duration) (models.ComponentStatus, string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
	if err != nil {
		return models.StatusUnhealthy, fmt.Sprintf("create request failed: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return models.StatusUnhealthy, fmt.Sprintf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return models.StatusHealthy, ""
	}

	return models.StatusUnhealthy, fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
}

func CheckTCP(address string, timeout time.Duration) (models.ComponentStatus, string) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return models.StatusUnhealthy, fmt.Sprintf("connection failed: %v", err)
	}
	defer conn.Close()

	return models.StatusHealthy, ""
}

func Check(component *models.Component) (models.ComponentStatus, string) {
	switch component.CheckType {
	case models.CheckTypeHTTP:
		return CheckHTTP(component.Address, component.Timeout)
	case models.CheckTypeTCP:
		return CheckTCP(component.Address, component.Timeout)
	default:
		return models.StatusUnknown, "unknown check type"
	}
}
