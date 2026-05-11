package core

import (
	"errors"
	"fmt"
	"parking-system/common"
	"strconv"
	"sync"
	"time"
)

type ParkingService struct {
	spots        *SpotManager
	monthlyCards *MonthlyCardManager
	records      map[string]*ParkingRecord
	activePlates map[string]string
	mu           sync.RWMutex
}

func NewParkingService() *ParkingService {
	return &ParkingService{
		spots:        NewSpotManager(),
		monthlyCards: NewMonthlyCardManager(),
		records:      make(map[string]*ParkingRecord),
		activePlates: make(map[string]string),
	}
}

func (s *ParkingService) AddParkingSpot(id, area, number string, spotType common.ParkingSpotType) error {
	return s.spots.AddSpot(id, area, number, spotType)
}

func (s *ParkingService) GetParkingSpot(id string) (*ParkingSpot, bool) {
	return s.spots.GetSpot(id)
}

func (s *ParkingService) UpdateParkingSpotStatus(id string, status common.ParkingSpotStatus) error {
	return s.spots.UpdateSpotStatus(id, status)
}

func (s *ParkingService) ListParkingSpots(area, status string) []*ParkingSpot {
	return s.spots.ListSpots(area, status)
}

func (s *ParkingService) GetGuidance() common.ParkingGuidanceDTO {
	return s.spots.GetGuidanceInfo(s.monthlyCards)
}

func (s *ParkingService) CheckIn(plateNumber string) (*common.CheckInResponse, error) {
	if plateNumber == "" {
		return nil, errors.New("plate number is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.activePlates[plateNumber]; exists {
		return nil, fmt.Errorf("vehicle with plate %s is already checked in", plateNumber)
	}

	now := time.Now()
	s.monthlyCards.CheckAndExpireCards(now)

	activeCard := s.monthlyCards.GetActiveCardForPlate(plateNumber)
	isMonthlyCard := activeCard != nil

	var spot *ParkingSpot
	var err error

	if isMonthlyCard {
		var preferredType common.ParkingSpotType
		switch activeCard.CardType {
		case common.CardTypeCharging:
			preferredType = common.SpotTypeCharging
		default:
			preferredType = common.SpotTypeNormal
		}
		spot, err = s.spots.AllocateSpot(true, preferredType, activeCard.SpotID)
	} else {
		spot, err = s.spots.AllocateSpot(false, common.SpotTypeNormal, "")
	}

	if err != nil {
		return nil, err
	}

	recordID := "REC-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	record := &ParkingRecord{
		ID:            recordID,
		PlateNumber:   plateNumber,
		CheckInTime:   now,
		SpotID:        spot.GetID(),
		IsActive:      true,
		IsMonthlyCard: isMonthlyCard,
	}

	s.records[recordID] = record
	s.activePlates[plateNumber] = recordID

	return &common.CheckInResponse{
		PlateNumber:   plateNumber,
		CheckInTime:   now,
		SpotID:        spot.GetID(),
		IsMonthlyCard: isMonthlyCard,
	}, nil
}

func (s *ParkingService) CheckOut(plateNumber string) (*common.CheckOutResponse, error) {
	if plateNumber == "" {
		return nil, errors.New("plate number is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	recordID, exists := s.activePlates[plateNumber]
	if !exists {
		return nil, fmt.Errorf("no active record found for plate %s", plateNumber)
	}

	record, ok := s.records[recordID]
	if !ok {
		delete(s.activePlates, plateNumber)
		return nil, fmt.Errorf("record not found for plate %s", plateNumber)
	}

	now := time.Now()
	checkIn := record.GetCheckInTime()
	duration := now.Sub(checkIn).Minutes()
	amount := 0.0
	isMonthlyCard := record.GetIsMonthlyCard()
	var freeTimeUsed float64

	if !isMonthlyCard {
		amount = CalculateFee(checkIn, now)
		freeTimeUsed = GetFreeMinutesUsed(checkIn, now)
	}

	record.SetAmount(amount)
	record.SetCheckOutAndInactive(now)
	delete(s.activePlates, plateNumber)

	err := s.spots.FreeSpot(record.GetSpotID())
	if err != nil {
		return nil, err
	}

	return &common.CheckOutResponse{
		PlateNumber:   plateNumber,
		CheckInTime:   checkIn,
		CheckOutTime:  now,
		Duration:      duration,
		Amount:        amount,
		IsMonthlyCard: isMonthlyCard,
		FreeTimeUsed:  freeTimeUsed,
	}, nil
}

func (s *ParkingService) QueryFee(plateNumber string) (*common.QueryFeeResponse, error) {
	if plateNumber == "" {
		return nil, errors.New("plate number is required")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	recordID, exists := s.activePlates[plateNumber]
	if !exists {
		return nil, fmt.Errorf("no active record found for plate %s", plateNumber)
	}

	record, ok := s.records[recordID]
	if !ok {
		return nil, fmt.Errorf("record not found for plate %s", plateNumber)
	}

	now := time.Now()
	checkIn := record.GetCheckInTime()
	duration := now.Sub(checkIn).Minutes()
	isMonthlyCard := record.GetIsMonthlyCard()

	estimatedAmount := 0.0
	if !isMonthlyCard {
		estimatedAmount = CalculateFee(checkIn, now)
	}

	return &common.QueryFeeResponse{
		PlateNumber:     plateNumber,
		CheckInTime:     checkIn,
		CurrentTime:     now,
		Duration:        duration,
		EstimatedAmount: estimatedAmount,
		IsMonthlyCard:   isMonthlyCard,
	}, nil
}

func (s *ParkingService) CreateMonthlyCard(
	cardType common.MonthlyCardType,
	ownerName string,
	plateNumber string,
	spotID string,
) (*MonthlyCard, error) {
	return s.monthlyCards.CreateCard(cardType, ownerName, plateNumber, spotID, time.Now())
}

func (s *ParkingService) RenewMonthlyCard(cardID string) error {
	return s.monthlyCards.RenewCard(cardID)
}

func (s *ParkingService) ListMonthlyCards(plateNumber string, activeOnly bool) []*MonthlyCard {
	return s.monthlyCards.ListCards(plateNumber, activeOnly)
}

func (s *ParkingService) ListParkingRecords(plateNumber string, activeOnly bool) []*ParkingRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*ParkingRecord
	for _, record := range s.records {
		matchesPlate := plateNumber == "" || record.GetPlateNumber() == plateNumber
		matchesActive := !activeOnly || record.GetIsActive()
		if matchesPlate && matchesActive {
			result = append(result, record)
		}
	}
	return result
}

func (s *ParkingService) GetMonthlyCardByID(id string) (*MonthlyCard, bool) {
	return s.monthlyCards.GetCardByID(id)
}
