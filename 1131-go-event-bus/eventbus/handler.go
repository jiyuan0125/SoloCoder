package eventbus

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func JSONHandler(dst interface{}, handler func(*EventContext, interface{}) HandlerResult) Handler {
	return func(ctx *EventContext) HandlerResult {
		payload := ctx.Event().Payload
		if err := json.Unmarshal(payload, dst); err != nil {
			return HandlerResult{
				Success: false,
				Error:   fmt.Errorf("unmarshal payload: %w", err),
			}
		}
		return handler(ctx, dst)
	}
}

func WebhookHandler(endpoint string, timeout time.Duration) Handler {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	return func(ctx *EventContext) HandlerResult {
		evt := ctx.Event()

		body := map[string]interface{}{
			"event_id":  evt.ID,
			"topic":     evt.Topic,
			"payload":   json.RawMessage(evt.Payload),
			"timestamp": evt.Timestamp.Format(time.RFC3339),
		}
		data, err := json.Marshal(body)
		if err != nil {
			return HandlerResult{Success: false, Error: err}
		}

		u, err := url.Parse(endpoint)
		if err != nil {
			return HandlerResult{Success: false, Error: err}
		}

		req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(data))
		if err != nil {
			return HandlerResult{Success: false, Error: err}
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Event-ID", evt.ID)
		req.Header.Set("X-Event-Topic", evt.Topic)

		resp, err := client.Do(req)
		if err != nil {
			return HandlerResult{Success: false, Error: err}
		}
		defer resp.Body.Close()

		_, _ = io.Copy(io.Discard, resp.Body)

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return HandlerResult{Success: true}
		}
		return HandlerResult{
			Success: false,
			Error:   fmt.Errorf("webhook returned status %d", resp.StatusCode),
		}
	}
}

func NopHandler() Handler {
	return func(ctx *EventContext) HandlerResult {
		return HandlerResult{Success: true}
	}
}
