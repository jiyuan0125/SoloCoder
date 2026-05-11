package core

import (
	"fmt"
	"sync"
	"time"

	"usedcar/api"
)

type Store struct {
	mu           sync.RWMutex
	vehicles     map[string]*VehicleInternal
	evaluations  map[string][]*api.Evaluation
	evaluationMap map[string]*api.Evaluation
	deposits     map[string]*DepositInternal
	transfers    map[string]*api.TransferRecord
	reportCounter map[string]int
	refPrices    *ReferencePriceDB
}

func NewStore() *Store {
	return &Store{
		vehicles:       make(map[string]*VehicleInternal),
		evaluations:    make(map[string][]*api.Evaluation),
		evaluationMap:  make(map[string]*api.Evaluation),
		deposits:       make(map[string]*DepositInternal),
		transfers:      make(map[string]*api.TransferRecord),
		reportCounter:  make(map[string]int),
		refPrices:      NewReferencePriceDB(),
	}
}

func (s *Store) CreateVehicle(req api.CreateVehicleRequest) (*api.Vehicle, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID("VH")
	now := time.Now().Unix()

	vehicle := &api.Vehicle{
		ID:           id,
		Brand:        req.Brand,
		Model:        req.Model,
		Year:         req.Year,
		Color:        req.Color,
		Mileage:      req.Mileage,
		Displacement: req.Displacement,
		Transmission: req.Transmission,
		Status:       api.StatusInStock,
		CreatedAt:    now,
	}

	s.vehicles[id] = &VehicleInternal{
		Vehicle: vehicle,
	}

	return vehicle, nil
}

func (s *Store) GetVehicle(vehicleID string) (*api.Vehicle, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.vehicles[vehicleID]
	if !ok {
		return nil, false
	}
	return v.Vehicle, true
}

func (s *Store) ListVehicles(status api.VehicleStatus) []*api.Vehicle {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vehicles := make([]*api.Vehicle, 0)
	for _, v := range s.vehicles {
		if status == "" || v.Status == status {
			vehicles = append(vehicles, v.Vehicle)
		}
	}
	return vehicles
}

func (s *Store) EvaluateVehicle(req api.EvaluateVehicleRequest) (*api.Evaluation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.vehicles[req.VehicleID]
	if !ok {
		return nil, fmt.Errorf("vehicle not found: %s", req.VehicleID)
	}

	if v.Status == api.StatusSold {
		return nil, fmt.Errorf("vehicle already sold")
	}

	refPrice, _ := s.refPrices.GetReferencePrice(v.Brand, v.Model, v.Year)
	finalPrice := CalculateFinalPrice(refPrice, v.Mileage, req.Condition)

	reportID := s.generateReportID()
	now := time.Now().Unix()

	isDuplicate := s.checkDuplicateEvaluation(req.VehicleID, req.Description, now)

	evaluation := &api.Evaluation{
		ReportID:          reportID,
		VehicleID:         req.VehicleID,
		Evaluator:         req.Evaluator,
		Condition:         req.Condition,
		ReferencePriceFen: refPrice,
		FinalPriceFen:     finalPrice,
		Description:       req.Description,
		CreatedAt:         now,
		IsDuplicate:       &isDuplicate,
	}

	s.evaluations[req.VehicleID] = append(s.evaluations[req.VehicleID], evaluation)
	s.evaluationMap[reportID] = evaluation
	v.LastEvaluation = evaluation

	return evaluation, nil
}

func (s *Store) ListEvaluations(vehicleID string) ([]*api.Evaluation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.vehicles[vehicleID]; !ok {
		return nil, false
	}

	evals, ok := s.evaluations[vehicleID]
	if !ok {
		return []*api.Evaluation{}, true
	}

	result := make([]*api.Evaluation, len(evals))
	copy(result, evals)
	return result, true
}

func (s *Store) PayDeposit(req api.PayDepositRequest) (*api.Deposit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.vehicles[req.VehicleID]
	if !ok {
		return nil, fmt.Errorf("vehicle not found: %s", req.VehicleID)
	}

	if v.Status != api.StatusInStock {
		return nil, fmt.Errorf("vehicle status is %s, cannot pay deposit", v.Status)
	}

	var eval *api.Evaluation
	if req.ReportID != "" {
		eval, ok = s.evaluationMap[req.ReportID]
		if !ok {
			return nil, fmt.Errorf("evaluation report not found: %s", req.ReportID)
		}
	} else if v.LastEvaluation != nil {
		eval = v.LastEvaluation
	} else {
		return nil, fmt.Errorf("vehicle has no evaluation")
	}

	var amount int64
	if req.AmountFen != nil {
		amount = *req.AmountFen
	} else {
		amount = CalculateDepositAmount(eval.FinalPriceFen)
	}

	depositID := generateID("DP")
	now := time.Now().Unix()

	deposit := &api.Deposit{
		ID:              depositID,
		VehicleID:       req.VehicleID,
		AmountFen:       amount,
		CustomerName:    req.CustomerName,
		CreatedAt:       now,
		TransactionType: "deposit",
	}

	s.deposits[depositID] = &DepositInternal{
		Deposit:  deposit,
		ReportID: eval.ReportID,
	}

	v.Status = api.StatusReserved

	return deposit, nil
}

func (s *Store) PayFull(req api.PayFullRequest) (*api.TransferRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.vehicles[req.VehicleID]
	if !ok {
		return nil, fmt.Errorf("vehicle not found: %s", req.VehicleID)
	}

	if v.Status != api.StatusReserved {
		return nil, fmt.Errorf("vehicle status is %s, cannot pay full", v.Status)
	}

	if v.LastEvaluation == nil {
		return nil, fmt.Errorf("vehicle has no evaluation")
	}

	transferID := generateID("TR")
	now := time.Now().Unix()

	transfer := &api.TransferRecord{
		ID:           transferID,
		VehicleID:    req.VehicleID,
		FullPriceFen: v.LastEvaluation.FinalPriceFen,
		CustomerName: req.CustomerName,
		CreatedAt:    now,
	}

	s.transfers[transferID] = transfer
	v.Status = api.StatusSold

	return transfer, nil
}

func (s *Store) generateReportID() string {
	today := time.Now().Format("20060102")
	key := "PG" + today
	count := s.reportCounter[key] + 1
	s.reportCounter[key] = count
	return fmt.Sprintf("%s%04d", key, count)
}

func (s *Store) checkDuplicateEvaluation(vehicleID, description string, now int64) bool {
	evals := s.evaluations[vehicleID]
	for i := len(evals) - 1; i >= 0; i-- {
		eval := evals[i]
		if eval.Description == description {
			diff := now - eval.CreatedAt
			if diff <= 7*24*60*60 {
				return true
			}
		}
	}
	return false
}

func generateID(prefix string) string {
	now := time.Now()
	return fmt.Sprintf("%s%s%d", prefix, now.Format("20060102150405"), now.Nanosecond()/1000)
}
