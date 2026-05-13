package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"onboard-flow/internal/model"
	"onboard-flow/internal/service"
)

type Handler struct {
	service *service.FlowService
}

func NewHandler(s *service.FlowService) *Handler {
	return &Handler{service: s}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func parseEmployeeID(r *http.Request) (int64, error) {
	idStr := r.PathValue("id")
	if idStr == "" {
		return 0, errors.New("missing employee id")
	}
	return strconv.ParseInt(idStr, 10, 64)
}

type CreateEmployeeRequest struct {
	Name        string  `json:"name"`
	Email       string  `json:"email"`
	Department  string  `json:"department"`
	HireDate    string  `json:"hire_date"`
	TotalAmount float64 `json:"total_amount"`
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid hire_date format, use YYYY-MM-DD")
		return
	}

	emp, err := h.service.CreateEmployee(req.Name, req.Email, req.Department, hireDate, req.TotalAmount)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, emp)
}

func (h *Handler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := parseEmployeeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	emp, steps, err := h.service.GetEmployeeFlow(id)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"employee": emp,
		"steps":    steps,
	})
}

type StepRequest struct {
	StepName model.StepName `json:"step_name"`
}

func (h *Handler) StartStep(w http.ResponseWriter, r *http.Request) {
	id, err := parseEmployeeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req StepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	emp, err := h.service.GetEmployee(id)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if model.StepOrder[req.StepName] > model.StepOrder[emp.CurrentStep] {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":               "cannot jump steps",
			"current_step":        emp.CurrentStep,
			"should_complete":     emp.CurrentStep,
			"requested_step":      req.StepName,
		})
		return
	}

	if emp.Status == model.StatusTerminated {
		writeError(w, http.StatusBadRequest, "process is terminated")
		return
	}

	err = h.service.StartStep(id, req.StepName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

func (h *Handler) CompleteStep(w http.ResponseWriter, r *http.Request) {
	id, err := parseEmployeeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		StepName model.StepName `json:"step_name"`
		Passed   *bool          `json:"passed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	emp, err := h.service.GetEmployee(id)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if model.StepOrder[req.StepName] != model.StepOrder[emp.CurrentStep] {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"error":           "cannot jump steps",
			"current_step":    emp.CurrentStep,
			"should_complete": emp.CurrentStep,
			"requested_step":  req.StepName,
		})
		return
	}

	if emp.Status == model.StatusTerminated {
		writeError(w, http.StatusBadRequest, "process is terminated")
		return
	}

	err = h.service.CompleteStep(id, req.StepName, req.Passed)
	if err != nil {
		if errors.Is(err, service.ErrBackgroundCheckFail) {
			writeError(w, http.StatusBadRequest, "background check failed, process terminated")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

type BackgroundCheckRequest struct {
	Passed bool   `json:"passed"`
	Note   string `json:"note"`
}

func (h *Handler) ProcessBackgroundCheck(w http.ResponseWriter, r *http.Request) {
	id, err := parseEmployeeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req BackgroundCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.ProcessBackgroundCheck(id, req.Passed, req.Note)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		if errors.Is(err, service.ErrBackgroundCheckFail) {
			writeError(w, http.StatusBadRequest, "background check failed, process terminated")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "processed"})
}

type AccountSetupRequest struct {
	Email  *bool `json:"email_done"`
	IM     *bool `json:"im_done"`
	VPN    *bool `json:"vpn_done"`
	DevEnv *bool `json:"dev_env_done"`
}

func (h *Handler) UpdateAccountSetup(w http.ResponseWriter, r *http.Request) {
	id, err := parseEmployeeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req AccountSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.UpdateAccountSetup(id, req.Email, req.IM, req.VPN, req.DevEnv)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

type CreateMentorRequest struct {
	Name       string `json:"name"`
	EmployeeID string `json:"employee_id"`
	Department string `json:"department"`
}

func (h *Handler) CreateMentor(w http.ResponseWriter, r *http.Request) {
	var req CreateMentorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	mentor, err := h.service.CreateMentor(req.Name, req.EmployeeID, req.Department)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, mentor)
}

type AssignMentorRequest struct {
	MentorID int64 `json:"mentor_id"`
}

func (h *Handler) AssignMentor(w http.ResponseWriter, r *http.Request) {
	id, err := parseEmployeeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req AssignMentorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.AssignMentor(id, req.MentorID)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		if errors.Is(err, service.ErrMentorCapacityFull) {
			writeError(w, http.StatusConflict, "mentor has reached maximum trainees (3)")
			return
		}
		if errors.Is(err, service.ErrStepNotReady) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "assigned"})
}

type ExamRequest struct {
	ExamType string `json:"exam_type"`
	Score    int    `json:"score"`
	MaxScore int    `json:"max_score"`
}

func (h *Handler) SubmitExam(w http.ResponseWriter, r *http.Request) {
	id, err := parseEmployeeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req ExamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.SubmitExam(id, req.ExamType, req.Score, req.MaxScore)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		if errors.Is(err, service.ErrExamMaxAttempts) {
			writeError(w, http.StatusBadRequest, "max exam attempts reached, process terminated")
			return
		}
		if errors.Is(err, service.ErrStepNotReady) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "submitted"})
}

type ReviewRequest struct {
	Score   int    `json:"score"`
	Comment string `json:"comment"`
}

func (h *Handler) SubmitMentorReview(w http.ResponseWriter, r *http.Request) {
	id, err := parseEmployeeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.SubmitMentorReview(id, req.Score, req.Comment)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		if errors.Is(err, service.ErrStepNotReady) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "submitted"})
}

func (h *Handler) SubmitManagerReview(w http.ResponseWriter, r *http.Request) {
	id, err := parseEmployeeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.SubmitManagerReview(id, req.Score, req.Comment)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		if errors.Is(err, service.ErrStepNotReady) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "submitted"})
}

type DelayRequest struct {
	StepName model.StepName `json:"step_name"`
	Reason   string         `json:"reason"`
}

func (h *Handler) RecordDelay(w http.ResponseWriter, r *http.Request) {
	id, err := parseEmployeeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req DelayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.RecordDelay(id, req.StepName, req.Reason)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "recorded"})
}

func (h *Handler) GetDelayRecords(w http.ResponseWriter, r *http.Request) {
	id, err := parseEmployeeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	records, err := h.service.GetDelayRecords(id)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, records)
}

type UpdateAmountRequest struct {
	NewTotal float64 `json:"new_total"`
}

func (h *Handler) UpdateTotalAmount(w http.ResponseWriter, r *http.Request) {
	id, err := parseEmployeeID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req UpdateAmountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err = h.service.UpdateTotalAmount(id, req.NewTotal)
	if err != nil {
		if errors.Is(err, service.ErrEmployeeNotFound) {
			writeError(w, http.StatusNotFound, "employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
