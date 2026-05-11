package main

import (
	"encoding/json"
	"net/http"
	"parking-system/common"
	"parking-system/core"
)

type Handler struct {
	service *core.ParkingService
}

func NewHandler(service *core.ParkingService) *Handler {
	return &Handler{service: service}
}

func respondJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

func respondError(w http.ResponseWriter, code int, message string) {
	respondJSON(w, code, common.Response{
		Code:    code,
		Message: message,
	})
}

func respondSuccess(w http.ResponseWriter, data interface{}) {
	respondJSON(w, http.StatusOK, common.Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

func (h *Handler) CreateSpot(w http.ResponseWriter, r *http.Request) {
	var req common.CreateParkingSpotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.AddParkingSpot(req.ID, req.Area, req.Number, req.Type); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	spot, _ := h.service.GetParkingSpot(req.ID)
	respondSuccess(w, toSpotDTO(spot))
}

func (h *Handler) ListSpots(w http.ResponseWriter, r *http.Request) {
	area := r.URL.Query().Get("area")
	status := r.URL.Query().Get("status")

	spots := h.service.ListParkingSpots(area, status)
	dtos := make([]common.ParkingSpotDTO, 0, len(spots))
	for _, spot := range spots {
		dtos = append(dtos, toSpotDTO(spot))
	}
	respondSuccess(w, dtos)
}

func (h *Handler) UpdateSpotStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "spot id is required")
		return
	}

	var req common.UpdateParkingSpotStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.UpdateParkingSpotStatus(id, req.Status); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	spot, _ := h.service.GetParkingSpot(id)
	respondSuccess(w, toSpotDTO(spot))
}

func (h *Handler) GetGuidance(w http.ResponseWriter, r *http.Request) {
	guidance := h.service.GetGuidance()
	respondSuccess(w, guidance)
}

func (h *Handler) CheckIn(w http.ResponseWriter, r *http.Request) {
	var req common.CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.service.CheckIn(req.PlateNumber)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(w, resp)
}

func (h *Handler) CheckOut(w http.ResponseWriter, r *http.Request) {
	var req common.CheckOutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.service.CheckOut(req.PlateNumber)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(w, resp)
}

func (h *Handler) QueryFee(w http.ResponseWriter, r *http.Request) {
	plateNumber := r.URL.Query().Get("plate_number")
	if plateNumber == "" {
		respondError(w, http.StatusBadRequest, "plate_number is required")
		return
	}

	resp, err := h.service.QueryFee(plateNumber)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(w, resp)
}

func (h *Handler) CreateMonthlyCard(w http.ResponseWriter, r *http.Request) {
	var req common.CreateMonthlyCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	card, err := h.service.CreateMonthlyCard(req.CardType, req.OwnerName, req.PlateNumber, req.SpotID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondSuccess(w, toMonthlyCardDTO(card))
}

func (h *Handler) RenewMonthlyCard(w http.ResponseWriter, r *http.Request) {
	var req common.RenewMonthlyCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.RenewMonthlyCard(req.CardID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	card, _ := h.service.GetMonthlyCardByID(req.CardID)
	respondSuccess(w, toMonthlyCardDTO(card))
}

func (h *Handler) ListMonthlyCards(w http.ResponseWriter, r *http.Request) {
	plateNumber := r.URL.Query().Get("plate_number")
	activeOnly := r.URL.Query().Get("active_only") == "true"

	cards := h.service.ListMonthlyCards(plateNumber, activeOnly)
	dtos := make([]common.MonthlyCardDTO, 0, len(cards))
	for _, card := range cards {
		dtos = append(dtos, toMonthlyCardDTO(card))
	}
	respondSuccess(w, dtos)
}

func (h *Handler) ListParkingRecords(w http.ResponseWriter, r *http.Request) {
	plateNumber := r.URL.Query().Get("plate_number")
	activeOnly := r.URL.Query().Get("active_only") == "true"

	records := h.service.ListParkingRecords(plateNumber, activeOnly)
	dtos := make([]common.ParkingRecordDTO, 0, len(records))
	for _, r := range records {
		dtos = append(dtos, toParkingRecordDTO(r))
	}
	respondSuccess(w, dtos)
}

func toSpotDTO(spot *core.ParkingSpot) common.ParkingSpotDTO {
	return common.ParkingSpotDTO{
		ID:     spot.GetID(),
		Area:   spot.GetArea(),
		Number: spot.GetNumber(),
		Type:   spot.GetType(),
		Status: spot.GetStatus(),
	}
}

func toMonthlyCardDTO(card *core.MonthlyCard) common.MonthlyCardDTO {
	status := "active"
	if !card.GetIsActive() {
		status = "expired"
	}

	return common.MonthlyCardDTO{
		ID:           card.GetID(),
		CardType:     card.GetCardType(),
		OwnerName:    card.GetOwnerName(),
		PlateNumber:  card.GetPlateNumber(),
		SpotID:       card.GetSpotID(),
		StartDate:    card.GetStartDate(),
		EndDate:      card.GetEndDate(),
		GraceEndDate: card.GetGraceEndDate(),
		IsActive:     card.GetIsActive(),
		Status:       status,
	}
}

func toParkingRecordDTO(record *core.ParkingRecord) common.ParkingRecordDTO {
	dto := common.ParkingRecordDTO{
		PlateNumber:  record.GetPlateNumber(),
		CheckInTime:  record.GetCheckInTime(),
		SpotID:       record.GetSpotID(),
		Amount:       record.GetAmount(),
		IsActive:     record.GetIsActive(),
	}

	checkOut := record.GetCheckOutTime()
	if checkOut != nil {
		dto.CheckOutTime = *checkOut
	}

	return dto
}
