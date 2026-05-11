package lockservice

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	MasterManager  *MasterManager
	OrderManager   *OrderManager
	Dispatcher     *Dispatcher
	PaymentManager *PaymentManager
}

func NewService() *Service {
	mm := NewMasterManager()
	return &Service{
		MasterManager:  mm,
		OrderManager:   NewOrderManager(),
		Dispatcher:     NewDispatcher(mm),
		PaymentManager: NewPaymentManager(),
	}
}

type CreateOrderRequest struct {
	UserID       string
	Address      string
	LockType     LockType
	Urgency      Urgency
	TimeSlot     TimeSlot
	TimeSlotDate time.Time
}

type CreateOrderResponse struct {
	Success     bool
	OrderID     string
	Order       *Order
	Alternative *AlternativeSlots
	Message     string
}

func (s *Service) CreateOrder(req *CreateOrderRequest) *CreateOrderResponse {
	order := &Order{
		ID:           uuid.New().String(),
		UserID:       req.UserID,
		Address:      req.Address,
		LockType:     req.LockType,
		Urgency:      req.Urgency,
		TimeSlot:     req.TimeSlot,
		TimeSlotDate: req.TimeSlotDate,
		Status:       OrderStatusPending,
		CreatedAt:    time.Now(),
	}

	s.PaymentManager.CalculateCost(order)

	result := s.Dispatcher.Dispatch(order)
	if !result.Success {
		return &CreateOrderResponse{
			Success:     false,
			Alternative: result.Alternative,
			Message:     "No available master found for selected slot",
		}
	}

	order.MasterID = result.Master.ID
	s.OrderManager.CreateOrder(order)

	if req.Urgency == UrgencyNormal {
		s.MasterManager.BookSlot(result.Master.ID, req.TimeSlotDate, req.TimeSlot)
	}

	s.Dispatcher.ReleaseMaster(result.Master.ID)

	return &CreateOrderResponse{
		Success: true,
		OrderID: order.ID,
		Order:   order,
	}
}

func (s *Service) AcceptOrder(orderID string) error {
	order := s.OrderManager.GetOrder(orderID)
	if order == nil {
		return fmt.Errorf("order not found")
	}
	if order.Status != OrderStatusPending {
		return fmt.Errorf("order cannot be accepted")
	}
	s.OrderManager.UpdateStatus(orderID, OrderStatusAccepted)
	s.MasterManager.UpdateStatus(order.MasterID, MasterStatusEnRoute)
	return nil
}

func (s *Service) StartService(orderID string) error {
	order := s.OrderManager.GetOrder(orderID)
	if order == nil {
		return fmt.Errorf("order not found")
	}
	if order.Status != OrderStatusAccepted {
		return fmt.Errorf("service cannot be started")
	}
	s.OrderManager.UpdateStatus(orderID, OrderStatusInService)
	s.MasterManager.UpdateStatus(order.MasterID, MasterStatusBusy)
	return nil
}

type CompleteServiceRequest struct {
	OrderID         string
	NeedReplaceLock bool
	LockBrand       string
	LockModel       string
	LockLevel       string
	PartsCost       float64
}

func (s *Service) CompleteService(req *CompleteServiceRequest) (*Order, error) {
	order := s.OrderManager.GetOrder(req.OrderID)
	if order == nil {
		return nil, fmt.Errorf("order not found")
	}
	if order.Status != OrderStatusInService {
		return nil, fmt.Errorf("service cannot be completed")
	}

	detail := OrderDetail{
		NeedReplaceLock: req.NeedReplaceLock,
		LockBrand:       req.LockBrand,
		LockModel:       req.LockModel,
		LockLevel:       req.LockLevel,
		PartsCost:       req.PartsCost,
	}
	s.OrderManager.UpdateDetail(req.OrderID, detail)

	order = s.OrderManager.GetOrder(req.OrderID)
	s.PaymentManager.CalculateFinalCost(order)
	s.OrderManager.UpdateStatus(req.OrderID, OrderStatusPendingPayment)

	s.MasterManager.UpdateStatus(order.MasterID, MasterStatusIdle)

	if order.Urgency == UrgencyNormal {
		s.MasterManager.CancelSlot(order.MasterID, order.TimeSlotDate, order.TimeSlot)
	}

	return order, nil
}

func (s *Service) PayOrder(orderID string) error {
	order := s.OrderManager.GetOrder(orderID)
	if order == nil {
		return fmt.Errorf("order not found")
	}
	if order.Status != OrderStatusPendingPayment {
		return fmt.Errorf("order cannot be paid")
	}
	if !s.PaymentManager.PayOrder(orderID) {
		return fmt.Errorf("payment failed")
	}
	s.OrderManager.UpdateStatus(orderID, OrderStatusCompleted)
	return nil
}

type RateOrderRequest struct {
	OrderID string
	Rating  int
	Comment string
}

type RateOrderResponse struct {
	Success      bool
	HasComplaint bool
	Message      string
}

func (s *Service) RateOrder(req *RateOrderRequest) *RateOrderResponse {
	order := s.OrderManager.GetOrder(req.OrderID)
	if order == nil {
		return &RateOrderResponse{Success: false, Message: "order not found"}
	}
	if order.Status != OrderStatusCompleted {
		return &RateOrderResponse{Success: false, Message: "order not completed"}
	}
	if !s.PaymentManager.IsPaid(req.OrderID) {
		return &RateOrderResponse{Success: false, Message: "order not paid"}
	}

	ok, hasComplaint := s.OrderManager.AddRating(req.OrderID, req.Rating, req.Comment)
	if !ok {
		return &RateOrderResponse{Success: false, Message: "invalid rating"}
	}

	if hasComplaint {
		s.PaymentManager.AddComplaint(req.OrderID)
	}

	return &RateOrderResponse{
		Success:      true,
		HasComplaint: hasComplaint,
	}
}

func (s *Service) GetOrder(id string) *Order {
	return s.OrderManager.GetOrder(id)
}

func (s *Service) ListOrders() []*Order {
	return s.OrderManager.ListOrders()
}

func (s *Service) GetMaster(id string) *Master {
	return s.MasterManager.GetMaster(id)
}

func (s *Service) ListMasters() []*Master {
	return s.MasterManager.ListMasters()
}

func (s *Service) AddMaster(master *Master) {
	s.MasterManager.AddMaster(master)
}

func (s *Service) GetComplaints() []string {
	return s.PaymentManager.GetComplaints()
}
