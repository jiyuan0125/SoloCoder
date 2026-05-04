package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"seating-arrangement/pkg/protocol"
)

type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *APIClient) doRequest(method, path string, body interface{}, result interface{}) error {
	var reader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求体失败: %w", err)
		}
		reader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("服务器返回错误状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}

	return nil
}

func (c *APIClient) CreateVenue(req *protocol.CreateVenueRequest) (*protocol.Venue, error) {
	var resp protocol.APIResponse
	if err := c.doRequest("POST", "/api/venues", req, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析场馆数据失败: %w", err)
	}

	var venue protocol.Venue
	if err := json.Unmarshal(dataBytes, &venue); err != nil {
		return nil, fmt.Errorf("解析场馆数据失败: %w", err)
	}

	return &venue, nil
}

func (c *APIClient) ListVenues() ([]*protocol.Venue, error) {
	var resp protocol.APIResponse
	if err := c.doRequest("GET", "/api/venues", nil, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析场馆列表失败: %w", err)
	}

	var venues []*protocol.Venue
	if err := json.Unmarshal(dataBytes, &venues); err != nil {
		return nil, fmt.Errorf("解析场馆列表失败: %w", err)
	}

	return venues, nil
}

func (c *APIClient) CreateSession(req *protocol.CreateSessionRequest) (*protocol.Session, error) {
	var resp protocol.APIResponse
	if err := c.doRequest("POST", "/api/sessions", req, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析场次数据失败: %w", err)
	}

	var session protocol.Session
	if err := json.Unmarshal(dataBytes, &session); err != nil {
		return nil, fmt.Errorf("解析场次数据失败: %w", err)
	}

	return &session, nil
}

func (c *APIClient) ListSessions() ([]*protocol.Session, error) {
	var resp protocol.APIResponse
	if err := c.doRequest("GET", "/api/sessions", nil, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析场次列表失败: %w", err)
	}

	var sessions []*protocol.Session
	if err := json.Unmarshal(dataBytes, &sessions); err != nil {
		return nil, fmt.Errorf("解析场次列表失败: %w", err)
	}

	return sessions, nil
}

func (c *APIClient) LockSeat(req *protocol.LockSeatRequest) (*protocol.Order, error) {
	var resp protocol.APIResponse
	if err := c.doRequest("POST", "/api/seats/lock", req, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析订单数据失败: %w", err)
	}

	var order protocol.Order
	if err := json.Unmarshal(dataBytes, &order); err != nil {
		return nil, fmt.Errorf("解析订单数据失败: %w", err)
	}

	return &order, nil
}

func (c *APIClient) ConfirmOrder(req *protocol.ConfirmOrderRequest) error {
	var resp protocol.APIResponse
	if err := c.doRequest("POST", "/api/orders/confirm", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func (c *APIClient) CancelOrder(req *protocol.CancelOrderRequest) error {
	var resp protocol.APIResponse
	if err := c.doRequest("POST", "/api/orders/cancel", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func (c *APIClient) ReleaseSeat(req *protocol.ReleaseSeatRequest) error {
	var resp protocol.APIResponse
	if err := c.doRequest("POST", "/api/seats/release", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	return nil
}

func (c *APIClient) GetSessionStats(sessionID string) (*protocol.SessionStats, error) {
	var resp protocol.APIResponse
	if err := c.doRequest("GET", "/api/sessions/"+sessionID+"/stats", nil, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析统计数据失败: %w", err)
	}

	var stats protocol.SessionStats
	if err := json.Unmarshal(dataBytes, &stats); err != nil {
		return nil, fmt.Errorf("解析统计数据失败: %w", err)
	}

	return &stats, nil
}

func (c *APIClient) GetAvailableSeats(sessionID, sectionName string) ([]*protocol.SeatDetail, error) {
	query := url.Values{}
	query.Set("session_id", sessionID)
	query.Set("section_name", sectionName)

	var resp protocol.APIResponse
	if err := c.doRequest("GET", "/api/seats/available?"+query.Encode(), nil, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析座位列表失败: %w", err)
	}

	var seats []*protocol.SeatDetail
	if err := json.Unmarshal(dataBytes, &seats); err != nil {
		return nil, fmt.Errorf("解析座位列表失败: %w", err)
	}

	return seats, nil
}

func (c *APIClient) GetSoldSeats(sessionID string) ([]*protocol.OrderDetail, error) {
	var resp protocol.APIResponse
	if err := c.doRequest("GET", "/api/sessions/"+sessionID+"/sold", nil, &resp); err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, fmt.Errorf(resp.Message)
	}

	dataBytes, err := json.Marshal(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("解析已售座位列表失败: %w", err)
	}

	var orders []*protocol.OrderDetail
	if err := json.Unmarshal(dataBytes, &orders); err != nil {
		return nil, fmt.Errorf("解析已售座位列表失败: %w", err)
	}

	return orders, nil
}
