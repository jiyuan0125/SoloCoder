package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"usedcar/api"
	"usedcar/core"
)

type Server struct {
	store *core.Store
}

func NewServer() *Server {
	return &Server{
		store: core.NewStore(),
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err string) {
	writeJSON(w, status, api.ErrorResponse{
		Success: false,
		Error:   err,
	})
}

func (s *Server) handleCreateVehicle(w http.ResponseWriter, r *http.Request) {
	var req api.CreateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	vehicle, err := s.store.CreateVehicle(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, api.CreateVehicleResponse{
		Success: true,
		Vehicle: *vehicle,
	})
}

func (s *Server) handleListVehicles(w http.ResponseWriter, r *http.Request) {
	status := api.VehicleStatus(r.URL.Query().Get("status"))
	vehicles := s.store.ListVehicles(status)

	result := make([]api.Vehicle, len(vehicles))
	for i, v := range vehicles {
		result[i] = *v
	}

	writeJSON(w, http.StatusOK, api.ListVehiclesResponse{
		Success:  true,
		Vehicles: result,
	})
}

func (s *Server) handleGetVehicle(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/vehicles/")
	if id == "" {
		writeError(w, http.StatusBadRequest, "vehicle id required")
		return
	}

	vehicle, ok := s.store.GetVehicle(id)
	if !ok {
		writeError(w, http.StatusNotFound, "vehicle not found")
		return
	}

	writeJSON(w, http.StatusOK, api.GetVehicleResponse{
		Success: true,
		Vehicle: *vehicle,
	})
}

func (s *Server) handleEvaluateVehicle(w http.ResponseWriter, r *http.Request) {
	var req api.EvaluateVehicleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	evaluation, err := s.store.EvaluateVehicle(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, api.EvaluateVehicleResponse{
		Success:    true,
		Evaluation: *evaluation,
	})
}

func (s *Server) handleListEvaluations(w http.ResponseWriter, r *http.Request) {
	vehicleID := r.URL.Query().Get("vehicle_id")
	if vehicleID == "" {
		writeError(w, http.StatusBadRequest, "vehicle_id required")
		return
	}

	evaluations, ok := s.store.ListEvaluations(vehicleID)
	if !ok {
		writeError(w, http.StatusNotFound, "vehicle not found")
		return
	}

	result := make([]api.Evaluation, len(evaluations))
	for i, e := range evaluations {
		result[i] = *e
	}

	writeJSON(w, http.StatusOK, api.ListEvaluationsResponse{
		Success:     true,
		Evaluations: result,
	})
}

func (s *Server) handlePayDeposit(w http.ResponseWriter, r *http.Request) {
	var req api.PayDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	deposit, err := s.store.PayDeposit(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, api.PayDepositResponse{
		Success: true,
		Deposit: *deposit,
	})
}

func (s *Server) handlePayFull(w http.ResponseWriter, r *http.Request) {
	var req api.PayFullRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	transfer, err := s.store.PayFull(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, api.PayFullResponse{
		Success:        true,
		TransferRecord: *transfer,
	})
}
