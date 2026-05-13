package engine

import (
	"fmt"
	"order-matching/internal/model"
	"sync"
	"time"
)

type MatchingEngine struct {
	db     *model.Database
	mutex  *sync.RWMutex
	symbolLocks map[string]*sync.Mutex
	globalLock sync.Mutex
}

func NewMatchingEngine(db *model.Database) *MatchingEngine {
	return &MatchingEngine{
		db:          db,
		mutex:       &sync.RWMutex{},
		symbolLocks: make(map[string]*sync.Mutex),
	}
}

func (e *MatchingEngine) getSymbolLock(symbol string) *sync.Mutex {
	e.globalLock.Lock()
	defer e.globalLock.Unlock()
	if lock, ok := e.symbolLocks[symbol]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	e.symbolLocks[symbol] = lock
	return lock
}

func (e *MatchingEngine) SubmitOrder(order *model.Order) ([]*model.Trade, error) {
	if order.Price <= 0 {
		return nil, fmt.Errorf("price must be positive")
	}
	if order.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if order.Symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	lock := e.getSymbolLock(order.Symbol)
	lock.Lock()
	defer lock.Unlock()

	var trades []*model.Trade
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		var err error
		trades, err = e.trySubmitOrder(order)
		if err == nil {
			return trades, nil
		}
		if i == maxRetries-1 {
			return nil, err
		}
	}
	return nil, fmt.Errorf("failed to submit order after retries")
}

func (e *MatchingEngine) trySubmitOrder(order *model.Order) ([]*model.Trade, error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	order.Status = model.OrderPending
	order.FilledQuantity = 0
	order.CreatedAt = time.Now()
	order.Version = 0

	if err := e.db.InsertOrder(tx, order); err != nil {
		return nil, err
	}

	var trades []*model.Trade
	remaining := order.Quantity - order.FilledQuantity

	oppositeSide := model.Sell
	if order.Side == model.Sell {
		oppositeSide = model.Buy
	}

	matchingOrders, err := e.db.GetPendingOrders(tx, order.Symbol, oppositeSide)
	if err != nil {
		return nil, err
	}

	for _, opposite := range matchingOrders {
		if remaining <= 0 {
			break
		}

		if order.Side == model.Buy {
			if order.Price < opposite.Price {
				continue
			}
		} else {
			if order.Price > opposite.Price {
				continue
			}
		}

		oppositeRemaining := opposite.Quantity - opposite.FilledQuantity
		if oppositeRemaining <= 0 {
			continue
		}

		tradeQty := min(remaining, oppositeRemaining)

		var tradePrice float64
		if opposite.CreatedAt.Before(order.CreatedAt) {
			tradePrice = opposite.Price
		} else {
			tradePrice = order.Price
		}

		if order.Side == model.Buy {
			if tradePrice > order.Price || tradePrice < opposite.Price {
				continue
			}
		} else {
			if tradePrice < order.Price || tradePrice > opposite.Price {
				continue
			}
		}

		trade := &model.Trade{
			Symbol:    order.Symbol,
			Price:     tradePrice,
			Quantity:  tradeQty,
			CreatedAt: time.Now(),
		}

		if order.Side == model.Buy {
			trade.BuyOrderID = order.ID
			trade.BuyerID = order.UserID
			trade.SellOrderID = opposite.ID
			trade.SellerID = opposite.UserID
		} else {
			trade.SellOrderID = order.ID
			trade.SellerID = order.UserID
			trade.BuyOrderID = opposite.ID
			trade.BuyerID = opposite.UserID
		}

		if err := e.db.InsertTrade(tx, trade); err != nil {
			return nil, err
		}

		trades = append(trades, trade)

		order.FilledQuantity += tradeQty
		if order.FilledQuantity == order.Quantity {
			order.Status = model.OrderFilled
		} else {
			order.Status = model.OrderPartial
		}

		if err := e.db.UpdateOrder(tx, order); err != nil {
			return nil, err
		}

		opposite.FilledQuantity += tradeQty
		if opposite.FilledQuantity == opposite.Quantity {
			opposite.Status = model.OrderFilled
		} else {
			opposite.Status = model.OrderPartial
		}

		if err := e.db.UpdateOrder(tx, opposite); err != nil {
			return nil, err
		}

		remaining -= tradeQty
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return trades, nil
}

func (e *MatchingEngine) CancelOrder(orderID int64) error {
	tx, err := e.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	order, err := e.db.GetOrder(tx, orderID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	if order.Status == model.OrderFilled {
		return fmt.Errorf("cannot cancel fully filled order")
	}

	if order.Status == model.OrderCancelled {
		return fmt.Errorf("order already cancelled")
	}

	order.Status = model.OrderCancelled

	if err := e.db.UpdateOrder(tx, order); err != nil {
		return err
	}

	return tx.Commit()
}

func (e *MatchingEngine) GetOrderBook(symbol string) (bids, asks []*model.OrderBookEntry, err error) {
	tx, err := e.db.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback()

	buyOrders, err := e.db.GetPendingOrders(tx, symbol, model.Buy)
	if err != nil {
		return nil, nil, err
	}

	sellOrders, err := e.db.GetPendingOrders(tx, symbol, model.Sell)
	if err != nil {
		return nil, nil, err
	}

	bids = aggregateOrders(buyOrders, 5)
	asks = aggregateOrders(sellOrders, 5)

	return bids, asks, nil
}

func aggregateOrders(orders []*model.Order, limit int) []*model.OrderBookEntry {
	priceMap := make(map[float64]int64)
	var prices []float64

	for _, order := range orders {
		remaining := order.Quantity - order.FilledQuantity
		if remaining <= 0 {
			continue
		}
		if _, exists := priceMap[order.Price]; !exists {
			prices = append(prices, order.Price)
		}
		priceMap[order.Price] += remaining
	}

	var result []*model.OrderBookEntry
	for i, p := range prices {
		if i >= limit {
			break
		}
		result = append(result, &model.OrderBookEntry{
			Price:    p,
			Quantity: priceMap[p],
		})
	}

	return result
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
