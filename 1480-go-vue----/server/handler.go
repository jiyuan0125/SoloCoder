package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"vehicle-inspection/common"
	"vehicle-inspection/core"
)

type Handler struct {
	service *core.InspectionService
}

func NewHandler(service *core.InspectionService) *Handler {
	return &Handler{service: service}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, common.NewErrorResponse(err.Error()))
}

func (h *Handler) RegisterVehicle(w http.ResponseWriter, r *http.Request) {
	var req common.RegisterVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	vehicle, err := h.service.RegisterVehicle(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.NewSuccessResponse(vehicle))
}

func (h *Handler) GetVehicleInfo(w http.ResponseWriter, r *http.Request) {
	plate := strings.TrimPrefix(r.URL.Path, "/api/vehicles/")
	if plate == "" {
		writeError(w, http.StatusBadRequest, http.ErrNoLocation)
		return
	}

	info, err := h.service.GetVehicleInfo(plate)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(info))
}

func (h *Handler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles := h.service.ListVehicles()
	writeJSON(w, http.StatusOK, common.NewSuccessResponse(vehicles))
}

func (h *Handler) CreateStation(w http.ResponseWriter, r *http.Request) {
	var req common.CreateStationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	station, err := h.service.CreateStation(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.NewSuccessResponse(station))
}

func (h *Handler) AddSchedule(w http.ResponseWriter, r *http.Request) {
	var req common.AddScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	schedule, err := h.service.AddSchedule(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.NewSuccessResponse(schedule))
}

func (h *Handler) ListStations(w http.ResponseWriter, r *http.Request) {
	stations := h.service.ListStations()
	writeJSON(w, http.StatusOK, common.NewSuccessResponse(stations))
}

func (h *Handler) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	var req common.CreateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	appointment, err := h.service.CreateAppointment(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.NewSuccessResponse(appointment))
}

func (h *Handler) GetAppointment(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/appointments/")
	if id == "" {
		writeError(w, http.StatusBadRequest, http.ErrNoLocation)
		return
	}

	appointment, err := h.service.GetAppointment(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(appointment))
}

func (h *Handler) ListAppointments(w http.ResponseWriter, r *http.Request) {
	appointments := h.service.ListAppointments()
	writeJSON(w, http.StatusOK, common.NewSuccessResponse(appointments))
}

func (h *Handler) StartInspection(w http.ResponseWriter, r *http.Request) {
	var req common.StartInspectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	process, err := h.service.StartInspection(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, common.NewSuccessResponse(process))
}

func (h *Handler) CompleteStep(w http.ResponseWriter, r *http.Request) {
	var req common.CompleteStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	process, err := h.service.CompleteStep(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(process))
}

func (h *Handler) StartRecheck(w http.ResponseWriter, r *http.Request) {
	var req common.StartRecheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	process, err := h.service.StartRecheck(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(process))
}

func (h *Handler) GetProcess(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/processes/")
	if id == "" {
		writeError(w, http.StatusBadRequest, http.ErrNoLocation)
		return
	}

	process, err := h.service.GetProcess(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, common.NewSuccessResponse(process))
}

func (h *Handler) ListProcesses(w http.ResponseWriter, r *http.Request) {
	processes := h.service.ListProcesses()
	writeJSON(w, http.StatusOK, common.NewSuccessResponse(processes))
}
