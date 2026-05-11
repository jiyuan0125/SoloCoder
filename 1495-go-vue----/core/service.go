package core

import (
	"crypto/rand"
	"encoding/hex"
	"sort"
	"sync"
	"time"
)

func GenerateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

type Service struct {
	store Store
	mu    sync.Mutex
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateCustomer(name, address string, placeType PlaceType, area float64, pestTypes []PestType) (*Customer, error) {
	if name == "" || address == "" || area <= 0 {
		return nil, ErrInvalidData
	}
	c := &Customer{
		ID:        GenerateID(),
		Name:      name,
		Address:   address,
		PlaceType: placeType,
		Area:      area,
		PestTypes: pestTypes,
	}
	if err := s.store.CreateCustomer(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) GetCustomer(id string) (*Customer, error) {
	return s.store.GetCustomer(id)
}

func (s *Service) ListCustomers() ([]*Customer, error) {
	return s.store.ListCustomers()
}

func (s *Service) CreateContract(customerID string, startDate, endDate time.Time, servicePerVisit float64) (*Contract, error) {
	if customerID == "" || startDate.After(endDate) || servicePerVisit <= 0 {
		return nil, ErrInvalidData
	}
	if _, err := s.store.GetCustomer(customerID); err != nil {
		return nil, err
	}
	c := &Contract{
		ID:              GenerateID(),
		CustomerID:      customerID,
		StartDate:       startDate,
		EndDate:         endDate,
		ServicePerVisit: servicePerVisit,
		AnnualTotal:     calculateAnnualTotal(servicePerVisit),
	}
	if err := s.store.CreateContract(c); err != nil {
		return nil, err
	}
	return c, nil
}

func calculateAnnualTotal(servicePerVisit float64) float64 {
	return servicePerVisit * 12
}

func (s *Service) GetContract(id string) (*Contract, error) {
	return s.store.GetContract(id)
}

func (s *Service) ListContracts() ([]*Contract, error) {
	contracts, err := s.store.ListContracts()
	if err != nil {
		return nil, err
	}
	for _, c := range contracts {
		if cust, err := s.store.GetCustomer(c.CustomerID); err == nil {
			c.Customer = cust
		}
	}
	return contracts, nil
}

func (s *Service) CreateControlPoint(contractID, code, location, description string) (*ControlPoint, error) {
	if contractID == "" || code == "" {
		return nil, ErrInvalidData
	}
	if _, err := s.store.GetContract(contractID); err != nil {
		return nil, err
	}
	cp := &ControlPoint{
		ID:          GenerateID(),
		ContractID:  contractID,
		Code:        code,
		Location:    location,
		Description: description,
	}
	if err := s.store.CreateControlPoint(cp); err != nil {
		return nil, err
	}
	return cp, nil
}

func (s *Service) GetControlPoints(contractID string) ([]*ControlPoint, error) {
	return s.store.GetControlPointsByContract(contractID)
}

func (s *Service) CreateStaff(name, role, phone string) (*Staff, error) {
	if name == "" {
		return nil, ErrInvalidData
	}
	st := &Staff{
		ID:     GenerateID(),
		Name:   name,
		Role:   role,
		Phone:  phone,
		Active: true,
	}
	if err := s.store.CreateStaff(st); err != nil {
		return nil, err
	}
	return st, nil
}

func (s *Service) ListActiveStaff() ([]*Staff, error) {
	return s.store.ListActiveStaff()
}

func (s *Service) CreateChemical(name, unit string, unitPrice, stock, safetyStock float64) (*Chemical, error) {
	if name == "" || unit == "" || unitPrice < 0 || stock < 0 || safetyStock < 0 {
		return nil, ErrInvalidData
	}
	c := &Chemical{
		ID:           GenerateID(),
		Name:         name,
		Unit:         unit,
		UnitPrice:    unitPrice,
		Stock:        stock,
		SafetyStock:  safetyStock,
		NeedsPurchase: stock <= safetyStock,
	}
	if err := s.store.CreateChemical(c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Service) ListChemicals() ([]*Chemical, error) {
	return s.store.ListChemicals()
}

func (s *Service) GetChemical(id string) (*Chemical, error) {
	return s.store.GetChemical(id)
}

type UsageRequest struct {
	ChemicalID string
	Amount     float64
}

func (s *Service) CreateInspection(contractID, inspectorID string, inspectionDate time.Time, controlPoints []InspectionPoint, notes string) (*Inspection, error) {
	if contractID == "" || inspectorID == "" {
		return nil, ErrInvalidData
	}
	contract, err := s.store.GetContract(contractID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if contract.EndDate.Before(now) {
		return nil, ErrContractExpired
	}
	inspector, err := s.store.GetStaff(inspectorID)
	if err != nil {
		return nil, err
	}
	pestRecords := aggregatePestRecords(controlPoints)
	level := determineInspectionLevel(pestRecords)
	insp := &Inspection{
		ID:             GenerateID(),
		ContractID:     contractID,
		Contract:       contract,
		InspectorID:    inspectorID,
		InspectorName:  inspector.Name,
		InspectionDate: inspectionDate,
		ControlPoints:  controlPoints,
		PestRecords:    pestRecords,
		Level:          level,
		Notes:          notes,
	}
	if err := s.store.CreateInspection(insp); err != nil {
		return nil, err
	}
	return insp, nil
}

func aggregatePestRecords(points []InspectionPoint) []PestRecord {
	counts := make(map[PestType]int)
	for _, p := range points {
		for _, r := range p.PestRecords {
			counts[r.PestType] += r.Count
		}
	}
	records := make([]PestRecord, 0, len(counts))
	for pt, cnt := range counts {
		if cnt > 0 {
			records = append(records, PestRecord{PestType: pt, Count: cnt})
		}
	}
	return records
}

func determineInspectionLevel(records []PestRecord) InspectionLevel {
	total := 0
	for _, r := range records {
		total += r.Count
	}
	if total == 0 {
		return InspectionLevelNormal
	}
	if total >= 10 {
		return InspectionLevelSevere
	}
	return InspectionLevelMinor
}

func (s *Service) ListInspections(contractID string) ([]*Inspection, error) {
	return s.store.ListInspectionsByContract(contractID)
}

func (s *Service) CreateOperation(contractID, operatorID string, operationDate time.Time, scope string, usages []UsageRequest, notes string) (*Operation, error) {
	if contractID == "" || operatorID == "" {
		return nil, ErrInvalidData
	}
	contract, err := s.store.GetContract(contractID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if contract.EndDate.Before(now) {
		return nil, ErrContractExpired
	}
	operator, err := s.store.GetStaff(operatorID)
	if err != nil {
		return nil, err
	}
	todayOps, err := s.store.ListOperationsByStaff(operatorID, operationDate)
	if err != nil {
		return nil, err
	}
	if len(todayOps) >= 4 {
		return nil, ErrStaffOverloaded
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	chemicalUsages := make([]ChemicalUsage, 0, len(usages))
	totalChemicalCost := 0.0
	for _, u := range usages {
		if u.Amount <= 0 {
			continue
		}
		chem, err := s.store.GetChemical(u.ChemicalID)
		if err != nil {
			return nil, err
		}
		if chem.Stock < u.Amount {
			return nil, ErrInsufficientStock
		}
		chem.Stock -= u.Amount
		if chem.Stock <= chem.SafetyStock {
			chem.NeedsPurchase = true
		}
		if err := s.store.UpdateChemical(chem); err != nil {
			return nil, err
		}
		cost := chem.UnitPrice * u.Amount
		totalChemicalCost += cost
		chemicalUsages = append(chemicalUsages, ChemicalUsage{
			ChemicalID: chem.ID,
			Name:       chem.Name,
			Amount:     u.Amount,
			Unit:       chem.Unit,
			UnitPrice:  chem.UnitPrice,
			TotalCost:  cost,
		})
	}
	op := &Operation{
		ID:                 GenerateID(),
		ContractID:         contractID,
		Contract:           contract,
		OperationDate:      operationDate,
		OperatorID:         operatorID,
		OperatorName:       operator.Name,
		Scope:              scope,
		ChemicalUsages:     chemicalUsages,
		TotalChemicalCost:  totalChemicalCost,
		ServiceCost:        contract.ServicePerVisit,
		TotalCost:          totalChemicalCost + contract.ServicePerVisit,
		Notes:              notes,
		Status:             "completed",
	}
	if err := s.store.CreateOperation(op); err != nil {
		return nil, err
	}
	return op, nil
}

func (s *Service) ListOperations(contractID string) ([]*Operation, error) {
	return s.store.ListOperationsByContract(contractID)
}

func (s *Service) EvaluateEffect(inspectionID string) (*EffectEvaluation, error) {
	insp, err := s.store.GetInspection(inspectionID)
	if err != nil {
		return nil, err
	}
	latest, err := s.store.GetLatestInspections(insp.ContractID, 2)
	if err != nil {
		return nil, err
	}
	if len(latest) < 2 {
		return nil, ErrInvalidData
	}
	current := latest[0]
	previous := latest[1]
	if current.ID != inspectionID {
		current, previous = previous, current
	}
	prevCount := countTotalPests(previous.PestRecords)
	currCount := countTotalPests(current.PestRecords)
	var reductionRate float64
	if prevCount > 0 {
		reductionRate = float64(prevCount-currCount) / float64(prevCount) * 100
	}
	isEffective := reductionRate >= 50
	needsReadjustment := reductionRate < 20 && currCount > 0
	eval := &EffectEvaluation{
		InspectionID:      inspectionID,
		RelatedOperationID: current.RelatedOperationID,
		PreviousCount:     prevCount,
		CurrentCount:      currCount,
		ReductionRate:     reductionRate,
		IsEffective:       isEffective,
		NeedsReadjustment: needsReadjustment,
	}
	if err := s.store.CreateEffectEvaluation(eval); err != nil {
		return nil, err
	}
	return eval, nil
}

func countTotalPests(records []PestRecord) int {
	total := 0
	for _, r := range records {
		total += r.Count
	}
	return total
}

func (s *Service) GetPendingServices() ([]*PendingService, error) {
	contracts, err := s.store.ListActiveContracts()
	if err != nil {
		return nil, err
	}
	pending := make([]*PendingService, 0)
	now := time.Now()
	for _, c := range contracts {
		cust, err := s.store.GetCustomer(c.CustomerID)
		if err != nil {
			continue
		}
		lastOp, err := s.store.GetLastOperation(c.ID)
		if err != nil && err != ErrNotFound {
			continue
		}
		var lastServiceDate time.Time
		if lastOp != nil {
			lastServiceDate = lastOp.OperationDate
		} else {
			lastServiceDate = c.StartDate
		}
		nextDate := calculateNextServiceDate(cust.PlaceType, lastServiceDate)
		if nextDate.Before(now) || nextDate.Equal(now) {
			pending = append(pending, &PendingService{
				ContractID:      c.ID,
				CustomerName:    cust.Name,
				PlaceType:       cust.PlaceType,
				LastServiceDate: lastServiceDate,
				NextServiceDate: nextDate,
				ContractEndDate: c.EndDate,
			})
		}
	}
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].NextServiceDate.Before(pending[j].NextServiceDate)
	})
	return pending, nil
}

func calculateNextServiceDate(placeType PlaceType, lastDate time.Time) time.Time {
	switch placeType {
	case PlaceTypeRestaurant:
		return lastDate.AddDate(0, 0, 15)
	case PlaceTypeFoodFactory:
		return lastDate.AddDate(0, 0, 7)
	case PlaceTypeOffice:
		return lastDate.AddDate(0, 3, 0)
	default:
		return lastDate.AddDate(0, 1, 0)
	}
}
