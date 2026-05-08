package core

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/health-aggregator/pkg/common"
)

type Prober interface {
	Probe(ctx context.Context, svc *common.ServiceConfig) *common.ProbeResult
}

type HTTPProber struct{}

func NewHTTPProber() *HTTPProber {
	return &HTTPProber{}
}

func (p *HTTPProber) Probe(ctx context.Context, svc *common.ServiceConfig) *common.ProbeResult {
	start := time.Now()
	result := &common.ProbeResult{
		Timestamp: start,
	}

	client := &http.Client{
		Timeout: svc.Timeout,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, svc.Target, nil)
	if err != nil {
		result.Error = err.Error()
		result.Latency = time.Since(start)
		return result
	}

	resp, err := client.Do(req)
	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && dnsErr.Err != "" {
			result.Error = "DNS resolution failed: " + dnsErr.Error()
		} else {
			result.Error = err.Error()
		}
		result.Latency = time.Since(start)
		return result
	}
	defer resp.Body.Close()

	result.Latency = time.Since(start)

	expectedCodes := svc.ExpectedStatusCodes
	if len(expectedCodes) == 0 {
		expectedCodes = []int{200, 201, 202, 203, 204, 205, 206, 207, 208, 226}
	}

	for _, code := range expectedCodes {
		if resp.StatusCode == code {
			result.Success = true
			return result
		}
	}

	result.Error = "unexpected status code: " + resp.Status
	return result
}

type TCPProber struct{}

func NewTCPProber() *TCPProber {
	return &TCPProber{}
}

func (p *TCPProber) Probe(ctx context.Context, svc *common.ServiceConfig) *common.ProbeResult {
	start := time.Now()
	result := &common.ProbeResult{
		Timestamp: start,
	}

	dialer := &net.Dialer{
		Timeout: svc.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", svc.Target)
	if err != nil {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) && dnsErr.Err != "" {
			result.Error = "DNS resolution failed: " + dnsErr.Error()
		} else {
			result.Error = err.Error()
		}
		result.Latency = time.Since(start)
		return result
	}
	defer conn.Close()

	result.Success = true
	result.Latency = time.Since(start)
	return result
}

type ProberFactory struct{}

func NewProberFactory() *ProberFactory {
	return &ProberFactory{}
}

func (f *ProberFactory) GetProber(probeType common.ProbeType) Prober {
	switch probeType {
	case common.ProbeTypeTCP:
		return NewTCPProber()
	case common.ProbeTypeHTTP:
		fallthrough
	default:
		return NewHTTPProber()
	}
}
