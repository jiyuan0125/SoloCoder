package main

import (
	"archivesystem/common"
	"archivesystem/core"
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct {
	service *core.ArchiveService
}

func writeJSON(w http.ResponseWriter, resp common.Response, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CreateArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, common.Response{Success: false, Message: "只支持POST请求"}, http.StatusMethodNotAllowed)
		return
	}

	var req common.CreateArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, common.Response{Success: false, Message: "请求解析失败"}, http.StatusBadRequest)
		return
	}

	id, err := h.service.CreateArchive(req)
	if err != nil {
		writeJSON(w, common.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}

	writeJSON(w, common.Response{Success: true, Data: common.CreateArchiveResponse{ArchiveID: id}}, http.StatusOK)
}

func (h *Handler) ListArchives(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, common.Response{Success: false, Message: "只支持POST请求"}, http.StatusMethodNotAllowed)
		return
	}

	var req common.ListArchivesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, common.Response{Success: false, Message: "请求解析失败"}, http.StatusBadRequest)
		return
	}

	archives := h.service.ListArchives(req)
	writeJSON(w, common.Response{Success: true, Data: archives}, http.StatusOK)
}

func (h *Handler) GetArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, common.Response{Success: false, Message: "只支持POST请求"}, http.StatusMethodNotAllowed)
		return
	}

	var req common.GetArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, common.Response{Success: false, Message: "请求解析失败"}, http.StatusBadRequest)
		return
	}

	arch, err := h.service.GetArchive(req)
	if err != nil {
		writeJSON(w, common.Response{Success: false, Message: err.Error()}, http.StatusNotFound)
		return
	}

	writeJSON(w, common.Response{Success: true, Data: arch}, http.StatusOK)
}

func (h *Handler) ExportArchives(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, common.Response{Success: false, Message: "只支持POST请求"}, http.StatusMethodNotAllowed)
		return
	}

	var req common.ListArchivesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, common.Response{Success: false, Message: "请求解析失败"}, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=archives.csv")

	if err := h.service.ExportArchives(w, req.UserRole); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) ApplyBorrow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, common.Response{Success: false, Message: "只支持POST请求"}, http.StatusMethodNotAllowed)
		return
	}

	var req common.ApplyBorrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, common.Response{Success: false, Message: "请求解析失败"}, http.StatusBadRequest)
		return
	}

	id, err := h.service.ApplyBorrow(req)
	if err != nil {
		writeJSON(w, common.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}

	writeJSON(w, common.Response{Success: true, Data: map[string]int64{"borrow_id": id}}, http.StatusOK)
}

func (h *Handler) ApproveBorrow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, common.Response{Success: false, Message: "只支持POST请求"}, http.StatusMethodNotAllowed)
		return
	}

	var req common.ApproveBorrowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, common.Response{Success: false, Message: "请求解析失败"}, http.StatusBadRequest)
		return
	}

	if err := h.service.ApproveBorrow(req); err != nil {
		writeJSON(w, common.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}

	writeJSON(w, common.Response{Success: true, Message: "审批成功"}, http.StatusOK)
}

func (h *Handler) ConfirmBorrow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, common.Response{Success: false, Message: "只支持POST请求"}, http.StatusMethodNotAllowed)
		return
	}

	var req map[string]int64
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, common.Response{Success: false, Message: "请求解析失败"}, http.StatusBadRequest)
		return
	}

	if err := h.service.ConfirmBorrow(req["borrow_id"]); err != nil {
		writeJSON(w, common.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}

	writeJSON(w, common.Response{Success: true, Message: "确认借阅成功"}, http.StatusOK)
}

func (h *Handler) ReturnArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, common.Response{Success: false, Message: "只支持POST请求"}, http.StatusMethodNotAllowed)
		return
	}

	var req common.ReturnArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, common.Response{Success: false, Message: "请求解析失败"}, http.StatusBadRequest)
		return
	}

	if err := h.service.ReturnArchive(req); err != nil {
		writeJSON(w, common.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}

	writeJSON(w, common.Response{Success: true, Message: "归还成功"}, http.StatusOK)
}

func (h *Handler) ListBorrowRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeJSON(w, common.Response{Success: false, Message: "不支持该请求方法"}, http.StatusMethodNotAllowed)
		return
	}

	records := h.service.ListBorrowRecords()
	writeJSON(w, common.Response{Success: true, Data: records}, http.StatusOK)
}

func (h *Handler) GetOverdueReminders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, common.Response{Success: false, Message: "只支持POST请求"}, http.StatusMethodNotAllowed)
		return
	}

	var req common.ListPendingDestroyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.CurrentDate = ""
	}

	reminders := h.service.GetOverdueReminders(req.CurrentDate)
	writeJSON(w, common.Response{Success: true, Data: reminders}, http.StatusOK)
}

func (h *Handler) ExportBorrowRecords(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=borrow_records.csv")

	if err := h.service.ExportBorrowRecords(w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) ListPendingDestroy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, common.Response{Success: false, Message: "只支持POST请求"}, http.StatusMethodNotAllowed)
		return
	}

	var req common.ListPendingDestroyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.CurrentDate = ""
	}

	archives := h.service.ListPendingDestroy(req)
	writeJSON(w, common.Response{Success: true, Data: archives}, http.StatusOK)
}

func (h *Handler) DestroyArchive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, common.Response{Success: false, Message: "只支持POST请求"}, http.StatusMethodNotAllowed)
		return
	}

	var req common.DestroyArchiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, common.Response{Success: false, Message: "请求解析失败"}, http.StatusBadRequest)
		return
	}

	if err := h.service.DestroyArchive(req); err != nil {
		writeJSON(w, common.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}

	writeJSON(w, common.Response{Success: true, Message: "销毁成功"}, http.StatusOK)
}

func (h *Handler) ListDestroyRecords(w http.ResponseWriter, r *http.Request) {
	records := h.service.ListDestroyRecords()
	writeJSON(w, common.Response{Success: true, Data: records}, http.StatusOK)
}

func parseParamInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
