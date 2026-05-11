package core

import (
	"errors"
	"time"

	"renovation-management/internal/api"
)

func (s *Store) CreateChangeOrder(req *api.CreateChangeOrderRequest) (*api.ChangeOrder, error) {
	if req.QuotationID == "" {
		return nil, errors.New("quotation_id is required")
	}
	if req.Content == "" {
		return nil, errors.New("change content is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.quotations[req.QuotationID]; !ok {
		return nil, errors.New("quotation not found")
	}

	changeOrder := &api.ChangeOrder{
		ID:            GenerateID("CHANGE"),
		QuotationID:   req.QuotationID,
		Content:       req.Content,
		Reason:        req.Reason,
		AmountDiffFen: req.AmountDiffFen,
		Confirmed:     false,
		CreatedAt:     time.Now(),
	}

	s.changeOrders[changeOrder.ID] = changeOrder
	s.changeOrdersByQuotation[req.QuotationID] = append(
		s.changeOrdersByQuotation[req.QuotationID],
		changeOrder,
	)

	return changeOrder, nil
}

func (s *Store) ConfirmChangeOrder(changeOrderID string) (*api.ChangeOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	changeOrder, ok := s.changeOrders[changeOrderID]
	if !ok {
		return nil, errors.New("change order not found")
	}
	if changeOrder.Confirmed {
		return nil, errors.New("change order already confirmed")
	}

	changeOrder.Confirmed = true
	return changeOrder, nil
}

func (s *Store) GetChangeOrder(id string) (*api.ChangeOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	changeOrder, ok := s.changeOrders[id]
	if !ok {
		return nil, errors.New("change order not found")
	}
	return changeOrder, nil
}

func (s *Store) ListChangeOrdersByQuotation(quotationID string) []*api.ChangeOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.changeOrdersByQuotation[quotationID]
}

func (s *Store) CalculateSettlement(quotationID string) (*api.Settlement, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	quotation, ok := s.quotations[quotationID]
	if !ok {
		return nil, errors.New("quotation not found")
	}

	if _, exists := s.settlementByQuotation[quotationID]; exists {
		return nil, errors.New("settlement already exists for this quotation")
	}

	var confirmedDiffTotal int64 = 0
	for _, co := range s.changeOrdersByQuotation[quotationID] {
		if co.Confirmed {
			confirmedDiffTotal += co.AmountDiffFen
		}
	}

	settlement := &api.Settlement{
		ID:               GenerateID("SETTLE"),
		QuotationID:      quotationID,
		OriginalTotalFen: quotation.TotalFen,
		ChangeTotalFen: confirmedDiffTotal,
		FinalTotalFen:  quotation.TotalFen + confirmedDiffTotal,
		CreatedAt:      time.Now(),
	}

	s.settlements[settlement.ID] = settlement
	s.settlementByQuotation[quotationID] = settlement

	return settlement, nil
}

func (s *Store) GetSettlement(id string) (*api.Settlement, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	settlement, ok := s.settlements[id]
	if !ok {
		return nil, errors.New("settlement not found")
	}
	return settlement, nil
}

func (s *Store) GetSettlementByQuotation(quotationID string) (*api.Settlement, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	settlement, ok := s.settlementByQuotation[quotationID]
	if !ok {
		return nil, errors.New("settlement not found")
	}
	return settlement, nil
}
