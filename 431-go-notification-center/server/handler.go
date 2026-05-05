package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"notification-center/common"
	"strconv"
	"time"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Max-Age", "86400")
}

func handleOptions(w http.ResponseWriter) {
	setCORSHeaders(w)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) writeResponse(w http.ResponseWriter, code common.ErrorCode, message string, data interface{}) {
	setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	
	var statusCode int
	switch code {
	case common.ErrCodeSuccess:
		statusCode = http.StatusOK
	case common.ErrCodeInvalidRequest:
		statusCode = http.StatusBadRequest
	case common.ErrCodeNotFound:
		statusCode = http.StatusNotFound
	case common.ErrCodeDuplicate:
		statusCode = http.StatusConflict
	case common.ErrCodeExceedLimit:
		statusCode = http.StatusBadRequest
	case common.ErrCodeInvalidTemplate:
		statusCode = http.StatusBadRequest
	case common.ErrCodeSendFailed:
		statusCode = http.StatusInternalServerError
	default:
		statusCode = http.StatusInternalServerError
	}
	
	w.WriteHeader(statusCode)
	
	response := common.APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
	
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) writeSuccess(w http.ResponseWriter, data interface{}) {
	h.writeResponse(w, common.ErrCodeSuccess, "success", data)
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	if apiErr, ok := err.(common.APIError); ok {
		h.writeResponse(w, apiErr.Code, apiErr.Message, nil)
		return
	}
	h.writeResponse(w, common.ErrCodeInternal, err.Error(), nil)
}

func (h *Handler) HandleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w)
		return
	}
	if r.Method != http.MethodPost {
		h.writeError(w, common.ErrInvalidRequest("method not allowed"))
		return
	}
	
	var req common.SendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, common.ErrInvalidRequest(fmt.Sprintf("invalid request body: %v", err)))
		return
	}
	
	resp, err := h.service.SendNotification(req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	
	h.writeSuccess(w, resp)
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w)
		return
	}
	if r.Method != http.MethodGet {
		h.writeError(w, common.ErrInvalidRequest("method not allowed"))
		return
	}
	
	query := r.URL.Query()
	receiver := query.Get("receiver")
	if receiver == "" {
		h.writeError(w, common.ErrInvalidRequest("receiver is required"))
		return
	}
	
	req := common.ListRequest{
		Receiver:        receiver,
		IncludeArchived: query.Get("include_archived") == "true",
	}
	
	if typeStr := query.Get("type"); typeStr != "" {
		notifType := common.NotificationType(typeStr)
		req.Type = &notifType
	}
	
	if statusStr := query.Get("status"); statusStr != "" {
		status := common.ReadStatus(statusStr)
		req.Status = &status
	}
	
	if startTimeStr := query.Get("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			req.StartTime = &t
		}
	}
	
	if endTimeStr := query.Get("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			req.EndTime = &t
		}
	}
	
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			req.Page = p
		}
	}
	
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil {
			req.PageSize = ps
		}
	}
	
	resp, err := h.service.ListNotifications(req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	
	h.writeSuccess(w, resp)
}

func (h *Handler) HandleMarkRead(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w)
		return
	}
	if r.Method != http.MethodPost {
		h.writeError(w, common.ErrInvalidRequest("method not allowed"))
		return
	}
	
	var req common.MarkReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, common.ErrInvalidRequest(fmt.Sprintf("invalid request body: %v", err)))
		return
	}
	
	resp, err := h.service.MarkAsRead(req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	
	h.writeSuccess(w, resp)
}

func (h *Handler) HandleUnreadCount(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w)
		return
	}
	if r.Method != http.MethodGet {
		h.writeError(w, common.ErrInvalidRequest("method not allowed"))
		return
	}
	
	receiver := r.URL.Query().Get("receiver")
	if receiver == "" {
		h.writeError(w, common.ErrInvalidRequest("receiver is required"))
		return
	}
	
	resp, err := h.service.GetUnreadCount(receiver)
	if err != nil {
		h.writeError(w, err)
		return
	}
	
	h.writeSuccess(w, resp)
}

func (h *Handler) HandleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w)
		return
	}
	if r.Method != http.MethodPost {
		h.writeError(w, common.ErrInvalidRequest("method not allowed"))
		return
	}
	
	var req common.TemplateCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, common.ErrInvalidRequest(fmt.Sprintf("invalid request body: %v", err)))
		return
	}
	
	template, err := h.service.CreateTemplate(req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	
	h.writeSuccess(w, template)
}

func (h *Handler) HandleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w)
		return
	}
	if r.Method != http.MethodPut {
		h.writeError(w, common.ErrInvalidRequest("method not allowed"))
		return
	}
	
	var req common.TemplateUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, common.ErrInvalidRequest(fmt.Sprintf("invalid request body: %v", err)))
		return
	}
	
	template, err := h.service.UpdateTemplate(req)
	if err != nil {
		h.writeError(w, err)
		return
	}
	
	h.writeSuccess(w, template)
}

func (h *Handler) HandleListTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w)
		return
	}
	if r.Method != http.MethodGet {
		h.writeError(w, common.ErrInvalidRequest("method not allowed"))
		return
	}
	
	templates := h.service.ListTemplates()
	h.writeSuccess(w, templates)
}

func (h *Handler) HandleDeleteTemplate(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w)
		return
	}
	if r.Method != http.MethodDelete {
		h.writeError(w, common.ErrInvalidRequest("method not allowed"))
		return
	}
	
	id := r.URL.Query().Get("id")
	if id == "" {
		h.writeError(w, common.ErrInvalidRequest("id is required"))
		return
	}
	
	err := h.service.DeleteTemplate(id)
	if err != nil {
		h.writeError(w, err)
		return
	}
	
	h.writeSuccess(w, map[string]bool{"success": true})
}

func (h *Handler) HandleRecordActivity(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w)
		return
	}
	if r.Method != http.MethodPost {
		h.writeError(w, common.ErrInvalidRequest("method not allowed"))
		return
	}
	
	var req common.UserActivityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, common.ErrInvalidRequest(fmt.Sprintf("invalid request body: %v", err)))
		return
	}
	
	h.service.RecordActivity(req.UserID)
	h.writeSuccess(w, map[string]bool{"success": true})
}

func (h *Handler) HandleListFailedLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleOptions(w)
		return
	}
	if r.Method != http.MethodGet {
		h.writeError(w, common.ErrInvalidRequest("method not allowed"))
		return
	}
	
	logs := h.service.ListFailedLogs()
	h.writeSuccess(w, logs)
}
