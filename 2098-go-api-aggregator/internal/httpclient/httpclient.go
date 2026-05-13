package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	MinTimeoutSeconds = 1
	MaxRetries        = 3
)

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Endpoint   string
}

type Error struct {
	Endpoint string
	Err      error
}

func (e *Error) Error() string {
	return fmt.Sprintf("endpoint %s error: %v", e.Endpoint, e.Err)
}

func CallEndpoint(method string, urlStr string, headers map[string]string, body []byte, timeoutSeconds int, retries int) (*Response, error) {
	if timeoutSeconds < MinTimeoutSeconds {
		timeoutSeconds = MinTimeoutSeconds
	}
	if retries > MaxRetries {
		retries = MaxRetries
	}
	if retries < 0 {
		retries = 0
	}

	var lastErr error

	for attempt := 0; attempt <= retries; attempt++ {
		resp, err := doCall(method, urlStr, headers, body, timeoutSeconds)
		if err == nil {
			return resp, nil
		}
		lastErr = err

		if attempt < retries {
			backoff := time.Duration(attempt+1) * 500 * time.Millisecond
			time.Sleep(backoff)
		}
	}

	return nil, &Error{Endpoint: urlStr, Err: lastErr}
}

func doCall(method string, urlStr string, headers map[string]string, body []byte, timeoutSeconds int) (*Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       respBody,
		Endpoint:   urlStr,
	}, nil
}
