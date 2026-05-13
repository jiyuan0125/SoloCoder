package outbound

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"warehouse-mgmt/internal/models"
	locSvc "warehouse-mgmt/internal/location"
)

type Service struct {
	locationSvc *locSvc.Service
}

func NewService(locationSvc *locSvc.Service) *Service {
	return &Service{locationSvc: locationSvc}
}

func (s *Service) CreateOrder(db *models.Database, productID string, quantity int) (*models.OutboundOrder, error) {
	if productID == "" {
		return nil, errors.New("product id is required")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}
	if db.GetProductByID(productID) == nil {
		return nil, errors.New("product not found")
	}

	totalStock := db.GetTotalStock(productID)
	if totalStock < quantity {
		return nil, fmt.Errorf("insufficient stock: available=%d, requested=%d, shortage=%d",
			totalStock, quantity, quantity-totalStock)
	}

	order := &models.OutboundOrder{
		ID:        models.GenerateID("OUT"),
		ProductID:  productID,
		Quantity:   quantity,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}
	db.OutboundOrders = append(db.OutboundOrders, *order)
	return order, nil
}

func (s *Service) ProcessOrder(db *models.Database, orderID string) (*models.OutboundOrder, error) {
	var order *models.OutboundOrder
	for i := range db.OutboundOrders {
		if db.OutboundOrders[i].ID == orderID {
			order = &db.OutboundOrders[i]
			break
		}
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	if order.Status != "pending" {
		return nil, errors.New("order already processed")
	}

	batches := db.GetProductBatches(order.ProductID)
	if len(batches) == 0 {
		return nil, errors.New("no available batches")
	}

	sort.SliceStable(batches, func(i, j int) bool {
		return batches[i].ManufactureDate.Before(batches[j].ManufactureDate)
	})

	remaining := order.Quantity
	var details []models.OutboundDetail

	for i := 0; remaining > 0 && i < len(batches); i++ {
		batch := &batches[i]
		take := batch.Quantity
		if take > remaining {
			take = remaining
		}

		remaining -= take
		details = append(details, models.OutboundDetail{
			BatchID:      batch.ID,
			Quantity:     take,
			LocationCode: batch.LocationCode,
		})

		if err := s.locationSvc.UpdateLocationUsage(db, batch.LocationCode, -take); err != nil {
			return nil, err
		}

		db.UpdateStockRecord(order.ProductID, batch.LocationCode, -take)
	}

	if remaining > 0 {
		return nil, fmt.Errorf("insufficient stock after allocation")
	}

	for i := range db.Batches {
		for _, d := range details {
			if db.Batches[i].ID == d.BatchID {
				db.Batches[i].Quantity -= d.Quantity
			}
		}
	}

	for i := range db.OutboundOrders {
		if db.OutboundOrders[i].ID == orderID {
			now := time.Now()
			db.OutboundOrders[i].Details = details
			db.OutboundOrders[i].Status = "processed"
			db.OutboundOrders[i].ProcessedAt = &now
			order = &db.OutboundOrders[i]
		}
	}

	s.syncRelatedRecords(db, order.ProductID)

	fmt.Printf("[OUTBOUND] Processed order %s: product=%s, qty=%d\n",
		orderID, order.ProductID, order.Quantity)

	return order, nil
}

func (s *Service) syncRelatedRecords(db *models.Database, productID string) {
	for i := range db.OutboundOrders {
		order := &db.OutboundOrders[i]
		if order.ProductID == productID && order.Status == "pending" {
			db2 := *db
			if db2.GetProductByID(productID) != nil {
				now := time.Now()
				order.ProcessedAt = &now
			}
		}
	}
}
