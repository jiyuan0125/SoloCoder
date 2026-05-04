package server

import (
	"return-exchange/pkg/models"
	"sync"
)

type Store struct {
	orders          map[string]*models.Order
	applications    map[string]*models.ReturnExchangeApplication
	refundRecords   map[string]*models.RefundRecord
	shippingOrders  map[string]*models.ShippingOrder
	orderAppMap     map[string]string
	mu              sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		orders:         make(map[string]*models.Order),
		applications:   make(map[string]*models.ReturnExchangeApplication),
		refundRecords:  make(map[string]*models.RefundRecord),
		shippingOrders: make(map[string]*models.ShippingOrder),
		orderAppMap:    make(map[string]string),
	}
}

func (s *Store) SaveOrder(order *models.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[order.ID] = order
}

func (s *Store) GetOrder(id string) (*models.Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, exists := s.orders[id]
	return order, exists
}

func (s *Store) SaveApplication(app *models.ReturnExchangeApplication) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.applications[app.ID] = app
	s.orderAppMap[app.OrderID] = app.ID
}

func (s *Store) GetApplication(id string) (*models.ReturnExchangeApplication, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	app, exists := s.applications[id]
	return app, exists
}

func (s *Store) GetApplicationByOrderID(orderID string) (*models.ReturnExchangeApplication, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	appID, exists := s.orderAppMap[orderID]
	if !exists {
		return nil, false
	}
	app, exists := s.applications[appID]
	return app, exists
}

func (s *Store) ListApplications() []models.ReturnExchangeApplication {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.ReturnExchangeApplication, 0, len(s.applications))
	for _, app := range s.applications {
		result = append(result, *app)
	}
	return result
}

func (s *Store) ListApplicationsByUserID(userID string) []models.ReturnExchangeApplication {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.ReturnExchangeApplication, 0)
	for _, app := range s.applications {
		if app.UserID == userID {
			result = append(result, *app)
		}
	}
	return result
}

func (s *Store) SaveRefundRecord(record *models.RefundRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refundRecords[record.ID] = record
}

func (s *Store) GetRefundRecord(id string) (*models.RefundRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, exists := s.refundRecords[id]
	return record, exists
}

func (s *Store) SaveShippingOrder(order *models.ShippingOrder) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shippingOrders[order.ID] = order
}

func (s *Store) GetShippingOrder(id string) (*models.ShippingOrder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, exists := s.shippingOrders[id]
	return order, exists
}

func (s *Store) GetAllOrders() map[string]*models.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*models.Order)
	for k, v := range s.orders {
		result[k] = v
	}
	return result
}

func (s *Store) GetAllApplications() map[string]*models.ReturnExchangeApplication {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*models.ReturnExchangeApplication)
	for k, v := range s.applications {
		result[k] = v
	}
	return result
}

func (s *Store) GetAllRefundRecords() map[string]*models.RefundRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*models.RefundRecord)
	for k, v := range s.refundRecords {
		result[k] = v
	}
	return result
}

func (s *Store) GetAllShippingOrders() map[string]*models.ShippingOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]*models.ShippingOrder)
	for k, v := range s.shippingOrders {
		result[k] = v
	}
	return result
}

func (s *Store) LoadOrders(orders map[string]*models.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders = orders
}

func (s *Store) LoadApplications(applications map[string]*models.ReturnExchangeApplication) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.applications = applications
	s.orderAppMap = make(map[string]string)
	for _, app := range applications {
		s.orderAppMap[app.OrderID] = app.ID
	}
}

func (s *Store) LoadRefundRecords(records map[string]*models.RefundRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refundRecords = records
}

func (s *Store) LoadShippingOrders(orders map[string]*models.ShippingOrder) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shippingOrders = orders
}
