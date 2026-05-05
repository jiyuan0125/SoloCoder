package main

import (
	"announcement-board/common"
	"encoding/json"
	"net/http"
	"strings"
)

type APIError struct {
	Code    int
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.APIResponse{
		Success: true,
		Data:    data,
	})
}

func respondError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	code := 500
	message := "内部服务器错误"
	if apiErr, ok := err.(*APIError); ok {
		code = apiErr.Code
		message = apiErr.Message
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(common.APIResponse{
		Success: false,
		Message: message,
	})
}

func getUserID(r *http.Request) string {
	return r.Header.Get("X-User-ID")
}

func handleListDepartments(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	depts := GetAllDepartments()
	result := make([]common.Department, 0, len(depts))
	for _, d := range depts {
		result = append(result, *d)
	}
	respondJSON(w, common.DepartmentListResponse{Departments: result})
}

func handleGetUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	user, err := GetUserInfo(userID)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, common.UserInfoResponse{User: *user})
}

func handleCreateAnnouncement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	var req common.CreateAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, &APIError{Code: 400, Message: "请求参数错误: " + err.Error()})
		return
	}

	ann, err := CreateAnnouncement(userID, &req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, ann)
}

func handleUpdateAnnouncement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		respondError(w, &APIError{Code: 400, Message: "公告ID不能为空"})
		return
	}
	annID := parts[len(parts)-1]

	var req common.UpdateAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, &APIError{Code: 400, Message: "请求参数错误: " + err.Error()})
		return
	}

	ann, err := UpdateAnnouncement(userID, annID, &req)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, ann)
}

func handleDeleteDraft(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		respondError(w, &APIError{Code: 400, Message: "公告ID不能为空"})
		return
	}
	annID := parts[len(parts)-1]

	if err := DeleteDraftAnnouncement(userID, annID); err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, map[string]string{"message": "删除成功"})
}

func handleGetAnnouncement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		respondError(w, &APIError{Code: 400, Message: "公告ID不能为空"})
		return
	}
	annID := parts[len(parts)-1]

	resp, err := GetAnnouncementDetail(userID, annID)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, resp)
}

func handleListEmployeeAnnouncements(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	keyword := r.URL.Query().Get("keyword")

	resp, err := ListEmployeeAnnouncements(userID, keyword)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, resp)
}

func handleListAdminAnnouncements(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	keyword := r.URL.Query().Get("keyword")

	resp, err := ListAdminAnnouncements(userID, keyword)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, resp)
}

func handleSubmitApproval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	var req common.SubmitApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, &APIError{Code: 400, Message: "请求参数错误: " + err.Error()})
		return
	}

	if err := SubmitApproval(userID, req.AnnouncementID); err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, map[string]string{"message": "已提交审批"})
}

func handleApproveAnnouncement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	var req common.ApproveAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, &APIError{Code: 400, Message: "请求参数错误: " + err.Error()})
		return
	}

	if err := ApproveAnnouncement(userID, req.AnnouncementID, req.Comment); err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, map[string]string{"message": "已审批通过"})
}

func handleRejectAnnouncement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	var req common.RejectAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, &APIError{Code: 400, Message: "请求参数错误: " + err.Error()})
		return
	}

	if err := RejectAnnouncement(userID, req.AnnouncementID, req.Comment); err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, map[string]string{"message": "已驳回审批"})
}

func handleListPendingApprovals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	resp, err := ListPendingApprovals(userID)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, resp)
}

func handleTogglePin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := getUserID(r)
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		respondError(w, &APIError{Code: 400, Message: "公告ID不能为空"})
		return
	}
	annID := parts[len(parts)-2]

	if err := TogglePin(userID, annID); err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, map[string]string{"message": "置顶状态已更新"})
}
