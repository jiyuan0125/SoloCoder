package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"notification-center/common"
	"strconv"
	"time"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://localhost:%d", common.DefaultPort)
	}
	return &Client{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) doRequest(method, path string, body interface{}, query url.Values) (*common.APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}
	
	reqURL := c.baseURL + path
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}
	
	req, err := http.NewRequest(method, reqURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	var apiResp common.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &apiResp, nil
}

func (c *Client) SendNotification(req common.SendRequest) (*common.SendResponse, error) {
	resp, err := c.doRequest(http.MethodPost, "/api/notification/send", req, nil)
	if err != nil {
		return nil, err
	}
	
	if resp.Code != common.ErrCodeSuccess {
		return nil, common.NewAPIError(resp.Code, resp.Message)
	}
	
	var data common.SendResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &data)
	return &data, nil
}

func (c *Client) ListNotifications(receiver string, notifType *common.NotificationType, 
	status *common.ReadStatus, startTime, endTime *time.Time, 
	includeArchived bool, page, pageSize int) (*common.ListResponse, error) {
	
	query := url.Values{}
	query.Set("receiver", receiver)
	
	if notifType != nil {
		query.Set("type", string(*notifType))
	}
	if status != nil {
		query.Set("status", string(*status))
	}
	if startTime != nil {
		query.Set("start_time", startTime.Format(time.RFC3339))
	}
	if endTime != nil {
		query.Set("end_time", endTime.Format(time.RFC3339))
	}
	if includeArchived {
		query.Set("include_archived", "true")
	}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		query.Set("page_size", strconv.Itoa(pageSize))
	}
	
	resp, err := c.doRequest(http.MethodGet, "/api/notification/list", nil, query)
	if err != nil {
		return nil, err
	}
	
	if resp.Code != common.ErrCodeSuccess {
		return nil, common.NewAPIError(resp.Code, resp.Message)
	}
	
	var data common.ListResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &data)
	return &data, nil
}

func (c *Client) MarkAsRead(receiver string, notificationIDs []string) (*common.MarkReadResponse, error) {
	req := common.MarkReadRequest{
		Receiver:        receiver,
		NotificationIDs: notificationIDs,
	}
	
	resp, err := c.doRequest(http.MethodPost, "/api/notification/mark-read", req, nil)
	if err != nil {
		return nil, err
	}
	
	if resp.Code != common.ErrCodeSuccess {
		return nil, common.NewAPIError(resp.Code, resp.Message)
	}
	
	var data common.MarkReadResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &data)
	return &data, nil
}

func (c *Client) GetUnreadCount(receiver string) (*common.UnreadCountResponse, error) {
	query := url.Values{}
	query.Set("receiver", receiver)
	
	resp, err := c.doRequest(http.MethodGet, "/api/notification/unread-count", nil, query)
	if err != nil {
		return nil, err
	}
	
	if resp.Code != common.ErrCodeSuccess {
		return nil, common.NewAPIError(resp.Code, resp.Message)
	}
	
	var data common.UnreadCountResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &data)
	return &data, nil
}

func (c *Client) CreateTemplate(req common.TemplateCreateRequest) (*common.NotificationTemplate, error) {
	resp, err := c.doRequest(http.MethodPost, "/api/template/create", req, nil)
	if err != nil {
		return nil, err
	}
	
	if resp.Code != common.ErrCodeSuccess {
		return nil, common.NewAPIError(resp.Code, resp.Message)
	}
	
	var data common.NotificationTemplate
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &data)
	return &data, nil
}

func (c *Client) UpdateTemplate(req common.TemplateUpdateRequest) (*common.NotificationTemplate, error) {
	resp, err := c.doRequest(http.MethodPut, "/api/template/update", req, nil)
	if err != nil {
		return nil, err
	}
	
	if resp.Code != common.ErrCodeSuccess {
		return nil, common.NewAPIError(resp.Code, resp.Message)
	}
	
	var data common.NotificationTemplate
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &data)
	return &data, nil
}

func (c *Client) ListTemplates() ([]common.NotificationTemplate, error) {
	resp, err := c.doRequest(http.MethodGet, "/api/template/list", nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resp.Code != common.ErrCodeSuccess {
		return nil, common.NewAPIError(resp.Code, resp.Message)
	}
	
	var data []common.NotificationTemplate
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &data)
	return data, nil
}

func (c *Client) DeleteTemplate(id string) error {
	query := url.Values{}
	query.Set("id", id)
	
	resp, err := c.doRequest(http.MethodDelete, "/api/template/delete", nil, query)
	if err != nil {
		return err
	}
	
	if resp.Code != common.ErrCodeSuccess {
		return common.NewAPIError(resp.Code, resp.Message)
	}
	
	return nil
}

func (c *Client) RecordActivity(userID string) error {
	req := common.UserActivityRequest{
		UserID: userID,
	}
	
	resp, err := c.doRequest(http.MethodPost, "/api/user/activity", req, nil)
	if err != nil {
		return err
	}
	
	if resp.Code != common.ErrCodeSuccess {
		return common.NewAPIError(resp.Code, resp.Message)
	}
	
	return nil
}

func (c *Client) ListFailedLogs() ([]common.FailedLog, error) {
	resp, err := c.doRequest(http.MethodGet, "/api/log/failed", nil, nil)
	if err != nil {
		return nil, err
	}
	
	if resp.Code != common.ErrCodeSuccess {
		return nil, common.NewAPIError(resp.Code, resp.Message)
	}
	
	var data []common.FailedLog
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &data)
	return data, nil
}
