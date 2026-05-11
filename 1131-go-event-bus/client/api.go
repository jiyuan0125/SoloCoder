package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"eventbus-demo/common"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) postJSON(path string, body interface{}, out interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(data)
	}

	u := c.baseURL + path
	req, err := http.NewRequest(http.MethodPost, u, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp common.Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("invalid response: %s", string(respBody))
	}

	if apiResp.Code != common.CodeSuccess {
		return fmt.Errorf("server error [%d]: %s", apiResp.Code, apiResp.Message)
	}

	if out != nil {
		data, _ := json.Marshal(apiResp.Data)
		_ = json.Unmarshal(data, out)
	}
	return nil
}

func (c *Client) getJSON(path string, query map[string]string, out interface{}) error {
	u := c.baseURL + path
	if len(query) > 0 {
		q := url.Values{}
		for k, v := range query {
			q.Set(k, v)
		}
		u += "?" + q.Encode()
	}

	resp, err := c.http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp common.Response
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return fmt.Errorf("invalid response: %s", string(respBody))
	}

	if apiResp.Code != common.CodeSuccess {
		return fmt.Errorf("server error [%d]: %s", apiResp.Code, apiResp.Message)
	}

	if out != nil {
		data, _ := json.Marshal(apiResp.Data)
		_ = json.Unmarshal(data, out)
	}
	return nil
}

func (c *Client) CreateTopic(topic string) error {
	err := c.postJSON("/topic/create", common.CreateTopicReq{Topic: topic}, nil)
	if err != nil {
		return err
	}
	fmt.Printf("OK: topic '%s' created\n", topic)
	return nil
}

func (c *Client) DeleteTopic(topic string) error {
	err := c.postJSON("/topic/delete", common.DeleteTopicReq{Topic: topic}, nil)
	if err != nil {
		return err
	}
	fmt.Printf("OK: topic '%s' deleted\n", topic)
	return nil
}

func (c *Client) Subscribe(topic, subID string, priority int, endpoint string) error {
	err := c.postJSON("/topic/subscribe", common.SubscribeReq{
		Topic:    topic,
		SubID:    subID,
		Priority: priority,
		Endpoint: endpoint,
	}, nil)
	if err != nil {
		return err
	}
	if endpoint != "" {
		fmt.Printf("OK: subscriber '%s' subscribed to '%s' (priority=%d, endpoint=%s)\n", subID, topic, priority, endpoint)
	} else {
		fmt.Printf("OK: subscriber '%s' subscribed to '%s' (priority=%d)\n", subID, topic, priority)
	}
	return nil
}

func (c *Client) Unsubscribe(topic, subID string) error {
	err := c.postJSON("/topic/unsubscribe", common.UnsubscribeReq{
		Topic: topic,
		SubID: subID,
	}, nil)
	if err != nil {
		return err
	}
	fmt.Printf("OK: subscriber '%s' unsubscribed from '%s'\n", subID, topic)
	return nil
}

func (c *Client) ListSubscribers(topic string) error {
	var data map[string]interface{}
	err := c.getJSON("/topic/subscribers", map[string]string{"topic": topic}, &data)
	if err != nil {
		return err
	}

	fmt.Printf("Subscribers of '%s':\n", topic)
	subsRaw, ok := data["subscribers"].([]interface{})
	if !ok || len(subsRaw) == 0 {
		fmt.Println("  (none)")
		return nil
	}

	for i, sRaw := range subsRaw {
		var info common.SubscriberInfo
		sData, _ := json.Marshal(sRaw)
		_ = json.Unmarshal(sData, &info)
		active := "inactive"
		if info.IsActive {
			active = "active"
		}
		fmt.Printf("  %d. %s [priority=%d, %s]", i+1, info.ID, info.Priority, active)
		if info.Endpoint != "" {
			fmt.Printf(" -> %s", info.Endpoint)
		}
		fmt.Println()
	}
	return nil
}

func (c *Client) Publish(topic, payload string, async bool) error {
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		return fmt.Errorf("invalid JSON payload: %w", err)
	}

	var result common.EventResult
	err := c.postJSON("/event/publish", common.PublishReq{
		Topic:   topic,
		Payload: raw,
		Async:   async,
	}, &result)
	if err != nil {
		return err
	}

	fmt.Println("Event published:")
	fmt.Printf("  Event ID:   %s\n", result.EventID)
	fmt.Printf("  Status:     %s\n", result.Status)
	fmt.Printf("  Total:      %d\n", result.Total)
	fmt.Printf("  Success:    %d\n", result.Success)
	fmt.Printf("  Failed:     %d\n", result.Failed)
	fmt.Printf("  Intercepted: %v\n", result.Intercepted)
	return nil
}

func (c *Client) EventStatus(eventID string) error {
	var result common.EventResult
	err := c.getJSON("/event/status", map[string]string{"event_id": eventID}, &result)
	if err != nil {
		return err
	}

	fmt.Println("Event Status:")
	fmt.Printf("  Event ID:   %s\n", result.EventID)
	fmt.Printf("  Status:     %s\n", result.Status)
	fmt.Printf("  Total:      %d\n", result.Total)
	fmt.Printf("  Success:    %d\n", result.Success)
	fmt.Printf("  Failed:     %d\n", result.Failed)
	fmt.Printf("  Intercepted: %v\n", result.Intercepted)
	return nil
}
