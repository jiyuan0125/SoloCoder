package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go-feedback-handler/pkg/protocol"
)

const (
	DefaultBaseURL = "http://localhost:8080"
	DefaultTimeout = 30 * time.Second
)

type APIClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &APIClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

func (c *APIClient) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp protocol.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return fmt.Errorf("API error: status code %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *APIClient) CreateFeedback(userID, userName string, ftype protocol.FeedbackType, content string) (*protocol.CreateFeedbackResponse, error) {
	req := protocol.CreateFeedbackRequest{
		UserID:   userID,
		UserName: userName,
		Type:     ftype,
		Content:  content,
	}

	var resp protocol.CreateFeedbackResponse
	if err := c.doRequest(http.MethodPost, "/api/v1/feedbacks", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) GetFeedback(id string) (*protocol.Feedback, error) {
	var fb protocol.Feedback
	if err := c.doRequest(http.MethodGet, "/api/v1/feedbacks/"+id, nil, &fb); err != nil {
		return nil, err
	}
	return &fb, nil
}

func (c *APIClient) ListFeedbacks(req *protocol.ListFeedbackRequest) (*protocol.ListFeedbackResponse, error) {
	query := url.Values{}
	if req.Status != "" {
		query.Set("status", string(req.Status))
	}
	if req.Type != "" {
		query.Set("type", string(req.Type))
	}
	if req.Priority != "" {
		query.Set("priority", string(req.Priority))
	}
	if req.UserID != "" {
		query.Set("user_id", req.UserID)
	}
	if req.HandlerID != "" {
		query.Set("handler_id", req.HandlerID)
	}
	if req.TagID != "" {
		query.Set("tag_id", req.TagID)
	}
	if req.InReview != nil {
		query.Set("in_review", strconv.FormatBool(*req.InReview))
	}
	if req.Page > 0 {
		query.Set("page", strconv.Itoa(req.Page))
	}
	if req.PageSize > 0 {
		query.Set("page_size", strconv.Itoa(req.PageSize))
	}

	path := "/api/v1/feedbacks"
	if queryStr := query.Encode(); queryStr != "" {
		path += "?" + queryStr
	}

	var resp protocol.ListFeedbackResponse
	if err := c.doRequest(http.MethodGet, path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *APIClient) UpdateStatus(feedbackID string, status protocol.FeedbackStatus, handlerID, handlerName string, isInvalid bool) error {
	req := protocol.UpdateStatusRequest{
		FeedbackID:  feedbackID,
		Status:      status,
		HandlerID:   handlerID,
		HandlerName: handlerName,
		IsInvalid:   isInvalid,
	}

	return c.doRequest(http.MethodPut, "/api/v1/feedbacks/"+feedbackID+"/status", req, nil)
}

func (c *APIClient) AssignHandler(feedbackID, handlerID, handlerName string) error {
	req := protocol.AssignHandlerRequest{
		FeedbackID:  feedbackID,
		HandlerID:   handlerID,
		HandlerName: handlerName,
	}

	return c.doRequest(http.MethodPut, "/api/v1/feedbacks/"+feedbackID+"/assign", req, nil)
}

func (c *APIClient) AddComment(feedbackID, userID, userName, content string, isFromSupport bool) error {
	req := protocol.AddCommentRequest{
		FeedbackID:    feedbackID,
		UserID:        userID,
		UserName:      userName,
		Content:       content,
		IsFromSupport: isFromSupport,
	}

	return c.doRequest(http.MethodPost, "/api/v1/comments", req, nil)
}

func (c *APIClient) ListTags() ([]*protocol.Tag, error) {
	var tags []*protocol.Tag
	if err := c.doRequest(http.MethodGet, "/api/v1/tags", nil, &tags); err != nil {
		return nil, err
	}
	return tags, nil
}

func (c *APIClient) CreateTag(name, color string) (*protocol.Tag, error) {
	req := protocol.CreateTagRequest{
		Name:  name,
		Color: color,
	}

	var tag protocol.Tag
	if err := c.doRequest(http.MethodPost, "/api/v1/tags", req, &tag); err != nil {
		return nil, err
	}
	return &tag, nil
}

func (c *APIClient) AddTagToFeedback(feedbackID, tagID string) error {
	req := protocol.AddTagRequest{
		FeedbackID: feedbackID,
		TagID:      tagID,
	}

	return c.doRequest(http.MethodPost, "/api/v1/feedbacks/"+feedbackID+"/tags", req, nil)
}

func (c *APIClient) RemoveTagFromFeedback(feedbackID, tagID string) error {
	return c.doRequest(http.MethodDelete, "/api/v1/feedbacks/"+feedbackID+"/tags/"+tagID, nil, nil)
}

func (c *APIClient) GetNotifications(userID string) ([]*protocol.Notification, error) {
	var notifs []*protocol.Notification
	path := "/api/v1/notifications?user_id=" + url.QueryEscape(userID)
	if err := c.doRequest(http.MethodGet, path, nil, &notifs); err != nil {
		return nil, err
	}
	return notifs, nil
}

func (c *APIClient) MarkNotificationRead(id string) error {
	return c.doRequest(http.MethodPut, "/api/v1/notifications/"+id+"/read", nil, nil)
}

func (c *APIClient) GetUserLimit(userID string) (*protocol.UserLimitInfo, error) {
	var limit protocol.UserLimitInfo
	if err := c.doRequest(http.MethodGet, "/api/v1/users/"+userID+"/limit", nil, &limit); err != nil {
		return nil, err
	}
	return &limit, nil
}

func (c *APIClient) GetKPIStats(handlerID, period string) (*protocol.KPIStats, error) {
	path := "/api/v1/kpi/" + handlerID
	if period != "" {
		path += "?period=" + url.QueryEscape(period)
	}

	var stats protocol.KPIStats
	if err := c.doRequest(http.MethodGet, path, nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

func (c *APIClient) GetMonthlyReport(month string) (*protocol.MonthlyReport, error) {
	path := "/api/v1/reports/monthly"
	if month != "" {
		path += "?month=" + url.QueryEscape(month)
	}

	var report protocol.MonthlyReport
	if err := c.doRequest(http.MethodGet, path, nil, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func (c *APIClient) GenerateMonthlyReport(month string) (*protocol.MonthlyReport, error) {
	path := "/api/v1/reports/monthly/generate"
	if month != "" {
		path += "?month=" + url.QueryEscape(month)
	}

	var report protocol.MonthlyReport
	if err := c.doRequest(http.MethodPost, path, nil, &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func (c *APIClient) HealthCheck() error {
	var resp protocol.SuccessResponse
	if err := c.doRequest(http.MethodGet, "/health", nil, &resp); err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("health check failed")
	}
	return nil
}
