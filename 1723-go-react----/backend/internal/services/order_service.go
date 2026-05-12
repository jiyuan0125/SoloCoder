package services

import (
	"learning-platform/internal/models"
	"learning-platform/internal/storage"
	"learning-platform/internal/utils"
	"time"
)

type OrderService struct {
	storage *storage.Storage
}

func NewOrderService(s *storage.Storage) *OrderService {
	return &OrderService{storage: s}
}

func (s *OrderService) CreateProduct(product *models.Product) (*models.Product, error) {
	if product.ID == "" {
		product.ID = utils.GenerateUUID()
	}
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	err := s.storage.CreateProduct(product)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (s *OrderService) GetProduct(id string) (*models.Product, bool) {
	return s.storage.GetProduct(id)
}

func (s *OrderService) GetProductsPaginated(page, size int) ([]*models.Product, int) {
	products := s.storage.GetAllProducts()
	total := len(products)

	start := (page - 1) * size
	end := start + size

	if start >= total {
		return []*models.Product{}, total
	}
	if end > total {
		end = total
	}

	return products[start:end], total
}

func (s *OrderService) UpdateProductPrice(productID string, newPrice float64, changedBy string) error {
	product, exists := s.storage.GetProduct(productID)
	if !exists {
		return &storage.NotFoundError{ID: productID, Type: "product"}
	}

	history := &models.PriceHistory{
		ID:        utils.GenerateUUID(),
		ProductID: productID,
		OldPrice:  product.Price,
		NewPrice:  newPrice,
		ChangedAt: time.Now(),
		ChangedBy: changedBy,
	}
	s.storage.CreatePriceHistory(history)

	product.Price = newPrice
	product.UpdatedAt = time.Now()
	s.storage.UpdateProduct(product)

	return nil
}

func (s *OrderService) GetPriceHistory(productID string) []*models.PriceHistory {
	return s.storage.GetPriceHistoryByProduct(productID)
}

func (s *OrderService) CreateOrder(studentID string, items []struct {
	ProductID string
	Quantity  int
}) (*models.Order, error) {
	orderItems := []models.OrderItem{}
	totalAmount := 0.0

	for _, item := range items {
		product, exists := s.storage.GetProduct(item.ProductID)
		if !exists {
			return nil, &storage.NotFoundError{ID: item.ProductID, Type: "product"}
		}

		if product.Stock < item.Quantity {
			return nil, &InsufficientStockError{ProductID: item.ProductID, Available: product.Stock, Requested: item.Quantity}
		}

		itemTotal := float64(item.Quantity) * product.Price
		orderItems = append(orderItems, models.OrderItem{
			ID:         utils.GenerateUUID(),
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  product.Price,
			TotalPrice: itemTotal,
		})
		totalAmount += itemTotal
	}

	order := &models.Order{
		ID:          utils.GenerateUUID(),
		StudentID:   studentID,
		OrderItems:  orderItems,
		TotalAmount: totalAmount,
		Status:      "pending",
		CreatedAt:   time.Now(),
		ConfirmedAt: nil,
	}

	err := s.storage.CreateOrder(order)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *OrderService) ConfirmOrder(orderID string) error {
	order, exists := s.storage.GetOrder(orderID)
	if !exists {
		return &storage.NotFoundError{ID: orderID, Type: "order"}
	}

	if order.Status != "pending" {
		return &InvalidOrderStatusError{Expected: "pending", Actual: order.Status}
	}

	for _, item := range order.OrderItems {
		product, _ := s.storage.GetProduct(item.ProductID)
		if product.Stock < item.Quantity {
			return &InsufficientStockError{ProductID: item.ProductID, Available: product.Stock, Requested: item.Quantity}
		}

		product.Stock -= item.Quantity
		product.UpdatedAt = time.Now()
		s.storage.UpdateProduct(product)

		if product.Stock < product.SafetyStock {
			request := &models.PurchaseRequest{
				ID:        utils.GenerateUUID(),
				ProductID: product.ID,
				Quantity:  product.SafetyStock * 2,
				Reason:    "stock below safety level",
				Status:    "pending",
				CreatedAt: time.Now(),
			}
			s.storage.CreatePurchaseRequest(request)
		}
	}

	now := time.Now()
	order.Status = "confirmed"
	order.ConfirmedAt = &now
	s.storage.UpdateOrder(order)

	return nil
}

func (s *OrderService) GetOrder(id string) (*models.Order, bool) {
	return s.storage.GetOrder(id)
}

func (s *OrderService) GetOrdersPaginated(page, size int) ([]*models.Order, int) {
	orders := s.storage.GetAllOrders()
	total := len(orders)

	start := (page - 1) * size
	end := start + size

	if start >= total {
		return []*models.Order{}, total
	}
	if end > total {
		end = total
	}

	return orders[start:end], total
}

func (s *OrderService) GetAllPurchaseRequests() []*models.PurchaseRequest {
	return s.storage.GetAllPurchaseRequests()
}

type InsufficientStockError struct {
	ProductID string
	Available int
	Requested int
}

func (e *InsufficientStockError) Error() string {
	return "insufficient stock"
}

type InvalidOrderStatusError struct {
	Expected string
	Actual   string
}

func (e *InvalidOrderStatusError) Error() string {
	return "invalid order status"
}
