package core

import (
	"errors"
	"time"
)

var (
	ErrStocktakingNotFound      = errors.New("stocktaking not found")
	ErrStocktakingNotPending    = errors.New("stocktaking is not in pending status")
	ErrStocktakingNotSubmitted  = errors.New("stocktaking is not in submitted status")
	ErrStocktakingAlreadyActive = errors.New("there is already an active stocktaking")
	ErrProductNotInStocktaking  = errors.New("product not found in stocktaking")
)

func (s *Store) CreateStocktaking(operatorID string) (*Stocktaking, error) {
	s.Lock()
	defer s.Unlock()

	for _, st := range s.stocktakings {
		if st.Status != StocktakingStatusCompleted {
			return nil, ErrStocktakingAlreadyActive
		}
	}

	items := make([]StocktakingItem, 0, len(s.products))
	for _, product := range s.products {
		items = append(items, StocktakingItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			Barcode:     product.Barcode,
			Price:       product.Price,
			SnapshotQty: product.StockQty,
			ActualQty:   0,
			DiffQty:     0,
			DiffAmount:  0,
		})
	}

	id := s.idGen.Generate()
	stocktaking := &Stocktaking{
		ID:         id,
		Status:     StocktakingStatusPending,
		OperatorID: operatorID,
		CreatedAt:  time.Now(),
		Items:      items,
	}

	s.stocktakings[id] = stocktaking
	s.pendingOps[id] = make([]*PendingOperation, 0)

	return stocktaking, nil
}

func (s *Store) GetAllStocktakings() []Stocktaking {
	s.RLock()
	defer s.RUnlock()

	result := make([]Stocktaking, 0, len(s.stocktakings))
	for _, st := range s.stocktakings {
		copy := *st
		copy.Items = nil
		result = append(result, copy)
	}
	return result
}

func (s *Store) GetStocktaking(id string) (*Stocktaking, error) {
	s.RLock()
	defer s.RUnlock()

	st, exists := s.stocktakings[id]
	if !exists {
		return nil, ErrStocktakingNotFound
	}

	stCopy := *st
	stCopy.Items = make([]StocktakingItem, len(st.Items))
	copy(stCopy.Items, st.Items)

	return &stCopy, nil
}

func (s *Store) SubmitStocktaking(id string, actualQtys map[string]int) error {
	s.Lock()
	defer s.Unlock()

	st, exists := s.stocktakings[id]
	if !exists {
		return ErrStocktakingNotFound
	}

	if st.Status != StocktakingStatusPending {
		return ErrStocktakingNotPending
	}

	for i, item := range st.Items {
		if actualQty, ok := actualQtys[item.ProductID]; ok {
			st.Items[i].ActualQty = actualQty
			st.Items[i].DiffQty = actualQty - item.SnapshotQty
			st.Items[i].DiffAmount = int64(st.Items[i].DiffQty) * item.Price
		}
	}

	st.Status = StocktakingStatusSubmitted
	return nil
}

func (s *Store) CompleteStocktaking(id string) error {
	s.Lock()
	defer s.Unlock()

	st, exists := s.stocktakings[id]
	if !exists {
		return ErrStocktakingNotFound
	}

	if st.Status != StocktakingStatusSubmitted {
		return ErrStocktakingNotSubmitted
	}

	for _, item := range st.Items {
		if product, exists := s.products[item.ProductID]; exists {
			product.StockQty = item.ActualQty
			product.UpdatedAt = time.Now()
		}
	}

	st.Status = StocktakingStatusCompleted
	st.CompletedAt = time.Now()

	delete(s.pendingOps, id)

	return nil
}

func (s *Store) GetPendingOperations(stocktakingID string) []PendingOperation {
	s.RLock()
	defer s.RUnlock()

	ops, exists := s.pendingOps[stocktakingID]
	if !exists {
		return nil
	}

	result := make([]PendingOperation, len(ops))
	for i, op := range ops {
		result[i] = *op
	}
	return result
}

func (s *Store) StockIn(productID string, quantity int, operatorID string) (string, error) {
	if quantity <= 0 {
		return "", ErrInvalidQuantity
	}

	s.Lock()
	defer s.Unlock()

	product, exists := s.products[productID]
	if !exists {
		return "", ErrProductNotFound
	}

	now := time.Now()
	product.StockQty += quantity
	product.UpdatedAt = now

	stockOpID := s.idGen.Generate()
	stockOp := &StockOperation{
		ID:            stockOpID,
		ProductID:     productID,
		OperationType: "stock_in",
		Quantity:      quantity,
		OperatorID:    operatorID,
		CreatedAt:     now,
	}
	s.stockOps = append(s.stockOps, stockOp)

	if s.hasActiveStocktaking() {
		s.addPendingOperationToActiveStocktaking(productID, quantity, PendingOpTypeStockIn, now)
	}

	return stockOpID, nil
}
