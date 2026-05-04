package handler

import (
	"delivery-tracking/models"
	"delivery-tracking/service"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type DeliveryHandler struct {
	deliveryService service.DeliveryService
}

func NewDeliveryHandler(ds service.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{deliveryService: ds}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

func (h *DeliveryHandler) ReportStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.DeliveryStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validateRequest(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.deliveryService.ReportStatus(&req); err != nil {
		if errors.Is(err, service.ErrInvalidStatusName) {
			h.respondWithError(w, http.StatusBadRequest, err.Error())
		} else if errors.Is(err, service.ErrInvalidLongitude) || errors.Is(err, service.ErrInvalidLatitude) {
			h.respondWithError(w, http.StatusBadRequest, err.Error())
		} else if errors.Is(err, service.ErrDuplicateStatus) {
			h.respondWithSuccess(w, "Duplicate status, ignored")
		} else {
			h.respondWithError(w, http.StatusInternalServerError, "Failed to report status")
		}
		return
	}

	h.respondWithSuccess(w, "Status reported successfully")
}

func (h *DeliveryHandler) ReportBatchStatuses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.DeliveryStatusBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if len(req.Statuses) == 0 {
		h.respondWithError(w, http.StatusBadRequest, "No statuses provided")
		return
	}

	errs := h.deliveryService.ReportBatchStatuses(&req)
	
	if len(errs) > 0 {
		var errorMessages []string
		for _, err := range errs {
			errorMessages = append(errorMessages, err.Error())
		}
		h.respondWithError(w, http.StatusBadRequest, strings.Join(errorMessages, "; "))
		return
	}

	h.respondWithSuccess(w, "All statuses reported successfully")
}

func (h *DeliveryHandler) GetTrajectory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orderNo := r.URL.Query().Get("order_no")
	if orderNo == "" {
		h.respondWithError(w, http.StatusBadRequest, "order_no is required")
		return
	}

	response, err := h.deliveryService.GetTrajectory(orderNo)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get trajectory")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *DeliveryHandler) validateRequest(req *models.DeliveryStatusRequest) error {
	if req.OrderNo == "" {
		return errors.New("order_no is required")
	}
	if req.StatusName == "" {
		return errors.New("status_name is required")
	}
	if req.OccurredAt <= 0 {
		return errors.New("occurred_at must be a positive timestamp")
	}
	return nil
}

func (h *DeliveryHandler) respondWithError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (h *DeliveryHandler) respondWithSuccess(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Message: message})
}
