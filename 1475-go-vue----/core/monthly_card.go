package core

import (
	"errors"
	"fmt"
	"parking-system/common"
	"strconv"
	"sync"
	"time"
)

type MonthlyCardManager struct {
	cards map[string]*MonthlyCard
	mu    sync.RWMutex
}

func NewMonthlyCardManager() *MonthlyCardManager {
	return &MonthlyCardManager{
		cards: make(map[string]*MonthlyCard),
	}
}

func (m *MonthlyCardManager) CreateCard(
	cardType common.MonthlyCardType,
	ownerName string,
	plateNumber string,
	spotID string,
	now time.Time,
) (*MonthlyCard, error) {
	if ownerName == "" || plateNumber == "" {
		return nil, errors.New("owner name and plate number are required")
	}

	if cardType == common.CardTypeFixedSpot && spotID == "" {
		return nil, errors.New("fixed spot card requires spot id")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	existingActive := m.findActiveCardByPlateLocked(plateNumber)
	if existingActive != nil {
		return nil, fmt.Errorf("active card already exists for plate %s", plateNumber)
	}

	startDate := firstDayOfNextMonth(now)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)
	graceEndDate := endDate.AddDate(0, 0, 3)

	id := "CARD-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	card := &MonthlyCard{
		ID:           id,
		CardType:     cardType,
		OwnerName:    ownerName,
		PlateNumber:  plateNumber,
		SpotID:       spotID,
		StartDate:    startDate,
		EndDate:      endDate,
		GraceEndDate: graceEndDate,
		IsActive:     true,
	}

	m.cards[id] = card
	return card, nil
}

func (m *MonthlyCardManager) RenewCard(cardID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	card, exists := m.cards[cardID]
	if !exists {
		return fmt.Errorf("card %s not found", cardID)
	}

	card.Renew()
	return nil
}

func (m *MonthlyCardManager) GetActiveCardForPlate(plateNumber string) *MonthlyCard {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.findActiveCardByPlateLocked(plateNumber)
}

func (m *MonthlyCardManager) findActiveCardByPlateLocked(plateNumber string) *MonthlyCard {
	for _, card := range m.cards {
		if card.PlateNumber == plateNumber && card.GetIsActive() {
			return card
		}
	}
	return nil
}

func (m *MonthlyCardManager) ListCards(plateNumber string, activeOnly bool) []*MonthlyCard {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*MonthlyCard
	for _, card := range m.cards {
		matchesPlate := plateNumber == "" || card.PlateNumber == plateNumber
		matchesActive := !activeOnly || card.GetIsActive()
		if matchesPlate && matchesActive {
			result = append(result, card)
		}
	}
	return result
}

func (m *MonthlyCardManager) CheckAndExpireCards(now time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, card := range m.cards {
		if card.GetIsActive() && now.After(card.GraceEndDate) {
			card.SetIsActive(false)
		}
	}
}

func (m *MonthlyCardManager) GetCardByID(id string) (*MonthlyCard, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	card, ok := m.cards[id]
	return card, ok
}

func firstDayOfNextMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	loc := t.Location()
	nextMonth := time.Date(y, m+1, 1, 0, 0, 0, 0, loc)
	return nextMonth
}
