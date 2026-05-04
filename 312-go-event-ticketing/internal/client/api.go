package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"event-ticketing/pkg/common"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) CreateEvent(req common.CreateEventRequest) (*common.CreateEventResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/events", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.CreateEventResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) PurchaseTicket(req common.PurchaseTicketRequest) (*common.PurchaseTicketResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/tickets/purchase", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.PurchaseTicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) CheckIn(req common.CheckInRequest) (*common.CheckInResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/tickets/checkin", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.CheckInResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) RefundTicket(req common.RefundTicketRequest) (*common.RefundTicketResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/tickets/refund", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.RefundTicketResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetEventStats(eventID string) (*common.GetEventStatsResponse, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/events/stats/%s", c.baseURL, eventID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.GetEventStatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) ListEvents() (*common.ListEventsResponse, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/events")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result common.ListEventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
