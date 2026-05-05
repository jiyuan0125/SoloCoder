package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"tag-manager/common"
)

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{BaseURL: strings.TrimRight(baseURL, "/")}
}

func (c *Client) do(method, path string, body interface{}) (*common.Response, error) {
	var bodyBytes []byte
	var err error

	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp common.Response
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("%s", apiResp.Error)
	}

	return &apiResp, nil
}

func (c *Client) CreateTag(req *common.CreateTagRequest) (*common.CreateTagResponse, error) {
	resp, err := c.do(http.MethodPost, "/tags/create", req)
	if err != nil {
		return nil, err
	}

	var result common.CreateTagResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)
	return &result, nil
}

func (c *Client) UpdateTag(tagID string, req *common.UpdateTagRequest) error {
	_, err := c.do(http.MethodPut, "/tags/"+tagID+"/update", req)
	return err
}

func (c *Client) GetTag(tagID string) (*common.Tag, error) {
	resp, err := c.do(http.MethodGet, "/tags/"+tagID, nil)
	if err != nil {
		return nil, err
	}

	var result common.Tag
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)
	return &result, nil
}

func (c *Client) DeleteTag(tagID, operator string) (int64, error) {
	resp, err := c.do(http.MethodDelete, "/tags/"+tagID+"/delete?operator="+operator, nil)
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)

	if count, ok := result["impacted_users"].(float64); ok {
		return int64(count), nil
	}
	return 0, nil
}

func (c *Client) GetTagImpact(tagID string) (*common.GetTagImpactResponse, error) {
	resp, err := c.do(http.MethodGet, "/tags/"+tagID+"/impact", nil)
	if err != nil {
		return nil, err
	}

	var result common.GetTagImpactResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)
	return &result, nil
}

func (c *Client) ListTags(groupID string) (*common.ListTagsResponse, error) {
	path := "/tags"
	if groupID != "" {
		path += "?group_id=" + groupID
	}
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var result common.ListTagsResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)
	return &result, nil
}

func (c *Client) CreateTagGroup(req *common.CreateTagGroupRequest) (*common.CreateTagGroupResponse, error) {
	resp, err := c.do(http.MethodPost, "/groups/create", req)
	if err != nil {
		return nil, err
	}

	var result common.CreateTagGroupResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)
	return &result, nil
}

func (c *Client) GetTagGroup(groupID string) (*common.TagGroup, error) {
	resp, err := c.do(http.MethodGet, "/groups/"+groupID, nil)
	if err != nil {
		return nil, err
	}

	var result common.TagGroup
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)
	return &result, nil
}

func (c *Client) ListTagGroups() (*common.ListTagGroupsResponse, error) {
	resp, err := c.do(http.MethodGet, "/groups", nil)
	if err != nil {
		return nil, err
	}

	var result common.ListTagGroupsResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)
	return &result, nil
}

func (c *Client) ApplyTagToUser(tagID string, req *common.ApplyTagToUserRequest) error {
	_, err := c.do(http.MethodPost, "/tags/"+tagID+"/apply", req)
	return err
}

func (c *Client) BatchApplyTag(tagID string, req *common.BatchApplyTagRequest) (int, error) {
	resp, err := c.do(http.MethodPost, "/tags/"+tagID+"/batch-apply", req)
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)

	if count, ok := result["processed_users"].(float64); ok {
		return int(count), nil
	}
	return 0, nil
}

func (c *Client) RemoveUserTag(tagID, userID, operator string) error {
	_, err := c.do(http.MethodDelete, "/tags/"+tagID+"/remove?user_id="+userID+"&operator="+operator, nil)
	return err
}

func (c *Client) GetUserTags(userID string) (*common.GetUserTagsResponse, error) {
	resp, err := c.do(http.MethodGet, "/users/"+userID+"/tags", nil)
	if err != nil {
		return nil, err
	}

	var result common.GetUserTagsResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)
	return &result, nil
}

func (c *Client) FilterUsers(req *common.FilterUsersRequest) (*common.FilterUsersResponse, error) {
	resp, err := c.do(http.MethodPost, "/users/filter", req)
	if err != nil {
		return nil, err
	}

	var result common.FilterUsersResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)
	return &result, nil
}

func (c *Client) GetTagStats() (*common.GetTagStatsResponse, error) {
	resp, err := c.do(http.MethodGet, "/stats", nil)
	if err != nil {
		return nil, err
	}

	var result common.GetTagStatsResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)
	return &result, nil
}

func (c *Client) GetTagTrend() (*common.GetTagTrendResponse, error) {
	resp, err := c.do(http.MethodGet, "/trend", nil)
	if err != nil {
		return nil, err
	}

	var result common.GetTagTrendResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)
	return &result, nil
}

func (c *Client) ListAuditLogs() (*common.ListAuditLogsResponse, error) {
	resp, err := c.do(http.MethodGet, "/audit", nil)
	if err != nil {
		return nil, err
	}

	var result common.ListAuditLogsResponse
	dataBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(dataBytes, &result)
	return &result, nil
}
