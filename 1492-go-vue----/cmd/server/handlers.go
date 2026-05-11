package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"renovation-management/internal/api"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, api.ErrorResponse{Error: err.Error()})
}

func (s *Server) handleCreateHouse(w http.ResponseWriter, r *http.Request) {
	var req api.CreateHouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	house, err := s.store.CreateHouse(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, house)
}

func (s *Server) handleListHouses(w http.ResponseWriter, r *http.Request) {
	houses := s.store.ListHouses()
	writeJSON(w, http.StatusOK, houses)
}

func (s *Server) handleGetHouse(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/houses/")
	if id == "" {
		writeError(w, http.StatusBadRequest, ErrMissingID)
		return
	}

	house, err := s.store.GetHouse(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, house)
}

func (s *Server) handleCreateDesign(w http.ResponseWriter, r *http.Request) {
	var req api.CreateDesignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	scheme, err := s.store.CreateDesignScheme(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, scheme)
}

func (s *Server) handleGetDesign(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/designs/")
	if id == "" {
		writeError(w, http.StatusBadRequest, ErrMissingID)
		return
	}

	scheme, err := s.store.GetDesignScheme(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, scheme)
}

func (s *Server) handleConfirmDesign(w http.ResponseWriter, r *http.Request) {
	var req api.ConfirmDesignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	scheme, err := s.store.ConfirmDesignScheme(req.SchemeID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, scheme)
}

func (s *Server) handleCreateQuotation(w http.ResponseWriter, r *http.Request) {
	var req api.CreateQuotationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	quotation, err := s.store.CreateQuotation(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, quotation)
}

func (s *Server) handleListQuotations(w http.ResponseWriter, r *http.Request) {
	quotations := s.store.ListQuotations()
	writeJSON(w, http.StatusOK, quotations)
}

func (s *Server) handleGetQuotation(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/quotations/")
	if id == "" {
		writeError(w, http.StatusBadRequest, ErrMissingID)
		return
	}

	quotation, err := s.store.GetQuotation(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, quotation)
}

func (s *Server) handleUpdateQuantity(w http.ResponseWriter, r *http.Request) {
	var req api.UpdateActualQuantityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	item, err := s.store.UpdateActualQuantity(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (s *Server) handleCreateChangeOrder(w http.ResponseWriter, r *http.Request) {
	var req api.CreateChangeOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	changeOrder, err := s.store.CreateChangeOrder(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, changeOrder)
}

func (s *Server) handleConfirmChangeOrder(w http.ResponseWriter, r *http.Request) {
	var req api.ConfirmChangeOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	changeOrder, err := s.store.ConfirmChangeOrder(req.ChangeOrderID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, changeOrder)
}

func (s *Server) handleCreatePhase(w http.ResponseWriter, r *http.Request) {
	var req api.CreatePhaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	phase, err := s.store.CreatePhase(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, phase)
}

func (s *Server) handleUpdateProgress(w http.ResponseWriter, r *http.Request) {
	var req api.UpdateProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	phase, err := s.store.UpdateProgress(&req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusOK, phase)
}

func (s *Server) handleGetPhase(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/phases/")
	if id == "" {
		writeError(w, http.StatusBadRequest, ErrMissingID)
		return
	}

	phase, err := s.store.GetPhase(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, phase)
}

func (s *Server) handleListDelayedPhases(w http.ResponseWriter, r *http.Request) {
	phases := s.store.ListDelayedPhases()
	writeJSON(w, http.StatusOK, phases)
}

func (s *Server) handleCalculateSettlement(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/settlements/calculate/")
	if id == "" {
		writeError(w, http.StatusBadRequest, ErrMissingID)
		return
	}

	settlement, err := s.store.CalculateSettlement(id)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	writeJSON(w, http.StatusCreated, settlement)
}

func (s *Server) handleGetSettlement(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/settlements/")
	if id == "" {
		writeError(w, http.StatusBadRequest, ErrMissingID)
		return
	}

	settlement, err := s.store.GetSettlement(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, settlement)
}
