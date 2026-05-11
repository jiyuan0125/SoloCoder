package main

import (
	"encoding/json"
	"net/http"

	"smart-park/common"
)

func (h *Handler) handleAccessPoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.CreateAccessPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	ap, err := h.accessService.CreateAccessPoint(req.ID, req.Name, req.AreaID, req.BuildingID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: ap})
}

func (h *Handler) handleAccessPointsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}
	points := h.accessService.GetAllAccessPoints()
	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: points})
}

func (h *Handler) handleAccessRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.CreateAccessRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	rule, err := h.accessService.CreateAccessRule(req.AccessPointID, req.AllowedDepartments, req.StartTime, req.EndTime)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: rule})
}

func (h *Handler) handleBatchAreaRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.BatchAreaRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	rules, err := h.accessService.BatchSetRulesByArea(req.AreaID, req.AllowedDepartments, req.StartTime, req.EndTime)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: rules})
}

func (h *Handler) handleAccessVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.AccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	record, err := h.accessService.ProcessAccess(req.EmployeeID, req.AccessPointID, req.AccessType)
	if err != nil {
		writeJSON(w, http.StatusForbidden, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: record})
}

func (h *Handler) handleAccessRecords(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}
	records := h.accessService.GetAllRecords()
	writeJSON(w, http.StatusOK, common.Response{Success: true, Data: records})
}

func (h *Handler) handleFixAccessPoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, common.Response{Success: false, Message: "method not allowed"})
		return
	}

	var req common.FixAccessPointRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: "invalid request body"})
		return
	}

	if err := h.accessService.FixAccessPoint(req.AccessPointID); err != nil {
		writeJSON(w, http.StatusBadRequest, common.Response{Success: false, Message: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, common.Response{Success: true, Message: "access point fixed"})
}

func writeJSON(w http.ResponseWriter, status int, resp common.Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
