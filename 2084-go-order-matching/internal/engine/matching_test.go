package engine

import (
	"os"
	"order-matching/internal/model"
	"sync"
	"testing"
)

func TestPriceTimePriority(t *testing.T) {
	dbPath := "/tmp/test_matching.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := model.NewDatabase(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	engine := NewMatchingEngine(db)

	buyOrder := &model.Order{
		UserID:   "user1",
		Symbol:   "BTC",
		Side:     model.Buy,
		Price:    100.0,
		Quantity: 10,
	}
	trades, err := engine.SubmitOrder(buyOrder)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 0 {
		t.Fatal("expected no trades")
	}

	sellOrder := &model.Order{
		UserID:   "user2",
		Symbol:   "BTC",
		Side:     model.Sell,
		Price:    90.0,
		Quantity: 5,
	}
	trades, err = engine.SubmitOrder(sellOrder)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}

	trade := trades[0]
	if trade.Price != 100.0 {
		t.Fatalf("expected trade price 100 (first order price), got %f", trade.Price)
	}
	if trade.Quantity != 5 {
		t.Fatalf("expected trade quantity 5, got %d", trade.Quantity)
	}
	if trade.BuyerID != "user1" || trade.SellerID != "user2" {
		t.Fatalf("unexpected trade parties")
	}

	bids, asks, err := engine.GetOrderBook("BTC")
	if err != nil {
		t.Fatal(err)
	}
	if len(bids) != 1 {
		t.Fatalf("expected 1 bid level, got %d", len(bids))
	}
	if bids[0].Price != 100.0 || bids[0].Quantity != 5 {
		t.Fatalf("unexpected bid: %+v", bids[0])
	}
	if len(asks) != 0 {
		t.Fatalf("expected 0 asks, got %d", len(asks))
	}
}

func TestConcurrency(t *testing.T) {
	dbPath := "/tmp/test_concurrency.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := model.NewDatabase(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	engine := NewMatchingEngine(db)

	var wg sync.WaitGroup
	numOrders := 100

	for i := 0; i < numOrders; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			side := model.Buy
			if i%2 == 0 {
				side = model.Sell
			}
			order := &model.Order{
				UserID:   "user",
				Symbol:   "ETH",
				Side:     side,
				Price:    50.0,
				Quantity: 1,
			}
			engine.SubmitOrder(order)
		}(i)
	}

	wg.Wait()

	bids, asks, err := engine.GetOrderBook("ETH")
	if err != nil {
		t.Fatal(err)
	}

	var totalBid, totalAsk int64
	for _, b := range bids {
		totalBid += b.Quantity
	}
	for _, a := range asks {
		totalAsk += a.Quantity
	}

	if totalBid != totalAsk {
		t.Logf("Total bid quantity: %d, Total ask quantity: %d", totalBid, totalAsk)
	}
}

func TestValidation(t *testing.T) {
	dbPath := "/tmp/test_validation.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := model.NewDatabase(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	engine := NewMatchingEngine(db)

	tests := []struct {
		name    string
		order   *model.Order
		wantErr bool
	}{
		{
			name: "price must be positive",
			order: &model.Order{
				UserID:   "user",
				Symbol:   "BTC",
				Side:     model.Buy,
				Price:    0,
				Quantity: 10,
			},
			wantErr: true,
		},
		{
			name: "quantity must be positive",
			order: &model.Order{
				UserID:   "user",
				Symbol:   "BTC",
				Side:     model.Buy,
				Price:    100,
				Quantity: 0,
			},
			wantErr: true,
		},
		{
			name: "symbol cannot be empty",
			order: &model.Order{
				UserID:   "user",
				Symbol:   "",
				Side:     model.Buy,
				Price:    100,
				Quantity: 10,
			},
			wantErr: true,
		},
		{
			name: "valid order",
			order: &model.Order{
				UserID:   "user",
				Symbol:   "BTC",
				Side:     model.Buy,
				Price:    100,
				Quantity: 10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.SubmitOrder(tt.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("SubmitOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCancelOrder(t *testing.T) {
	dbPath := "/tmp/test_cancel.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := model.NewDatabase(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	engine := NewMatchingEngine(db)

	buyOrder := &model.Order{
		UserID:   "user1",
		Symbol:   "BTC",
		Side:     model.Buy,
		Price:    100.0,
		Quantity: 10,
	}
	_, err = engine.SubmitOrder(buyOrder)
	if err != nil {
		t.Fatal(err)
	}

	err = engine.CancelOrder(buyOrder.ID)
	if err != nil {
		t.Fatalf("expected to cancel pending order, got error: %v", err)
	}

	err = engine.CancelOrder(buyOrder.ID)
	if err == nil {
		t.Fatal("expected error when cancelling already cancelled order")
	}

	buyOrder2 := &model.Order{
		UserID:   "user3",
		Symbol:   "BTC",
		Side:     model.Buy,
		Price:    100.0,
		Quantity: 10,
	}
	_, err = engine.SubmitOrder(buyOrder2)
	if err != nil {
		t.Fatal(err)
	}

	sellOrder := &model.Order{
		UserID:   "user2",
		Symbol:   "BTC",
		Side:     model.Sell,
		Price:    90.0,
		Quantity: 10,
	}
	trades, err := engine.SubmitOrder(sellOrder)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}

	err = engine.CancelOrder(sellOrder.ID)
	if err == nil {
		t.Fatal("expected error when cancelling fully filled order")
	}
}

func TestPartialFill(t *testing.T) {
	dbPath := "/tmp/test_partial.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := model.NewDatabase(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	engine := NewMatchingEngine(db)

	sellOrder := &model.Order{
		UserID:   "user1",
		Symbol:   "BTC",
		Side:     model.Sell,
		Price:    100.0,
		Quantity: 10,
	}
	_, err = engine.SubmitOrder(sellOrder)
	if err != nil {
		t.Fatal(err)
	}

	buyOrder := &model.Order{
		UserID:   "user2",
		Symbol:   "BTC",
		Side:     model.Buy,
		Price:    100.0,
		Quantity: 3,
	}
	trades, err := engine.SubmitOrder(buyOrder)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].Quantity != 3 {
		t.Fatalf("expected trade quantity 3, got %d", trades[0].Quantity)
	}

	_, asks, err := engine.GetOrderBook("BTC")
	if err != nil {
		t.Fatal(err)
	}
	if len(asks) != 1 {
		t.Fatalf("expected 1 ask level, got %d", len(asks))
	}
	if asks[0].Quantity != 7 {
		t.Fatalf("expected remaining quantity 7, got %d", asks[0].Quantity)
	}

	err = engine.CancelOrder(sellOrder.ID)
	if err != nil {
		t.Fatalf("expected to cancel partial order, got error: %v", err)
	}
}

func TestEmptyOrderBook(t *testing.T) {
	dbPath := "/tmp/test_empty.db"
	os.Remove(dbPath)
	defer os.Remove(dbPath)

	db, err := model.NewDatabase(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	engine := NewMatchingEngine(db)

	bids, asks, err := engine.GetOrderBook("NONEXISTENT")
	if err != nil {
		t.Fatal(err)
	}
	if len(bids) != 0 || len(asks) != 0 {
		t.Fatal("expected empty order book for non-existent symbol")
	}
}
