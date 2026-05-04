package handler

import (
	"encoding/json"
	"hospital-bed/internal/server/service"
	"hospital-bed/shared/protocol"
	"net/http"
	"strings"
)

type APIHandler struct {
	service *service.BedService
}

func NewAPIHandler(svc *service.BedService) *APIHandler {
	return &APIHandler{service: svc}
}

func jsonResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func errorResponse(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, protocol.Response{
		Success: false,
		Message: message,
	})
}

func decodeBody(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func (h *APIHandler) ConfigureDepartment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req protocol.DepartmentConfigRequest
	if err := decodeBody(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}

	if err := h.service.ConfigureDepartment(req); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, protocol.Response{Success: true})
}

func (h *APIHandler) AdmitPatient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req protocol.AdmissionRequest
	if err := decodeBody(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}

	resp, err := h.service.AdmitPatient(req)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (h *APIHandler) DischargePatient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req protocol.DischargeRequest
	if err := decodeBody(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}

	resp, err := h.service.DischargePatient(req)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (h *APIHandler) GetDepartmentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	deptName := r.URL.Query().Get("department")
	if deptName == "" {
		errorResponse(w, http.StatusBadRequest, "缺少科室参数")
		return
	}

	resp, err := h.service.GetDepartmentStatus(deptName)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (h *APIHandler) GetPatientBed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	patientID := r.URL.Query().Get("patient_id")
	if patientID == "" {
		errorResponse(w, http.StatusBadRequest, "缺少患者ID参数")
		return
	}

	req := protocol.PatientBedRequest{PatientID: patientID}
	resp, err := h.service.GetPatientBed(req)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

func (h *APIHandler) GetQueueStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	deptName := r.URL.Query().Get("department")
	if deptName == "" {
		errorResponse(w, http.StatusBadRequest, "缺少科室参数")
		return
	}

	resp, err := h.service.GetQueueStatus(deptName)
	if err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, resp)
}

type TransferRequest struct {
	PatientID     string          `json:"patient_id"`
	TargetDept    string          `json:"target_department"`
	TargetWard    *string         `json:"target_ward,omitempty"`
	PreferredType protocol.BedType `json:"preferred_type"`
}

func (h *APIHandler) TransferPatient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		errorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req TransferRequest
	if err := decodeBody(r, &req); err != nil {
		errorResponse(w, http.StatusBadRequest, "请求体解析失败: "+err.Error())
		return
	}

	if err := h.service.TransferPatient(req.PatientID, req.TargetDept, req.TargetWard, req.PreferredType); err != nil {
		errorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, protocol.Response{Success: true})
}

func (h *APIHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/configure", h.ConfigureDepartment)
	mux.HandleFunc("/api/admit", h.AdmitPatient)
	mux.HandleFunc("/api/discharge", h.DischargePatient)
	mux.HandleFunc("/api/department/status", h.GetDepartmentStatus)
	mux.HandleFunc("/api/patient/bed", h.GetPatientBed)
	mux.HandleFunc("/api/queue/status", h.GetQueueStatus)
	mux.HandleFunc("/api/transfer", h.TransferPatient)
}

func stripPrefix(prefix string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
		h(w, r)
	}
}
