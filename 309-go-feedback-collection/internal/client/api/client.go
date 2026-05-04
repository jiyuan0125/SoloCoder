package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"feedback-system/pkg/common"
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

func (c *Client) doRequest(method, path string, body interface{}, respObj interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if respObj != nil {
		if err := json.NewDecoder(resp.Body).Decode(respObj); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) SubmitFeedback(req common.SubmitFeedbackRequest) (*common.SubmitFeedbackResponse, error) {
	var resp common.SubmitFeedbackResponse
	if err := c.doRequest("POST", "/api/feedback", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) ListFeedbacks(req common.ListFeedbackRequest) (*common.ListFeedbackResponse, error) {
	query := url.Values{}
	if req.Page > 0 {
		query.Set("page", strconv.Itoa(req.Page))
	}
	if req.PageSize > 0 {
		query.Set("page_size", strconv.Itoa(req.PageSize))
	}
	if req.Type != "" {
		query.Set("type", req.Type)
	}
	if req.Status != "" {
		query.Set("status", req.Status)
	}

	path := "/api/feedback"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	var resp common.ListFeedbackResponse
	if err := c.doRequest("GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetFeedback(id string) (*common.GetFeedbackResponse, error) {
	var resp common.GetFeedbackResponse
	if err := c.doRequest("GET", "/api/feedback/"+id, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) AddNote(id, note string) (*common.AddNoteResponse, error) {
	req := common.AddNoteRequest{
		FeedbackID: id,
		Note:       note,
	}
	var resp common.AddNoteResponse
	if err := c.doRequest("POST", "/api/feedback/"+id+"/note", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) UpdateStatus(id, status, processNote string) (*common.UpdateStatusResponse, error) {
	req := common.UpdateStatusRequest{
		FeedbackID:  id,
		Status:      status,
		ProcessNote: processNote,
	}
	var resp common.UpdateStatusResponse
	if err := c.doRequest("PUT", "/api/feedback/"+id+"/status", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) GetStatistics() (*common.StatisticsResponse, error) {
	var resp common.StatisticsResponse
	if err := c.doRequest("GET", "/api/statistics", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
