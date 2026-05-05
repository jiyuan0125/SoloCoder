package main

import (
	"announcement-board/common"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const ServerBaseURL = "http://localhost:8080/api"

var currentUserID string

func SetCurrentUser(userID string) {
	currentUserID = userID
}

func GetCurrentUser() string {
	return currentUserID
}

func doRequest(method, endpoint string, body interface{}) (*common.APIResponse, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求体失败: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, ServerBaseURL+endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("X-User-ID", currentUserID)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	var apiResp common.APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w, 响应内容: %s", err, string(respBody))
	}

	return &apiResp, nil
}

func GetUserInfo() (*common.User, error) {
	resp, err := doRequest(http.MethodGet, "/user/info", nil)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Message)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	dataJSON, _ := json.Marshal(dataMap)
	var userResp common.UserInfoResponse
	json.Unmarshal(dataJSON, &userResp)

	return &userResp.User, nil
}

func ListDepartments() ([]common.Department, error) {
	resp, err := doRequest(http.MethodGet, "/departments", nil)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Message)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	dataJSON, _ := json.Marshal(dataMap)
	var deptResp common.DepartmentListResponse
	json.Unmarshal(dataJSON, &deptResp)

	return deptResp.Departments, nil
}

func CreateAnnouncement(req *common.CreateAnnouncementRequest) (*common.Announcement, error) {
	resp, err := doRequest(http.MethodPost, "/announcements", req)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Message)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	dataJSON, _ := json.Marshal(dataMap)
	var ann common.Announcement
	json.Unmarshal(dataJSON, &ann)

	return &ann, nil
}

func UpdateAnnouncement(annID string, content string) (*common.Announcement, error) {
	req := common.UpdateAnnouncementRequest{Content: content}
	resp, err := doRequest(http.MethodPut, "/announcements/"+annID, req)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Message)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	dataJSON, _ := json.Marshal(dataMap)
	var ann common.Announcement
	json.Unmarshal(dataJSON, &ann)

	return &ann, nil
}

func DeleteDraft(annID string) error {
	resp, err := doRequest(http.MethodDelete, "/announcements/"+annID, nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Message)
	}
	return nil
}

func GetAnnouncementDetail(annID string) (*common.AnnouncementDetailResponse, error) {
	resp, err := doRequest(http.MethodGet, "/announcements/"+annID, nil)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Message)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	dataJSON, _ := json.Marshal(dataMap)
	var detailResp common.AnnouncementDetailResponse
	json.Unmarshal(dataJSON, &detailResp)

	return &detailResp, nil
}

func ListAnnouncements(keyword string) (*common.AnnouncementListResponse, error) {
	endpoint := "/announcements"
	if keyword != "" {
		endpoint += "?keyword=" + keyword
	}

	resp, err := doRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Message)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	dataJSON, _ := json.Marshal(dataMap)
	var listResp common.AnnouncementListResponse
	json.Unmarshal(dataJSON, &listResp)

	return &listResp, nil
}

func ListAdminAnnouncements(keyword string) (*common.AnnouncementListResponse, error) {
	endpoint := "/admin/announcements"
	if keyword != "" {
		endpoint += "?keyword=" + keyword
	}

	resp, err := doRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Message)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	dataJSON, _ := json.Marshal(dataMap)
	var listResp common.AnnouncementListResponse
	json.Unmarshal(dataJSON, &listResp)

	return &listResp, nil
}

func SubmitApproval(annID string) error {
	req := common.SubmitApprovalRequest{AnnouncementID: annID}
	resp, err := doRequest(http.MethodPost, "/approvals/submit", req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Message)
	}
	return nil
}

func ApproveAnnouncement(annID string, comment string) error {
	req := common.ApproveAnnouncementRequest{
		AnnouncementID: annID,
		Comment:        comment,
	}
	resp, err := doRequest(http.MethodPost, "/approvals/approve", req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Message)
	}
	return nil
}

func RejectAnnouncement(annID string, comment string) error {
	req := common.RejectAnnouncementRequest{
		AnnouncementID: annID,
		Comment:        comment,
	}
	resp, err := doRequest(http.MethodPost, "/approvals/reject", req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Message)
	}
	return nil
}

func ListPendingApprovals() (*common.ApprovalListResponse, error) {
	resp, err := doRequest(http.MethodGet, "/approvals/pending", nil)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return nil, fmt.Errorf("%s", resp.Message)
	}

	dataMap, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("响应数据格式错误")
	}

	dataJSON, _ := json.Marshal(dataMap)
	var listResp common.ApprovalListResponse
	json.Unmarshal(dataJSON, &listResp)

	return &listResp, nil
}

func TogglePin(annID string) error {
	resp, err := doRequest(http.MethodPost, "/announcements/"+annID+"/toggle-pin", nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("%s", resp.Message)
	}
	return nil
}
