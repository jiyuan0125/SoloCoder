package main

import (
	"charging-station/common"
	"charging-station/core"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Server struct {
	service *core.ChargingService
}

func NewServer() *Server {
	repo := core.NewMemoryRepository()
	
	station1 := &core.Station{
		ID:       "station-001",
		Name:     "中央公园充电站",
		Location: "北京市朝阳区中央公园",
		Chargers: []*core.Charger{
			{
				ID:           "charger-001",
				StationID:    "station-001",
				Code:         "A-01",
				Type:         core.FastCharger,
				Status:       core.StatusIdle,
				MeterReading: 1000.0,
			},
			{
				ID:           "charger-002",
				StationID:    "station-001",
				Code:         "A-02",
				Type:         core.SlowCharger,
				Status:       core.StatusIdle,
				MeterReading: 2000.0,
			},
			{
				ID:           "charger-003",
				StationID:    "station-001",
				Code:         "A-03",
				Type:         core.FastCharger,
				Status:       core.StatusIdle,
				MeterReading: 1500.0,
			},
		},
	}
	
	station2 := &core.Station{
		ID:       "station-002",
		Name:     "科技园充电站",
		Location: "北京市海淀区科技园",
		Chargers: []*core.Charger{
			{
				ID:           "charger-004",
				StationID:    "station-002",
				Code:         "B-01",
				Type:         core.SlowCharger,
				Status:       core.StatusIdle,
				MeterReading: 500.0,
			},
			{
				ID:           "charger-005",
				StationID:    "station-002",
				Code:         "B-02",
				Type:         core.FastCharger,
				Status:       core.StatusIdle,
				MeterReading: 800.0,
			},
		},
	}
	
	repo.AddStation(station1)
	repo.AddStation(station2)
	
	return &Server{
		service: core.NewChargingService(repo),
	}
}

func (s *Server) writeResponse(w http.ResponseWriter, statusCode int, data interface{}, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := &common.Response{
		Success: err == nil,
		Data:    data,
	}
	
	if err != nil {
		response.Error = err.Error()
	}
	
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleListStations(w http.ResponseWriter, r *http.Request) {
	s.service.CleanupExpired()
	
	stations := s.service.GetStations()
	
	stationDTOs := make([]*common.StationDTO, 0, len(stations))
	for _, station := range stations {
		chargerDTOs := make([]*common.ChargerDTO, 0, len(station.Chargers))
		for _, charger := range station.Chargers {
			chargerDTOs = append(chargerDTOs, &common.ChargerDTO{
				ID:           charger.ID,
				Code:         charger.Code,
				Type:         string(charger.Type),
				Status:       string(charger.Status),
				MeterReading: charger.MeterReading,
			})
		}
		
		stationDTOs = append(stationDTOs, &common.StationDTO{
			ID:       station.ID,
			Name:     station.Name,
			Location: station.Location,
			Chargers: chargerDTOs,
		})
	}
	
	s.writeResponse(w, http.StatusOK, &common.ListStationsResponse{
		Stations: stationDTOs,
	}, nil)
}

func (s *Server) handleCreateReservation(w http.ResponseWriter, r *http.Request) {
	s.service.CleanupExpired()
	
	var req common.CreateReservationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeResponse(w, http.StatusBadRequest, nil, err)
		return
	}
	
	reservation, err := s.service.CreateReservation(
		req.UserID,
		req.StationID,
		req.ChargerID,
		req.StartTime,
		req.EndTime,
	)
	
	if err != nil {
		s.writeResponse(w, http.StatusBadRequest, nil, err)
		return
	}
	
	s.writeResponse(w, http.StatusOK, &common.CreateReservationResponse{
		Reservation: &common.ReservationDTO{
			ID:        reservation.ID,
			UserID:    reservation.UserID,
			StationID: reservation.StationID,
			ChargerID: reservation.ChargerID,
			StartTime: reservation.StartTime,
			EndTime:   reservation.EndTime,
			Status:    string(reservation.Status),
		},
	}, nil)
}

func (s *Server) handleStartCharging(w http.ResponseWriter, r *http.Request) {
	s.service.CleanupExpired()
	
	var req common.StartChargingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeResponse(w, http.StatusBadRequest, nil, err)
		return
	}
	
	session, err := s.service.StartCharging(req.ReservationID)
	if err != nil {
		s.writeResponse(w, http.StatusBadRequest, nil, err)
		return
	}
	
	s.writeResponse(w, http.StatusOK, &common.StartChargingResponse{
		Session: convertSessionToDTO(session),
	}, nil)
}

func (s *Server) handlePauseCharging(w http.ResponseWriter, r *http.Request) {
	var req common.PauseChargingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeResponse(w, http.StatusBadRequest, nil, err)
		return
	}
	
	err := s.service.PauseCharging(req.SessionID)
	if err != nil {
		s.writeResponse(w, http.StatusBadRequest, nil, err)
		return
	}
	
	s.writeResponse(w, http.StatusOK, nil, nil)
}

func (s *Server) handleResumeCharging(w http.ResponseWriter, r *http.Request) {
	var req common.ResumeChargingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeResponse(w, http.StatusBadRequest, nil, err)
		return
	}
	
	err := s.service.ResumeCharging(req.SessionID)
	if err != nil {
		s.writeResponse(w, http.StatusBadRequest, nil, err)
		return
	}
	
	s.writeResponse(w, http.StatusOK, nil, nil)
}

func (s *Server) handleEndCharging(w http.ResponseWriter, r *http.Request) {
	var req common.EndChargingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeResponse(w, http.StatusBadRequest, nil, err)
		return
	}
	
	session, err := s.service.EndCharging(req.SessionID, req.EndMeter)
	if err != nil {
		s.writeResponse(w, http.StatusBadRequest, nil, err)
		return
	}
	
	s.writeResponse(w, http.StatusOK, &common.EndChargingResponse{
		Session: convertSessionToDTO(session),
	}, nil)
}

func (s *Server) handleListChargingRecords(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		s.writeResponse(w, http.StatusBadRequest, nil, fmt.Errorf("user_id is required"))
		return
	}
	
	records := s.service.GetChargingRecords(userID)
	
	recordDTOs := make([]*common.ChargingRecordDTO, 0, len(records))
	for _, record := range records {
		recordDTOs = append(recordDTOs, &common.ChargingRecordDTO{
			ID:           record.ID,
			ReservationID: record.ReservationID,
			UserID:       record.UserID,
			StationID:    record.StationID,
			ChargerID:    record.ChargerID,
			StartTime:    record.StartTime,
			EndTime:      record.EndTime,
			EnergyUsed:   record.EnergyUsed,
			Amount:       record.Amount,
			ChargingType: string(record.ChargingType),
			Duration:     record.Duration,
			Month:        record.Month,
		})
	}
	
	s.writeResponse(w, http.StatusOK, &common.ListChargingRecordsResponse{
		Records: recordDTOs,
	}, nil)
}

func (s *Server) handleGetMonthlySummary(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		s.writeResponse(w, http.StatusBadRequest, nil, fmt.Errorf("user_id is required"))
		return
	}
	
	summaries := s.service.GetMonthlySummary(userID)
	
	summaryDTOs := make([]*common.MonthlySummaryDTO, 0, len(summaries))
	for _, summary := range summaries {
		summaryDTOs = append(summaryDTOs, &common.MonthlySummaryDTO{
			Month:         summary.Month,
			TotalSessions: summary.TotalSessions,
			TotalEnergy:   summary.TotalEnergy,
			TotalAmount:   summary.TotalAmount,
		})
	}
	
	s.writeResponse(w, http.StatusOK, &common.GetMonthlySummaryResponse{
		Summaries: summaryDTOs,
	}, nil)
}

func convertSessionToDTO(session *core.ChargingSession) *common.SessionDTO {
	return &common.SessionDTO{
		ID:                 session.ID,
		ReservationID:      session.ReservationID,
		UserID:             session.UserID,
		StationID:          session.StationID,
		ChargerID:          session.ChargerID,
		StartTime:          session.StartTime,
		EndTime:            session.EndTime,
		StartMeter:         session.StartMeter,
		EndMeter:           session.EndMeter,
		EnergyUsed:         session.EnergyUsed,
		Amount:             session.Amount,
		Status:             string(session.Status),
		TotalPauseDuration: session.TotalPauseDuration,
	}
}

func main() {
	port := flag.String("port", "", "Server port")
	flag.Parse()
	
	if *port == "" {
		if envPort := os.Getenv("PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = "8906"
		}
	}
	
	portNum, err := strconv.Atoi(*port)
	if err != nil {
		fmt.Printf("Invalid port: %s\n", *port)
		os.Exit(1)
	}
	
	server := NewServer()
	
	mux := http.NewServeMux()
	
	mux.HandleFunc("/api/stations", server.handleListStations)
	mux.HandleFunc("/api/reservations", server.handleCreateReservation)
	mux.HandleFunc("/api/charging/start", server.handleStartCharging)
	mux.HandleFunc("/api/charging/pause", server.handlePauseCharging)
	mux.HandleFunc("/api/charging/resume", server.handleResumeCharging)
	mux.HandleFunc("/api/charging/end", server.handleEndCharging)
	mux.HandleFunc("/api/records", server.handleListChargingRecords)
	mux.HandleFunc("/api/summary", server.handleGetMonthlySummary)
	
	addr := fmt.Sprintf(":%d", portNum)
	fmt.Printf("Server starting on %s...\n", addr)
	
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			server.service.CleanupExpired()
		}
	}()
	
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}
