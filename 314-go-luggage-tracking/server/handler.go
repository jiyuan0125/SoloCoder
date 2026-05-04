package server

import (
	"encoding/json"
	"net/http"

	"luggage-tracking/common"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/scan", h.Scan)
	mux.HandleFunc("POST /api/backfill", h.Backfill)
	mux.HandleFunc("GET /api/luggage/{tag}", h.GetLuggage)
	mux.HandleFunc("GET /api/flight/{flight}", h.GetFlightLuggages)
	mux.HandleFunc("POST /api/flight/cancel", h.CancelFlight)
	mux.HandleFunc("GET /api/logs/operations", h.GetOperationLogs)
	mux.HandleFunc("GET /api/logs/audit", h.GetAuditLogs)
	mux.HandleFunc("GET /health", h.Health)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) Scan(w http.ResponseWriter, r *http.Request) {
	var req common.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.Scan(req); err != nil {
		switch err {
		case common.ErrInvalidLuggageTag:
			h.sendError(w, http.StatusBadRequest, err.Error())
		case common.ErrInvalidStage:
			h.sendError(w, http.StatusBadRequest, err.Error())
		case common.ErrLuggageNotFound:
			h.sendError(w, http.StatusNotFound, err.Error())
		case common.ErrLuggageAlreadyExists:
			h.sendError(w, http.StatusConflict, err.Error())
		case common.ErrStageAlreadyScanned:
			h.sendError(w, http.StatusConflict, err.Error())
		default:
			h.sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) Backfill(w http.ResponseWriter, r *http.Request) {
	var req common.BackfillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.Backfill(req); err != nil {
		switch err {
		case common.ErrInvalidLuggageTag:
			h.sendError(w, http.StatusBadRequest, err.Error())
		case common.ErrInvalidStage:
			h.sendError(w, http.StatusBadRequest, err.Error())
		case common.ErrLuggageNotFound:
			h.sendError(w, http.StatusNotFound, err.Error())
		case common.ErrLuggageAlreadyExists:
			h.sendError(w, http.StatusConflict, err.Error())
		case common.ErrStageAlreadyScanned:
			h.sendError(w, http.StatusConflict, err.Error())
		default:
			h.sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) GetLuggage(w http.ResponseWriter, r *http.Request) {
	tag := r.PathValue("tag")
	if tag == "" {
		h.sendError(w, http.StatusBadRequest, "luggage tag is required")
		return
	}

	luggage, err := h.service.GetLuggage(tag)
	if err != nil {
		switch err {
		case common.ErrInvalidLuggageTag:
			h.sendError(w, http.StatusBadRequest, err.Error())
		case common.ErrLuggageNotFound:
			h.sendError(w, http.StatusNotFound, err.Error())
		default:
			h.sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(luggage)
}

func (h *Handler) GetFlightLuggages(w http.ResponseWriter, r *http.Request) {
	flight := r.PathValue("flight")
	if flight == "" {
		h.sendError(w, http.StatusBadRequest, "flight number is required")
		return
	}

	date := r.URL.Query().Get("date")

	response, err := h.service.GetFlightLuggages(flight, date)
	if err != nil {
		switch err {
		case common.ErrInvalidFlightNumber:
			h.sendError(w, http.StatusBadRequest, err.Error())
		default:
			h.sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) CancelFlight(w http.ResponseWriter, r *http.Request) {
	var req common.FlightCancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.CancelFlight(req); err != nil {
		switch err {
		case common.ErrInvalidFlightNumber:
			h.sendError(w, http.StatusBadRequest, err.Error())
		default:
			h.sendError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) GetOperationLogs(w http.ResponseWriter, r *http.Request) {
	logs := h.service.GetOperationLogs()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (h *Handler) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	logs := h.service.GetAuditLogs()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (h *Handler) sendError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(common.NewErrorResponse(code, message))
}
