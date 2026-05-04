package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"notification-hub/pkg/models"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) CreateNotification(req *models.CreateNotificationRequest) (*models.CreateNotificationResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/notifications", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, c.parseError(resp)
	}

	var result models.CreateNotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetStatistics() (*models.StatisticsResponse, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/api/statistics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result models.StatisticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) GetUserNotifications(userID string, isRead *bool) (*models.UserListNotificationsResponse, error) {
	apiURL := fmt.Sprintf("%s/api/users/%s/notifications", c.baseURL, url.PathEscape(userID))
	if isRead != nil {
		apiURL += fmt.Sprintf("?is_read=%v", *isRead)
	}

	resp, err := c.httpClient.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result models.UserListNotificationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) MarkAsRead(userID, deliveryID string) (*models.MarkAsReadResponse, error) {
	apiURL := fmt.Sprintf("%s/api/users/%s/notifications/%s/read", c.baseURL, url.PathEscape(userID), url.PathEscape(deliveryID))

	req, err := http.NewRequest(http.MethodPut, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.parseError(resp)
	}

	var result models.MarkAsReadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Client) parseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var errResp models.ErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return fmt.Errorf("API error [%d]: %s", resp.StatusCode, errResp.Error)
	}
	return fmt.Errorf("API error [%d]: %s", resp.StatusCode, string(body))
}
