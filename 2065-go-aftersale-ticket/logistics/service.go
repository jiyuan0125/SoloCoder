package logistics

import (
	"crypto/rand"
	"encoding/hex"

	"aftersale-ticket/db"
	"aftersale-ticket/models"
)

type Service struct {
	store *db.Store
}

func NewService(store *db.Store) *Service {
	return &Service{store: store}
}

func generateTrackingNo() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "SF" + hex.EncodeToString(b)
}

func (s *Service) CreateReturnOrder(ticketID string) (*models.LogisticsOrder, error) {
	lo := &models.LogisticsOrder{
		ID:         "LG" + generateTrackingNo()[2:],
		TicketID:   ticketID,
		Type:       "return",
		Carrier:    "顺丰速运",
		TrackingNo: generateTrackingNo(),
		Status:     "created",
	}

	if err := s.store.CreateLogisticsOrder(lo); err != nil {
		return nil, err
	}

	return lo, nil
}

func (s *Service) CreateExchangeOrder(ticketID string) (*models.LogisticsOrder, error) {
	lo := &models.LogisticsOrder{
		ID:         "LG" + generateTrackingNo()[2:],
		TicketID:   ticketID,
		Type:       "exchange",
		Carrier:    "顺丰速运",
		TrackingNo: generateTrackingNo(),
		Status:     "created",
	}

	if err := s.store.CreateLogisticsOrder(lo); err != nil {
		return nil, err
	}

	return lo, nil
}

func (s *Service) GetByTicket(ticketID string) (*models.LogisticsOrder, error) {
	return s.store.GetLogisticsByTicket(ticketID)
}
