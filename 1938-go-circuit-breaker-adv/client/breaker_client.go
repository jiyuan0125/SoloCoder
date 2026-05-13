package client

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"circuit-breaker/circuit"
)

type BreakerClient struct {
	breaker *circuit.Breaker
	client  *http.Client
}

type FallbackFunc func(*http.Request) (*http.Response, error)

func NewBreakerClient(breaker *circuit.Breaker) *BreakerClient {
	return &BreakerClient{
		breaker: breaker,
		client:  &http.Client{},
	}
}

func NewBreakerClientWithHTTPClient(breaker *circuit.Breaker, httpClient *http.Client) *BreakerClient {
	return &BreakerClient{
		breaker: breaker,
		client:  httpClient,
	}
}

func (bc *BreakerClient) Do(req *http.Request) (*http.Response, error) {
	return bc.DoWithFallback(req, nil)
}

func (bc *BreakerClient) DoWithFallback(req *http.Request, fallback FallbackFunc) (*http.Response, error) {
	allowed, err := bc.breaker.Allow()
	if err != nil {
		if fallback != nil {
			return fallback(req)
		}
		return nil, err
	}

	if !allowed {
		if fallback != nil {
			return fallback(req)
		}
		return nil, circuit.ErrBreakerOpen
	}

	start := time.Now()
	
	resp, err := bc.client.Do(req)
	duration := time.Since(start)
	
	if err != nil {
		if isConnectionError(err) {
			bc.breaker.Failure(duration)
			if fallback != nil {
				return fallback(req)
			}
			return nil, err
		}
		
		bc.breaker.Failure(duration)
		if fallback != nil {
			return fallback(req)
		}
		return resp, err
	}
	
	if resp.StatusCode >= 500 {
		bc.breaker.Failure(duration)
		if fallback != nil {
			return fallback(req)
		}
		return resp, nil
	}
	
	bc.breaker.Success(duration)
	return resp, nil
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}
	
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	
	errMsg := fmt.Sprintf("%v", err)
	connectionErrors := []string{
		"connection refused",
		"connection reset",
		"i/o timeout",
		"no route to host",
		"no such host",
		"tls handshake timeout",
	}
	
	for _, connErr := range connectionErrors {
		if contains(errMsg, connErr) {
			return true
		}
	}
	
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > 0 && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
			indexOf(s, substr) >= 0)))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
