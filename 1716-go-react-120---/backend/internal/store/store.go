package store

import (
	"sync"

	"ambulance-scheduler/internal/models"
)

type Store struct {
	mu             sync.RWMutex
	calls          map[string]*models.EmergencyCall
	vehicles       map[string]*models.Ambulance
	dispatchRecords map[string]*models.DispatchRecord
	triageRecords  map[string]*models.TriageRecord
	bills          map[string]*models.Bill
}

func New() *Store {
	return &Store{
		calls:          make(map[string]*models.EmergencyCall),
		vehicles:       make(map[string]*models.Ambulance),
		dispatchRecords: make(map[string]*models.DispatchRecord),
		triageRecords:  make(map[string]*models.TriageRecord),
		bills:          make(map[string]*models.Bill),
	}
}

func (s *Store) CreateCall(call *models.EmergencyCall) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls[call.ID] = call
}

func (s *Store) GetCall(id string) (*models.EmergencyCall, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	call, ok := s.calls[id]
	return call, ok
}

func (s *Store) UpdateCall(call *models.EmergencyCall) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls[call.ID] = call
}

func (s *Store) ListCalls() []*models.EmergencyCall {
	s.mu.RLock()
	defer s.mu.RUnlock()
	calls := make([]*models.EmergencyCall, 0, len(s.calls))
	for _, call := range s.calls {
		calls = append(calls, call)
	}
	return calls
}

func (s *Store) CreateVehicle(vehicle *models.Ambulance) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vehicles[vehicle.ID] = vehicle
}

func (s *Store) GetVehicle(id string) (*models.Ambulance, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vehicle, ok := s.vehicles[id]
	return vehicle, ok
}

func (s *Store) UpdateVehicle(vehicle *models.Ambulance) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vehicles[vehicle.ID] = vehicle
}

func (s *Store) ListVehicles() []*models.Ambulance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vehicles := make([]*models.Ambulance, 0, len(s.vehicles))
	for _, vehicle := range s.vehicles {
		vehicles = append(vehicles, vehicle)
	}
	return vehicles
}

func (s *Store) CreateDispatchRecord(record *models.DispatchRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dispatchRecords[record.ID] = record
}

func (s *Store) GetDispatchRecord(id string) (*models.DispatchRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.dispatchRecords[id]
	return record, ok
}

func (s *Store) ListDispatchRecords() []*models.DispatchRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	records := make([]*models.DispatchRecord, 0, len(s.dispatchRecords))
	for _, record := range s.dispatchRecords {
		records = append(records, record)
	}
	return records
}

func (s *Store) CreateTriageRecord(record *models.TriageRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.triageRecords[record.ID] = record
}

func (s *Store) ListTriageRecords() []*models.TriageRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	records := make([]*models.TriageRecord, 0, len(s.triageRecords))
	for _, record := range s.triageRecords {
		records = append(records, record)
	}
	return records
}

func (s *Store) GetBill(id string) (*models.Bill, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	bill, ok := s.bills[id]
	return bill, ok
}

func (s *Store) CreateBill(bill *models.Bill) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bills[bill.ID] = bill
}

func (s *Store) UpdateBill(bill *models.Bill) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bills[bill.ID] = bill
}

func (s *Store) AtomicallyDispatch(callID, vehicleID string, apply func(call *models.EmergencyCall, vehicle *models.Ambulance) *models.DispatchRecord) *models.DispatchRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	call, callOk := s.calls[callID]
	vehicle, vehicleOk := s.vehicles[vehicleID]
	
	if !callOk || !vehicleOk {
		return nil
	}
	
	record := apply(call, vehicle)
	if record != nil {
		s.calls[call.ID] = call
		s.vehicles[vehicle.ID] = vehicle
		s.dispatchRecords[record.ID] = record
	}
	
	return record
}

func (s *Store) AtomicallyUpdateStatus(vehicleID string, newStatus models.VehicleStatus, apply func(vehicle *models.Ambulance) bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	vehicle, ok := s.vehicles[vehicleID]
	if !ok {
		return false
	}
	
	if apply(vehicle) {
		s.vehicles[vehicle.ID] = vehicle
		return true
	}
	
	return false
}

func (s *Store) AtomicallyUpgradeSeverity(callID string, apply func(call *models.EmergencyCall) bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	call, ok := s.calls[callID]
	if !ok {
		return false
	}
	
	if apply(call) {
		s.calls[call.ID] = call
		return true
	}
	
	return false
}
