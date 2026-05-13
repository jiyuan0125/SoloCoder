package inbound

import (
	"errors"
	"fmt"
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

func (s *Service) CreateOrder(db *models.Database, productID string, quantity int, manufactureDate time.Time) (*models.InboundOrder, error) {
	if productID == "" {
		return nil, errors.New("product id is required")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be greater than 0")
	}
	if db.GetProductByID(productID) == nil {
		return nil, errors.New("product not found")
	}

	order := &models.InboundOrder{
		ID:              models.GenerateID("IN"),
		ProductID:       productID,
		Quantity:        quantity,
		ManufactureDate: manufactureDate,
		Status:          "pending",
		CreatedAt:       time.Now(),
	}
	db.InboundOrders = append(db.InboundOrders, *order)
	return order, nil
}

func (s *Service) ProcessOrder(db *models.Database, orderID string) (*models.InboundOrder, error) {
	var order *models.InboundOrder
	for i := range db.InboundOrders {
		if db.InboundOrders[i].ID == orderID {
			order = &db.InboundOrders[i]
			break
		}
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	if order.Status != "pending" {
		return nil, errors.New("order already processed")
	}

	locationCode, err := s.locationSvc.AllocateLocation(db, order.ProductID, order.Quantity)
	if err != nil {
		return nil, err
	}

	if err := s.locationSvc.UpdateLocationUsage(db, locationCode, order.Quantity); err != nil {
		return nil, err
	}

	batchID := models.GenerateID("B")
	batch := models.Batch{
		ID:              batchID,
		ProductID:       order.ProductID,
		LocationCode:    locationCode,
		ManufactureDate: order.ManufactureDate,
		Quantity:        order.Quantity,
		CreatedAt:       time.Now(),
	}
	db.Batches = append(db.Batches, batch)

	db.UpdateStockRecord(order.ProductID, locationCode, order.Quantity)

	now := time.Now()
	order.BatchID = batchID
	order.LocationCode = locationCode
	order.Status = "processed"
	order.ProcessedAt = &now

	s.syncRelatedRecords(db, order.ProductID)

	fmt.Printf("[INBOUND] Processed order %s: product=%s, qty=%d, location=%s\n",
		orderID, order.ProductID, order.Quantity, locationCode)

	return order, nil
}

func (s *Service) syncRelatedRecords(db *models.Database, productID string) {
	for i := range db.InboundOrders {
		order := &db.InboundOrders[i]
		if order.ProductID == productID && order.Status == "pending" {
			db2 := *db
			if db2.GetProductByID(productID) != nil {
				now := time.Now()
				order.ProcessedAt = &now
			}
		}
	}
}
