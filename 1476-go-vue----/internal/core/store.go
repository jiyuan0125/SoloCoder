package core

import (
	"errors"
	"fmt"
	"gasstation/internal/common"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	StockThreshold int64 = 500
)

var (
	ErrStationNotFound      = errors.New("station not found")
	ErrFuelNotFound         = errors.New("fuel not found at station")
	ErrMemberNotFound       = errors.New("member not found")
	ErrInsufficientStock    = errors.New("insufficient stock")
	ErrPriceRequestNotFound = errors.New("price change request not found")
	ErrPriceAlreadyProcessed = errors.New("price change request already processed")
	ErrInvalidLiters        = errors.New("liters must be positive with max 2 decimal places")
)

type Store struct {
	mu            sync.RWMutex
	stations      map[string]*common.GasStation
	stationFuels  map[string]map[string]*common.StationFuel
	members       map[string]*common.Member
	membersByPhone map[string]string
	records       map[string]*common.RefuelRecord
	priceRequests map[string]*common.PriceChangeRequest
	priceHistory  []*common.PriceChangeRequest
	restockTodos  map[string]*common.RestockTodo
	nextID        int64
}

func NewStore() *Store {
	return &Store{
		stations:       make(map[string]*common.GasStation),
		stationFuels:   make(map[string]map[string]*common.StationFuel),
		members:        make(map[string]*common.Member),
		membersByPhone: make(map[string]string),
		records:        make(map[string]*common.RefuelRecord),
		priceRequests:  make(map[string]*common.PriceChangeRequest),
		priceHistory:   make([]*common.PriceChangeRequest, 0),
		restockTodos:   make(map[string]*common.RestockTodo),
		nextID:         1,
	}
}

func (s *Store) GenerateID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextID
	s.nextID++
	return strconv.FormatInt(id, 10)
}

func (s *Store) CreateStation(station *common.GasStation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	station.ID = s.GenerateID()
	station.CreatedAt = time.Now()
	station.UpdatedAt = station.CreatedAt
	s.stations[station.ID] = station
	s.stationFuels[station.ID] = make(map[string]*common.StationFuel)
	return nil
}

func (s *Store) GetStation(id string) (*common.GasStation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	station, ok := s.stations[id]
	if !ok {
		return nil, ErrStationNotFound
	}
	return station, nil
}

func (s *Store) UpdateStation(id string, req common.UpdateStationRequest) (*common.GasStation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	station, ok := s.stations[id]
	if !ok {
		return nil, ErrStationNotFound
	}
	if req.Name != nil {
		station.Name = *req.Name
	}
	if req.Address != nil {
		station.Address = *req.Address
	}
	if req.Latitude != nil {
		station.Latitude = *req.Latitude
	}
	if req.Longitude != nil {
		station.Longitude = *req.Longitude
	}
	if req.Phone != nil {
		station.Phone = *req.Phone
	}
	if req.IsOpen != nil {
		station.IsOpen = *req.IsOpen
	}
	station.UpdatedAt = time.Now()
	return station, nil
}

func (s *Store) ListStations() []*common.GasStation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stations := make([]*common.GasStation, 0, len(s.stations))
	for _, station := range s.stations {
		stations = append(stations, station)
	}
	return stations
}

func (s *Store) AddFuelToStation(stationID string, fuelCode, fuelName string, price, stock int64) (*common.StationFuel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.stations[stationID]; !ok {
		return nil, ErrStationNotFound
	}
	sf := &common.StationFuel{
		StationID: stationID,
		FuelCode:  fuelCode,
		Price:     price,
		Stock:     stock,
		UpdatedAt: time.Now(),
	}
	s.stationFuels[stationID][fuelCode] = sf
	s.checkStockAndCreateTodoLocked(stationID, fuelCode, stock)
	return sf, nil
}

func (s *Store) GetStationFuel(stationID, fuelCode string) (*common.StationFuel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fuels, ok := s.stationFuels[stationID]
	if !ok {
		return nil, ErrStationNotFound
	}
	sf, ok := fuels[fuelCode]
	if !ok {
		return nil, ErrFuelNotFound
	}
	return sf, nil
}

func (s *Store) ListStationFuels(stationID string) ([]*common.StationFuel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fuels, ok := s.stationFuels[stationID]
	if !ok {
		return nil, ErrStationNotFound
	}
	result := make([]*common.StationFuel, 0, len(fuels))
	for _, sf := range fuels {
		result = append(result, sf)
	}
	return result, nil
}

func (s *Store) UpdateFuelPrice(stationID, fuelCode string, newPrice int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fuels, ok := s.stationFuels[stationID]
	if !ok {
		return ErrStationNotFound
	}
	sf, ok := fuels[fuelCode]
	if !ok {
		return ErrFuelNotFound
	}
	sf.Price = newPrice
	sf.UpdatedAt = time.Now()
	return nil
}

func (s *Store) DeductStock(stationID, fuelCode string, liters int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	fuels, ok := s.stationFuels[stationID]
	if !ok {
		return ErrStationNotFound
	}
	sf, ok := fuels[fuelCode]
	if !ok {
		return ErrFuelNotFound
	}
	if sf.Stock < liters {
		return ErrInsufficientStock
	}
	sf.Stock -= liters
	sf.UpdatedAt = time.Now()
	s.checkStockAndCreateTodoLocked(stationID, fuelCode, sf.Stock)
	return nil
}

func (s *Store) checkStockAndCreateTodoLocked(stationID, fuelCode string, stock int64) {
	if stock >= StockThreshold {
		return
	}
	for _, todo := range s.restockTodos {
		if todo.StationID == stationID && todo.FuelCode == fuelCode && !todo.IsCompleted {
			todo.CurrentStock = stock
			return
		}
	}
	todo := &common.RestockTodo{
		ID:           s.GenerateID(),
		StationID:    stationID,
		FuelCode:     fuelCode,
		CurrentStock: stock,
		Threshold:    StockThreshold,
		IsCompleted:  false,
		CreatedAt:    time.Now(),
	}
	s.restockTodos[todo.ID] = todo
}

func (s *Store) ListRestockTodos() []*common.RestockTodo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	todos := make([]*common.RestockTodo, 0, len(s.restockTodos))
	for _, todo := range s.restockTodos {
		todos = append(todos, todo)
	}
	return todos
}

func (s *Store) CompleteRestockTodo(todoID string, newStock int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	todo, ok := s.restockTodos[todoID]
	if !ok {
		return errors.New("restock todo not found")
	}
	todo.IsCompleted = true
	now := time.Now()
	todo.CompletedAt = &now
	fuels := s.stationFuels[todo.StationID]
	if sf, ok := fuels[todo.FuelCode]; ok {
		sf.Stock = newStock
		sf.UpdatedAt = time.Now()
	}
	return nil
}

func (s *Store) RegisterMember(phone string) (*common.Member, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.membersByPhone[phone]; ok {
		return nil, errors.New("member already exists with this phone")
	}
	member := &common.Member{
		ID:           s.GenerateID(),
		Phone:        phone,
		Level:        common.LevelNormal,
		Points:       0,
		TotalSpent:   0,
		RegisteredAt: time.Now(),
	}
	s.members[member.ID] = member
	s.membersByPhone[phone] = member.ID
	return member, nil
}

func (s *Store) GetMember(id string) (*common.Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	member, ok := s.members[id]
	if !ok {
		return nil, ErrMemberNotFound
	}
	return member, nil
}

func (s *Store) GetMemberByPhone(phone string) (*common.Member, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.membersByPhone[phone]
	if !ok {
		return nil, ErrMemberNotFound
	}
	return s.members[id], nil
}

func (s *Store) UpdateMemberPointsAndLevel(memberID string, pointsDelta int64, amountSpent int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	member, ok := s.members[memberID]
	if !ok {
		return ErrMemberNotFound
	}
	member.Points += pointsDelta
	if member.Points < 0 {
		member.Points = 0
	}
	if amountSpent > 0 {
		member.TotalSpent += amountSpent
		oldLevel := member.Level
		newLevel := oldLevel
		if member.TotalSpent >= common.UpgradeGoldAmount {
			newLevel = common.LevelGold
		} else if member.TotalSpent >= common.UpgradeSilverAmount {
			newLevel = common.LevelSilver
		}
		if newLevel != oldLevel {
			member.Level = newLevel
			now := time.Now()
			member.LastLevelUpAt = &now
		}
	}
	return nil
}

func (s *Store) ListMembers() []*common.Member {
	s.mu.RLock()
	defer s.mu.RUnlock()
	members := make([]*common.Member, 0, len(s.members))
	for _, member := range s.members {
		members = append(members, member)
	}
	return members
}

func (s *Store) AddRefuelRecord(record *common.RefuelRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[record.ID] = record
	return nil
}

func (s *Store) GetRecord(id string) (*common.RefuelRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.records[id]
	if !ok {
		return nil, errors.New("record not found")
	}
	return record, nil
}

func (s *Store) ListRecords(stationID string, start, end time.Time) []*common.RefuelRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	records := make([]*common.RefuelRecord, 0)
	for _, record := range s.records {
		if record.CreatedAt.Before(start) || record.CreatedAt.After(end) {
			continue
		}
		if stationID != "" && record.StationID != stationID {
			continue
		}
		records = append(records, record)
	}
	return records
}

func (s *Store) ExportRecordsCSV(stationID string, start, end time.Time) string {
	records := s.ListRecords(stationID, start, end)
	var builder strings.Builder
	builder.WriteString("ID,StationID,FuelCode,Liters,UnitPrice,OriginalAmount,DiscountAmount,PointsUsed,PointsDeducted,FinalAmount,MemberID,MemberLevel,CreatedAt\n")
	for _, r := range records {
		memberID := ""
		if r.MemberID != nil {
			memberID = *r.MemberID
		}
		memberLevel := ""
		if r.MemberLevel != nil {
			memberLevel = *r.MemberLevel
		}
		line := fmt.Sprintf("%s,%s,%s,%.2f,%d,%d,%d,%d,%d,%d,%s,%s,%s\n",
			r.ID, r.StationID, r.FuelCode, r.Liters, r.UnitPrice,
			r.OriginalAmount, r.DiscountAmount, r.PointsUsed, r.PointsDeducted, r.FinalAmount,
			memberID, memberLevel, r.CreatedAt.Format(time.RFC3339))
		builder.WriteString(line)
	}
	return builder.String()
}

func (s *Store) SubmitPriceChangeRequest(stationID, fuelCode string, newPrice int64, submittedBy string) (*common.PriceChangeRequest, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fuels, ok := s.stationFuels[stationID]
	if !ok {
		return nil, ErrStationNotFound
	}
	sf, ok := fuels[fuelCode]
	if !ok {
		return nil, ErrFuelNotFound
	}
	req := &common.PriceChangeRequest{
		ID:          s.GenerateID(),
		StationID:   stationID,
		FuelCode:    fuelCode,
		OldPrice:    sf.Price,
		NewPrice:    newPrice,
		Status:      common.ApprovalStatusPending,
		SubmittedBy: submittedBy,
		SubmittedAt: time.Now(),
	}
	s.priceRequests[req.ID] = req
	s.priceHistory = append(s.priceHistory, req)
	return req, nil
}

func (s *Store) GetPriceChangeRequest(id string) (*common.PriceChangeRequest, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	req, ok := s.priceRequests[id]
	if !ok {
		return nil, ErrPriceRequestNotFound
	}
	return req, nil
}

func (s *Store) ListPriceChangeRequests(stationID string) []*common.PriceChangeRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*common.PriceChangeRequest, 0)
	for _, req := range s.priceRequests {
		if stationID != "" && req.StationID != stationID {
			continue
		}
		result = append(result, req)
	}
	return result
}

func (s *Store) ReviewPriceChangeRequest(id string, approved bool, reviewedBy string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	req, ok := s.priceRequests[id]
	if !ok {
		return ErrPriceRequestNotFound
	}
	if req.Status != common.ApprovalStatusPending {
		return ErrPriceAlreadyProcessed
	}
	if approved {
		req.Status = common.ApprovalStatusApproved
		fuels := s.stationFuels[req.StationID]
		if sf, ok := fuels[req.FuelCode]; ok {
			sf.Price = req.NewPrice
			sf.UpdatedAt = time.Now()
		}
	} else {
		req.Status = common.ApprovalStatusRejected
	}
	req.ApprovedBy = &reviewedBy
	now := time.Now()
	req.ApprovedAt = &now
	return nil
}

func (s *Store) GetPriceHistory(stationID, fuelCode string) []*common.PriceChangeRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*common.PriceChangeRequest, 0)
	for _, req := range s.priceHistory {
		if req.StationID != stationID || req.FuelCode != fuelCode {
			continue
		}
		result = append(result, req)
	}
	return result
}

func RoundLiters(liters float64) (float64, error) {
	if liters <= 0 {
		return 0, ErrInvalidLiters
	}
	scaled := liters * 100
	rounded := math.Round(scaled)
	if math.Abs(rounded-scaled) > 0.0001 {
		return 0, ErrInvalidLiters
	}
	return rounded / 100, nil
}

func LitersToInt64(liters float64) int64 {
	return int64(math.Round(liters))
}

func CalculateDiscount(level common.MemberLevel, amount int64) int64 {
	switch level {
	case common.LevelSilver:
		return amount * 5 / 100
	case common.LevelGold:
		return amount * 10 / 100
	default:
		return 0
	}
}

func CalculatePointsFromAmount(amount int64) int64 {
	return amount / 100
}

func PointsToDeduction(points int64) int64 {
	return (points / 100) * 100
}
