package server

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"seating-arrangement/pkg/protocol"

	"github.com/google/uuid"
)

const (
	LockTimeout = 10 * time.Minute
)

type SessionManager struct {
	sessions    map[string]*protocol.Session
	venueManager *VenueManager
	mu          sync.RWMutex
}

func NewSessionManager(vm *VenueManager) *SessionManager {
	return &SessionManager{
		sessions:     make(map[string]*protocol.Session),
		venueManager: vm,
	}
}

func (sm *SessionManager) CreateSession(req *protocol.CreateSessionRequest) (*protocol.Session, error) {
	if req.Name == "" {
		return nil, errors.New("场次名称不能为空")
	}
	if len(req.Name) > 50 {
		return nil, errors.New("场次名称不能超过50个字")
	}

	venue, exists := sm.venueManager.GetVenue(req.VenueID)
	if !exists {
		return nil, errors.New("场馆不存在")
	}

	session := &protocol.Session{
		ID:         uuid.New().String(),
		VenueID:    req.VenueID,
		Name:       req.Name,
		SeatStates: make(map[string]*protocol.SeatState),
		CreatedAt:  time.Now(),
		TotalSeats: 0,
	}

	totalSeats := 0
	for sectionName, section := range venue.Sections {
		for row := 1; row <= section.Rows; row++ {
			rowLetter := getRowLetter(row)
			for seatNum := 1; seatNum <= section.SeatsPerRow; seatNum++ {
				seatID := fmt.Sprintf("%s-%s-%d", sectionName, rowLetter, seatNum)
				
				status := protocol.SeatStatusAvailable
				if section.Aisles[seatID] {
					status = protocol.SeatStatusAisle
				}

				seatState := &protocol.SeatState{
					SeatID:      seatID,
					SectionName: sectionName,
					Row:         rowLetter,
					Number:      seatNum,
					Status:      status,
					LockedAt:    nil,
					OrderID:     "",
				}
				session.SeatStates[seatID] = seatState
				
				if status != protocol.SeatStatusAisle {
					totalSeats++
				}
			}
		}
	}
	session.TotalSeats = totalSeats

	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[session.ID] = session

	return session, nil
}

func (sm *SessionManager) GetSession(id string) (*protocol.Session, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	session, exists := sm.sessions[id]
	return session, exists
}

func (sm *SessionManager) ListSessions() []*protocol.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	sessions := make([]*protocol.Session, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

func (sm *SessionManager) LockSeat(sessionID, sectionName, seatID string) (*protocol.Order, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, errors.New("场次不存在")
	}

	seatState, exists := session.SeatStates[seatID]
	if !exists {
		return nil, errors.New("座位不存在")
	}

	if seatState.SectionName != sectionName {
		return nil, errors.New("区域与座位不匹配")
	}

	if seatState.Status == protocol.SeatStatusAisle {
		return nil, errors.New("座位是过道，不能选择")
	}

	now := time.Now()
	if seatState.Status == protocol.SeatStatusLocked && seatState.LockedAt != nil {
		if now.Sub(*seatState.LockedAt) < LockTimeout {
			return nil, errors.New("座位已被选中")
		}
	}

	if seatState.Status == protocol.SeatStatusSold {
		return nil, errors.New("座位已售出")
	}

	lockedAt := now
	order := &protocol.Order{
		ID:          uuid.New().String(),
		SessionID:   sessionID,
		SectionName: sectionName,
		SeatID:      seatID,
		Status:      protocol.OrderStatusCreated,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	seatState.Status = protocol.SeatStatusLocked
	seatState.LockedAt = &lockedAt
	seatState.OrderID = order.ID

	return order, nil
}

func (sm *SessionManager) ReleaseSeat(sessionID, seatID string, isAdmin bool) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return errors.New("场次不存在")
	}

	seatState, exists := session.SeatStates[seatID]
	if !exists {
		return errors.New("座位不存在")
	}

	if seatState.Status == protocol.SeatStatusSold {
		if isAdmin {
			return errors.New("管理员不能释放已售出的座位")
		}
		return errors.New("座位已售出，无法释放")
	}

	if seatState.Status == protocol.SeatStatusLocked {
		seatState.Status = protocol.SeatStatusAvailable
		seatState.LockedAt = nil
		seatState.OrderID = ""
		return nil
	}

	return errors.New("座位当前未被锁定")
}

func (sm *SessionManager) GetSessionStats(sessionID string) (*protocol.SessionStats, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, errors.New("场次不存在")
	}

	stats := &protocol.SessionStats{
		SessionID:   session.ID,
		SessionName: session.Name,
		TotalSeats:  session.TotalSeats,
		SectionStats: make(map[string]*protocol.SectionStat),
	}

	venue, exists := sm.venueManager.GetVenue(session.VenueID)
	if exists {
		for sectionName := range venue.Sections {
			stats.SectionStats[sectionName] = &protocol.SectionStat{
				SectionName: sectionName,
			}
		}
	}

	for _, seatState := range session.SeatStates {
		sectionStat, exists := stats.SectionStats[seatState.SectionName]
		if !exists {
			continue
		}

		sectionStat.TotalSeats++

		switch seatState.Status {
		case protocol.SeatStatusAvailable:
			sectionStat.AvailableSeats++
		case protocol.SeatStatusLocked:
			sectionStat.LockedSeats++
		case protocol.SeatStatusSold:
			sectionStat.SoldSeats++
		}
	}

	return stats, nil
}

func (sm *SessionManager) GetAvailableSeats(sessionID, sectionName string) ([]*protocol.SeatDetail, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, errors.New("场次不存在")
	}

	var availableSeats []*protocol.SeatDetail
	for _, seatState := range session.SeatStates {
		if seatState.SectionName != sectionName {
			continue
		}

		if seatState.Status == protocol.SeatStatusAisle {
			continue
		}

		seatDetail := &protocol.SeatDetail{
			SeatID:      seatState.SeatID,
			SectionName: seatState.SectionName,
			Row:         seatState.Row,
			Number:      seatState.Number,
			Status:      seatState.Status,
		}
		availableSeats = append(availableSeats, seatDetail)
	}

	return availableSeats, nil
}

func (sm *SessionManager) GetSoldSeats(sessionID string) ([]*protocol.SeatDetail, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, errors.New("场次不存在")
	}

	var soldSeats []*protocol.SeatDetail
	for _, seatState := range session.SeatStates {
		if seatState.Status == protocol.SeatStatusSold {
			seatDetail := &protocol.SeatDetail{
				SeatID:      seatState.SeatID,
				SectionName: seatState.SectionName,
				Row:         seatState.Row,
				Number:      seatState.Number,
				Status:      seatState.Status,
			}
			soldSeats = append(soldSeats, seatDetail)
		}
	}

	return soldSeats, nil
}

func (sm *SessionManager) CheckExpiredLocks() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for _, session := range sm.sessions {
		for _, seatState := range session.SeatStates {
			if seatState.Status == protocol.SeatStatusLocked && seatState.LockedAt != nil {
				if now.Sub(*seatState.LockedAt) >= LockTimeout {
					seatState.Status = protocol.SeatStatusAvailable
					seatState.LockedAt = nil
					seatState.OrderID = ""
				}
			}
		}
	}
}

func (sm *SessionManager) SetSessions(sessions map[string]*protocol.Session) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions = sessions
}

func (sm *SessionManager) GetSessions() map[string]*protocol.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions
}

func getRowLetter(row int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if row <= 26 {
		return string(letters[row-1])
	}
	
	result := ""
	for row > 0 {
		row--
		result = string(letters[row%26]) + result
		row /= 26
	}
	return result
}
