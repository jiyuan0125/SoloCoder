package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/comment-moderator/pkg/common"
)

type APIClient struct {
	BaseURL string
}

func NewAPIClient(host string, port int) *APIClient {
	return &APIClient{
		BaseURL: fmt.Sprintf("http://%s:%d", host, port),
	}
}

func (c *APIClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *APIClient) HealthCheck() (string, error) {
	resp, err := c.doRequest("GET", "/health", nil)
	return string(resp), err
}

func (c *APIClient) SubmitComment(userID, content string) (*common.SubmitCommentResponse, error) {
	req := common.SubmitCommentRequest{
		UserID:  userID,
		Content: content,
	}

	respBytes, err := c.doRequest("POST", "/comments/submit", req)
	if err != nil {
		return nil, err
	}

	var resp common.SubmitCommentResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) EditComment(commentID, userID, content string) (*common.EditCommentResponse, error) {
	req := common.EditCommentRequest{
		CommentID: commentID,
		UserID:    userID,
		Content:   content,
	}

	respBytes, err := c.doRequest("PUT", "/comments/edit", req)
	if err != nil {
		return nil, err
	}

	var resp common.EditCommentResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) ReportComment(commentID, userID, reason string) (*common.ReportCommentResponse, error) {
	req := common.ReportCommentRequest{
		CommentID: commentID,
		UserID:    userID,
		Reason:    reason,
	}

	respBytes, err := c.doRequest("POST", "/comments/report", req)
	if err != nil {
		return nil, err
	}

	var resp common.ReportCommentResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) ViewRejectedComment(userID, commentID string) (*common.ViewRejectedCommentResponse, error) {
	path := fmt.Sprintf("/comments/rejected/%s?user_id=%s", commentID, userID)
	respBytes, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp common.ViewRejectedCommentResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) GetPendingComments(moderatorID string, limit, offset int) (*common.GetPendingCommentsResponse, error) {
	path := fmt.Sprintf("/moderate/pending?moderator_id=%s&limit=%d&offset=%d", moderatorID, limit, offset)
	respBytes, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var resp common.GetPendingCommentsResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) BatchApprove(moderatorID string, commentIDs []string) (*common.BatchActionResponse, error) {
	req := common.BatchApproveRequest{
		ModeratorID: moderatorID,
		CommentIDs:  commentIDs,
	}

	respBytes, err := c.doRequest("POST", "/moderate/approve", req)
	if err != nil {
		return nil, err
	}

	var resp common.BatchActionResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) BatchReject(moderatorID string, commentIDs []string, reason string) (*common.BatchActionResponse, error) {
	req := common.BatchRejectRequest{
		ModeratorID: moderatorID,
		CommentIDs:  commentIDs,
		Reason:      reason,
	}

	respBytes, err := c.doRequest("POST", "/moderate/reject", req)
	if err != nil {
		return nil, err
	}

	var resp common.BatchActionResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) AddSensitiveWord(adminID, word string, level common.SensitiveWordLevel) (*common.AddSensitiveWordResponse, error) {
	req := common.AddSensitiveWordRequest{
		AdminID: adminID,
		Word:    word,
		Level:   level,
	}

	respBytes, err := c.doRequest("POST", "/sensitive-words", req)
	if err != nil {
		return nil, err
	}

	var resp common.AddSensitiveWordResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) RemoveSensitiveWord(adminID, word string) (*common.RemoveSensitiveWordResponse, error) {
	req := common.RemoveSensitiveWordRequest{
		AdminID: adminID,
		Word:    word,
	}

	respBytes, err := c.doRequest("DELETE", "/sensitive-words", req)
	if err != nil {
		return nil, err
	}

	var resp common.RemoveSensitiveWordResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) GetSensitiveWords() (*common.GetSensitiveWordsResponse, error) {
	respBytes, err := c.doRequest("GET", "/sensitive-words", nil)
	if err != nil {
		return nil, err
	}

	var resp common.GetSensitiveWordsResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) RegisterModerator(adminID, name string) (*common.RegisterModeratorResponse, error) {
	req := common.RegisterModeratorRequest{
		AdminID: adminID,
		Name:    name,
	}

	respBytes, err := c.doRequest("POST", "/moderators", req)
	if err != nil {
		return nil, err
	}

	var resp common.RegisterModeratorResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) GetModeratorStats() (*common.GetModeratorStatsResponse, error) {
	respBytes, err := c.doRequest("GET", "/moderators/stats", nil)
	if err != nil {
		return nil, err
	}

	var resp common.GetModeratorStatsResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) GetAuditLogs() (*common.GetAuditLogsResponse, error) {
	respBytes, err := c.doRequest("GET", "/audit-logs", nil)
	if err != nil {
		return nil, err
	}

	var resp common.GetAuditLogsResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func ParseSensitiveWordLevel(levelStr string) (common.SensitiveWordLevel, error) {
	levelStr = strings.ToLower(levelStr)
	switch levelStr {
	case "severe", "严重":
		return common.LevelSevere, nil
	case "medium", "中等":
		return common.LevelMedium, nil
	case "mild", "轻微":
		return common.LevelMild, nil
	default:
		return "", fmt.Errorf("无效的敏感词级别: %s，有效值: severe/medium/mild 或 严重/中等/轻微", levelStr)
	}
}
