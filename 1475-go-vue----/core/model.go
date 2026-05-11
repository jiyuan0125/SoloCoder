package core

import (
	"parking-system/common"
	"sync"
	"time"
)

type ParkingSpot struct {
	ID     string
	Area   string
	Number string
	Type   common.ParkingSpotType
	Status common.ParkingSpotStatus
	mu     sync.RWMutex
}

func (s *ParkingSpot) GetID() string { return s.ID }
func (s *ParkingSpot) GetArea() string { return s.Area }
func (s *ParkingSpot) GetNumber() string { return s.Number }
func (s *ParkingSpot) GetType() common.ParkingSpotType { return s.Type }

func (s *ParkingSpot) GetStatus() common.ParkingSpotStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status
}

func (s *ParkingSpot) SetStatus(status common.ParkingSpotStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
}

type ParkingRecord struct {
	ID           string
	PlateNumber  string
	CheckInTime  time.Time
	CheckOutTime *time.Time
	SpotID       string
	Amount       float64
	IsActive     bool
	IsMonthlyCard bool
	mu           sync.RWMutex
}

func (r *ParkingRecord) GetID() string { return r.ID }
func (r *ParkingRecord) GetPlateNumber() string { return r.PlateNumber }
func (r *ParkingRecord) GetCheckInTime() time.Time { return r.CheckInTime }
func (r *ParkingRecord) GetCheckOutTime() *time.Time { return r.CheckOutTime }
func (r *ParkingRecord) GetSpotID() string { return r.SpotID }

func (r *ParkingRecord) GetAmount() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Amount
}

func (r *ParkingRecord) SetAmount(amount float64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Amount = amount
}

func (r *ParkingRecord) GetIsActive() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.IsActive
}

func (r *ParkingRecord) SetCheckOutAndInactive(checkOutTime time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.CheckOutTime = &checkOutTime
	r.IsActive = false
}

func (r *ParkingRecord) GetIsMonthlyCard() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.IsMonthlyCard
}

type MonthlyCard struct {
	ID           string
	CardType     common.MonthlyCardType
	OwnerName    string
	PlateNumber  string
	SpotID       string
	StartDate    time.Time
	EndDate      time.Time
	GraceEndDate time.Time
	IsActive     bool
	mu           sync.RWMutex
}

func (c *MonthlyCard) GetID() string { return c.ID }
func (c *MonthlyCard) GetCardType() common.MonthlyCardType { return c.CardType }
func (c *MonthlyCard) GetOwnerName() string { return c.OwnerName }
func (c *MonthlyCard) GetPlateNumber() string { return c.PlateNumber }
func (c *MonthlyCard) GetSpotID() string { return c.SpotID }
func (c *MonthlyCard) GetStartDate() time.Time { return c.StartDate }
func (c *MonthlyCard) GetEndDate() time.Time { return c.EndDate }
func (c *MonthlyCard) GetGraceEndDate() time.Time { return c.GraceEndDate }

func (c *MonthlyCard) GetIsActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IsActive
}

func (c *MonthlyCard) SetIsActive(active bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.IsActive = active
}

func (c *MonthlyCard) Renew() {
	c.mu.Lock()
	defer c.mu.Unlock()
	newEnd := c.EndDate.AddDate(0, 1, 0)
	c.EndDate = newEnd
	c.GraceEndDate = newEnd.AddDate(0, 0, 3)
	c.IsActive = true
}
