package core

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	ErrStationNotFound       = errors.New("station not found")
	ErrChargerNotFound       = errors.New("charger not found")
	ErrReservationNotFound   = errors.New("reservation not found")
	ErrSessionNotFound       = errors.New("charging session not found")
	ErrChargerUnavailable    = errors.New("charger is unavailable")
	ErrTimeSlotInvalid       = errors.New("time slot is invalid")
	ErrTimeSlotOverlap       = errors.New("time slot overlaps with existing reservation")
	ErrNoActiveReservation   = errors.New("no active reservation found")
	ErrSessionAlreadyStarted = errors.New("charging session already started")
	ErrSessionNotCharging    = errors.New("charging session is not in charging state")
	ErrSessionNotPaused      = errors.New("charging session is not in paused state")
	ErrInvalidMeterReading   = errors.New("invalid meter reading")
)

type StationRepository interface {
	GetStations() []*Station
	GetStationByID(id string) *Station
	GetChargerByID(stationID, chargerID string) *Charger
	UpdateChargerStatus(stationID, chargerID string, status ChargerStatus) error
}

type ReservationRepository interface {
	CreateReservation(reservation *Reservation) error
	GetReservationsByCharger(chargerID string) []*Reservation
	GetActiveReservation(chargerID string, now time.Time) *Reservation
	GetReservationByID(id string) *Reservation
	UpdateReservationStatus(id string, status ReservationStatus) error
	CancelExpiredReservations(now time.Time) []string
}

type SessionRepository interface {
	CreateSession(session *ChargingSession) error
	GetActiveSession(chargerID string) *ChargingSession
	GetSessionByID(id string) *ChargingSession
	UpdateSession(session *ChargingSession) error
	EndPausedSessions(now time.Time) []string
	GetRecordsByUser(userID string) []*ChargingRecord
	GetMonthlySummary(userID string) []*MonthlySummary
}

type MemoryRepository struct {
	stations      map[string]*Station
	reservations  map[string]*Reservation
	sessions      map[string]*ChargingSession
	records       map[string]*ChargingRecord
	
	chargerToReservations map[string][]*Reservation
	chargerToSessions     map[string]*ChargingSession
	userToRecords         map[string][]*ChargingRecord
	
	mu sync.RWMutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		stations:              make(map[string]*Station),
		reservations:          make(map[string]*Reservation),
		sessions:              make(map[string]*ChargingSession),
		records:               make(map[string]*ChargingRecord),
		chargerToReservations: make(map[string][]*Reservation),
		chargerToSessions:     make(map[string]*ChargingSession),
		userToRecords:         make(map[string][]*ChargingRecord),
	}
}

func (r *MemoryRepository) AddStation(station *Station) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stations[station.ID] = station
}

func (r *MemoryRepository) GetStations() []*Station {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]*Station, 0, len(r.stations))
	for _, station := range r.stations {
		result = append(result, station)
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	
	return result
}

func (r *MemoryRepository) GetStationByID(id string) *Station {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stations[id]
}

func (r *MemoryRepository) GetChargerByID(stationID, chargerID string) *Charger {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	station := r.stations[stationID]
	if station == nil {
		return nil
	}
	
	for _, charger := range station.Chargers {
		if charger.ID == chargerID {
			return charger
		}
	}
	
	return nil
}

func (r *MemoryRepository) UpdateChargerStatus(stationID, chargerID string, status ChargerStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	station := r.stations[stationID]
	if station == nil {
		return ErrStationNotFound
	}
	
	for _, charger := range station.Chargers {
		if charger.ID == chargerID {
			charger.Status = status
			return nil
		}
	}
	
	return ErrChargerNotFound
}

func (r *MemoryRepository) CreateReservation(reservation *Reservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.reservations[reservation.ID] = reservation
	r.chargerToReservations[reservation.ChargerID] = append(
		r.chargerToReservations[reservation.ChargerID],
		reservation,
	)
	
	return nil
}

func (r *MemoryRepository) GetReservationsByCharger(chargerID string) []*Reservation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]*Reservation, len(r.chargerToReservations[chargerID]))
	copy(result, r.chargerToReservations[chargerID])
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime.Before(result[j].StartTime)
	})
	
	return result
}

func (r *MemoryRepository) GetActiveReservation(chargerID string, now time.Time) *Reservation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, reservation := range r.chargerToReservations[chargerID] {
		if reservation.Status != ReservationActive && 
		   reservation.Status != ReservationInProgress {
			continue
		}
		
		if reservation.StartTime.Before(now) && reservation.EndTime.After(now) {
			return reservation
		}
	}
	
	return nil
}

func (r *MemoryRepository) GetReservationByID(id string) *Reservation {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.reservations[id]
}

func (r *MemoryRepository) UpdateReservationStatus(id string, status ReservationStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	reservation := r.reservations[id]
	if reservation == nil {
		return ErrReservationNotFound
	}
	
	reservation.Status = status
	return nil
}

func (r *MemoryRepository) CancelExpiredReservations(now time.Time) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	var cancelledIDs []string
	
	for id, reservation := range r.reservations {
		if reservation.Status != ReservationActive {
			continue
		}
		
		if now.After(reservation.AutoCancelTime) {
			reservation.Status = ReservationCancelled
			cancelledIDs = append(cancelledIDs, id)
		}
	}
	
	return cancelledIDs
}

func (r *MemoryRepository) CreateSession(session *ChargingSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.sessions[session.ID] = session
	r.chargerToSessions[session.ChargerID] = session
	
	return nil
}

func (r *MemoryRepository) GetActiveSession(chargerID string) *ChargingSession {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	session := r.chargerToSessions[chargerID]
	if session != nil && session.Status != SessionCompleted {
		return session
	}
	
	return nil
}

func (r *MemoryRepository) GetSessionByID(id string) *ChargingSession {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sessions[id]
}

func (r *MemoryRepository) UpdateSession(session *ChargingSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	existing := r.sessions[session.ID]
	if existing == nil {
		return ErrSessionNotFound
	}
	
	*existing = *session
	
	if session.Status == SessionCompleted {
		delete(r.chargerToSessions, session.ChargerID)
		
		record := &ChargingRecord{
			ID:            session.ID,
			ReservationID: session.ReservationID,
			UserID:        session.UserID,
			StationID:     session.StationID,
			ChargerID:     session.ChargerID,
			StartTime:     session.StartTime,
			EndTime:       session.EndTime,
			EnergyUsed:    session.EnergyUsed,
			Amount:        session.Amount,
			Duration:      session.EndTime.Sub(session.StartTime) - session.TotalPauseDuration,
			Month:         session.EndTime.Format("2006-01"),
		}
		
		r.records[record.ID] = record
		r.userToRecords[record.UserID] = append(r.userToRecords[record.UserID], record)
	}
	
	return nil
}

func (r *MemoryRepository) EndPausedSessions(now time.Time) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	var endedIDs []string
	
	for chargerID, session := range r.chargerToSessions {
		if session.Status != SessionPaused {
			continue
		}
		
		if !session.PauseStartTime.IsZero() && IsPauseTimedOut(session.PauseStartTime, now) {
			delete(r.chargerToSessions, chargerID)
			session.Status = SessionCompleted
			session.EndTime = now
			endedIDs = append(endedIDs, session.ID)
		}
	}
	
	return endedIDs
}

func (r *MemoryRepository) GetRecordsByUser(userID string) []*ChargingRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	result := make([]*ChargingRecord, len(r.userToRecords[userID]))
	copy(result, r.userToRecords[userID])
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].StartTime.After(result[j].StartTime)
	})
	
	return result
}

func (r *MemoryRepository) GetMonthlySummary(userID string) []*MonthlySummary {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	records := r.userToRecords[userID]
	summaryMap := make(map[string]*MonthlySummary)
	
	for _, record := range records {
		summary := summaryMap[record.Month]
		if summary == nil {
			summary = &MonthlySummary{Month: record.Month}
			summaryMap[record.Month] = summary
		}
		
		summary.TotalSessions++
		summary.TotalEnergy += record.EnergyUsed
		summary.TotalAmount += record.Amount
	}
	
	result := make([]*MonthlySummary, 0, len(summaryMap))
	for _, summary := range summaryMap {
		summary.TotalEnergy = roundToTwoDecimals(summary.TotalEnergy)
		result = append(result, summary)
	}
	
	sort.Slice(result, func(i, j int) bool {
		return result[i].Month > result[j].Month
	})
	
	return result
}

type ChargingService struct {
	repo *MemoryRepository
	mu   sync.RWMutex
}

func NewChargingService(repo *MemoryRepository) *ChargingService {
	return &ChargingService{repo: repo}
}

func (s *ChargingService) GetStations() []*Station {
	return s.repo.GetStations()
}

func (s *ChargingService) GetStationByID(id string) *Station {
	return s.repo.GetStationByID(id)
}

func (s *ChargingService) GetChargerByID(stationID, chargerID string) *Charger {
	return s.repo.GetChargerByID(stationID, chargerID)
}

func (s *ChargingService) CreateReservation(userID, stationID, chargerID string, startTime, endTime time.Time) (*Reservation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if !ValidateTimeSlot(startTime, endTime) {
		return nil, ErrTimeSlotInvalid
	}
	
	station := s.repo.GetStationByID(stationID)
	if station == nil {
		return nil, ErrStationNotFound
	}
	
	charger := s.repo.GetChargerByID(stationID, chargerID)
	if charger == nil {
		return nil, ErrChargerNotFound
	}
	
	if charger.Status != StatusIdle && charger.Status != StatusReserved {
		return nil, ErrChargerUnavailable
	}
	
	existingReservations := s.repo.GetReservationsByCharger(chargerID)
	for _, existing := range existingReservations {
		if existing.Status != ReservationActive {
			continue
		}
		
		if DoTimeSlotsOverlap(startTime, endTime, existing.StartTime, existing.EndTime) {
			return nil, ErrTimeSlotOverlap
		}
	}
	
	reservation := &Reservation{
		ID:             generateID(),
		UserID:         userID,
		StationID:      stationID,
		ChargerID:      chargerID,
		StartTime:      startTime,
		EndTime:        endTime,
		Status:         ReservationActive,
		CreatedAt:      time.Now(),
		AutoCancelTime: startTime.Add(ReservationGraceMinutes * time.Minute),
	}
	
	if err := s.repo.CreateReservation(reservation); err != nil {
		return nil, err
	}
	
	now := time.Now()
	if startTime.Before(now) || startTime.Equal(now) {
		s.repo.UpdateChargerStatus(stationID, chargerID, StatusReserved)
	}
	
	return reservation, nil
}

func (s *ChargingService) GetReservationByID(id string) *Reservation {
	return s.repo.GetReservationByID(id)
}

func (s *ChargingService) StartCharging(reservationID string) (*ChargingSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	reservation := s.repo.GetReservationByID(reservationID)
	if reservation == nil {
		return nil, ErrReservationNotFound
	}
	
	if reservation.Status != ReservationActive {
		return nil, ErrNoActiveReservation
	}
	
	charger := s.repo.GetChargerByID(reservation.StationID, reservation.ChargerID)
	if charger == nil {
		return nil, ErrChargerNotFound
	}
	
	existingSession := s.repo.GetActiveSession(reservation.ChargerID)
	if existingSession != nil {
		return nil, ErrSessionAlreadyStarted
	}
	
	reservation.Status = ReservationInProgress
	s.repo.UpdateReservationStatus(reservationID, ReservationInProgress)
	
	session := &ChargingSession{
		ID:                  generateID(),
		ReservationID:       reservationID,
		UserID:              reservation.UserID,
		StationID:           reservation.StationID,
		ChargerID:           reservation.ChargerID,
		StartTime:           time.Now(),
		StartMeter:          charger.MeterReading,
		Status:              SessionCharging,
		TotalPauseDuration:  0,
	}
	
	s.repo.CreateSession(session)
	s.repo.UpdateChargerStatus(reservation.StationID, reservation.ChargerID, StatusCharging)
	
	return session, nil
}

func (s *ChargingService) PauseCharging(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	session := s.repo.GetSessionByID(sessionID)
	if session == nil {
		return ErrSessionNotFound
	}
	
	if session.Status != SessionCharging {
		return ErrSessionNotCharging
	}
	
	session.Status = SessionPaused
	session.PauseStartTime = time.Now()
	
	s.repo.UpdateSession(session)
	s.repo.UpdateChargerStatus(session.StationID, session.ChargerID, StatusPaused)
	
	return nil
}

func (s *ChargingService) ResumeCharging(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	session := s.repo.GetSessionByID(sessionID)
	if session == nil {
		return ErrSessionNotFound
	}
	
	if session.Status != SessionPaused {
		return ErrSessionNotPaused
	}
	
	if !session.PauseStartTime.IsZero() {
		session.TotalPauseDuration += time.Since(session.PauseStartTime)
		session.PauseStartTime = time.Time{}
	}
	
	session.Status = SessionCharging
	
	s.repo.UpdateSession(session)
	s.repo.UpdateChargerStatus(session.StationID, session.ChargerID, StatusCharging)
	
	return nil
}

func (s *ChargingService) EndCharging(sessionID string, endMeter float64) (*ChargingSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	session := s.repo.GetSessionByID(sessionID)
	if session == nil {
		return nil, ErrSessionNotFound
	}
	
	if session.Status == SessionCompleted {
		return nil, ErrSessionNotFound
	}
	
	reservation := s.repo.GetReservationByID(session.ReservationID)
	if reservation == nil {
		return nil, ErrReservationNotFound
	}
	
	charger := s.repo.GetChargerByID(session.StationID, session.ChargerID)
	if charger == nil {
		return nil, ErrChargerNotFound
	}
	
	if endMeter < session.StartMeter {
		return nil, ErrInvalidMeterReading
	}
	
	if session.Status == SessionPaused && !session.PauseStartTime.IsZero() {
		session.TotalPauseDuration += time.Since(session.PauseStartTime)
	}
	
	session.EndTime = time.Now()
	session.EndMeter = endMeter
	
	chargerType := charger.Type
	amount, energyUsed := CalculateSessionAmount(session, chargerType, reservation.EndTime)
	
	session.EnergyUsed = energyUsed
	session.Amount = amount
	session.Status = SessionCompleted
	
	charger.MeterReading = endMeter
	
	s.repo.UpdateSession(session)
	s.repo.UpdateReservationStatus(session.ReservationID, ReservationCompleted)
	s.repo.UpdateChargerStatus(session.StationID, session.ChargerID, StatusIdle)
	
	return session, nil
}

func (s *ChargingService) GetChargingRecords(userID string) []*ChargingRecord {
	return s.repo.GetRecordsByUser(userID)
}

func (s *ChargingService) GetMonthlySummary(userID string) []*MonthlySummary {
	return s.repo.GetMonthlySummary(userID)
}

func (s *ChargingService) CleanupExpired() {
	now := time.Now()
	
	cancelledReservations := s.repo.CancelExpiredReservations(now)
	for _, id := range cancelledReservations {
		reservation := s.repo.GetReservationByID(id)
		if reservation != nil {
			s.repo.UpdateChargerStatus(reservation.StationID, reservation.ChargerID, StatusIdle)
		}
	}
	
	endedSessions := s.repo.EndPausedSessions(now)
	for _, id := range endedSessions {
		session := s.repo.GetSessionByID(id)
		if session != nil {
			s.repo.UpdateChargerStatus(session.StationID, session.ChargerID, StatusIdle)
		}
	}
}

func (s *ChargingService) GetActiveSession(chargerID string) *ChargingSession {
	return s.repo.GetActiveSession(chargerID)
}

func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	now := time.Now().UnixNano()
	for i := range b {
		b[i] = charset[int(now)%len(charset)]
		now = now / 2
	}
	return string(b)
}
