package server

import (
	"errors"
	"fmt"
	"return-exchange/pkg/common"
	"return-exchange/pkg/models"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOrderNotFound            = errors.New("order not found")
	ErrApplicationExists        = errors.New("application already exists for this order")
	ErrApplicationNotFound      = errors.New("application not found")
	ErrEvidenceRequired         = errors.New("evidence images required for this reason")
	ErrInvalidStatus            = errors.New("invalid application status")
	ErrRefundExceedsOrder       = errors.New("refund amount cannot exceed order amount")
	ErrPriceDifferenceNotHandled = errors.New("price difference must be handled before creating shipping order")
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) generateID() string {
	return uuid.New().String()
}

func (s *Service) requiresEvidence(reason models.ReturnReason) bool {
	return reason == models.ReasonQualityIssue || reason == models.ReasonDamaged
}

func (s *Service) getShippingFeePayer(reason models.ReturnReason) string {
	switch reason {
	case models.ReasonQualityIssue, models.ReasonDamaged, models.ReasonDescriptionMismatch:
		return "platform"
	case models.ReasonNotWanted, models.ReasonWrongOrder:
		return "buyer"
	default:
		return "buyer"
	}
}

func (s *Service) SubmitReturn(req common.SubmitReturnRequest) (string, error) {
	_, exists := s.store.GetOrder(req.OrderID)
	if !exists {
		return "", ErrOrderNotFound
	}

	existingApp, exists := s.store.GetApplicationByOrderID(req.OrderID)
	if exists {
		if existingApp.Status == models.StatusCompleted || existingApp.Status == models.StatusApproved || existingApp.Status == models.StatusProcessing {
			return "", ErrApplicationExists
		}
	}

	if s.requiresEvidence(req.Reason) && len(req.EvidenceImages) == 0 {
		return "", ErrEvidenceRequired
	}

	app := &models.ReturnExchangeApplication{
		ID:               s.generateID(),
		OrderID:          req.OrderID,
		UserID:           req.UserID,
		Type:             models.ReturnType,
		Reason:           req.Reason,
		EvidenceImages:   req.EvidenceImages,
		Status:           models.StatusPending,
		ShippingFeePaidBy: s.getShippingFeePayer(req.Reason),
		CreateTime:       time.Now(),
		UpdateTime:       time.Now(),
	}

	s.store.SaveApplication(app)
	return app.ID, nil
}

func (s *Service) SubmitExchange(req common.SubmitExchangeRequest) (string, error) {
	order, exists := s.store.GetOrder(req.OrderID)
	if !exists {
		return "", ErrOrderNotFound
	}

	existingApp, exists := s.store.GetApplicationByOrderID(req.OrderID)
	if exists {
		if existingApp.Status == models.StatusCompleted || existingApp.Status == models.StatusApproved || existingApp.Status == models.StatusProcessing {
			return "", ErrApplicationExists
		}
	}

	if s.requiresEvidence(req.Reason) && len(req.EvidenceImages) == 0 {
		return "", ErrEvidenceRequired
	}

	priceDifference := order.Price - req.NewSKUPrice

	app := &models.ReturnExchangeApplication{
		ID:                s.generateID(),
		OrderID:           req.OrderID,
		UserID:            req.UserID,
		Type:              models.ExchangeType,
		Reason:            req.Reason,
		EvidenceImages:    req.EvidenceImages,
		NewSKU:            req.NewSKU,
		NewSpecification:  req.NewSpecification,
		NewSKUPrice:       req.NewSKUPrice,
		Status:            models.StatusPending,
		ShippingFeePaidBy: s.getShippingFeePayer(req.Reason),
		PriceDifference:   priceDifference,
		CreateTime:        time.Now(),
		UpdateTime:        time.Now(),
	}

	s.store.SaveApplication(app)
	return app.ID, nil
}

func (s *Service) ReviewApplication(req common.ReviewRequest) error {
	app, exists := s.store.GetApplication(req.ApplicationID)
	if !exists {
		return ErrApplicationNotFound
	}

	if app.Status != models.StatusPending {
		return ErrInvalidStatus
	}

	if req.Approved {
		app.Status = models.StatusApproved
	} else {
		app.Status = models.StatusRejected
	}

	app.ShippingFee = req.ShippingFee
	app.UpdateTime = time.Now()
	s.store.SaveApplication(app)

	return nil
}

func (s *Service) ProcessRefund(req common.ProcessRefundRequest) (string, float64, error) {
	app, exists := s.store.GetApplication(req.ApplicationID)
	if !exists {
		return "", 0, ErrApplicationNotFound
	}

	if app.Status != models.StatusApproved {
		return "", 0, ErrInvalidStatus
	}

	if app.Type != models.ReturnType {
		return "", 0, errors.New("only return applications can be processed for refund")
	}

	order, exists := s.store.GetOrder(app.OrderID)
	if !exists {
		return "", 0, ErrOrderNotFound
	}

	refundAmount := order.Price
	var shippingFeeRefund float64

	if app.ShippingFeePaidBy == "platform" {
		shippingFeeRefund = app.ShippingFee
		if shippingFeeRefund == 0 {
			shippingFeeRefund = order.ShippingFee
		}
	}

	totalRefund := refundAmount + shippingFeeRefund

	if totalRefund > order.TotalAmount {
		totalRefund = order.TotalAmount
		shippingFeeRefund = totalRefund - refundAmount
	}

	refundRecord := &models.RefundRecord{
		ID:            s.generateID(),
		ApplicationID: app.ID,
		OrderID:       app.OrderID,
		UserID:        app.UserID,
		Amount:        refundAmount,
		ShippingFee:   shippingFeeRefund,
		TotalRefund:   totalRefund,
		CreateTime:    time.Now(),
	}

	s.store.SaveRefundRecord(refundRecord)

	app.RefundID = refundRecord.ID
	app.RefundAmount = totalRefund
	app.Status = models.StatusProcessing
	app.UpdateTime = time.Now()
	s.store.SaveApplication(app)

	return refundRecord.ID, totalRefund, nil
}

func (s *Service) CreateShippingOrder(req common.CreateShippingOrderRequest) (string, error) {
	app, exists := s.store.GetApplication(req.ApplicationID)
	if !exists {
		return "", ErrApplicationNotFound
	}

	if app.Status != models.StatusApproved && app.Status != models.StatusProcessing {
		return "", ErrInvalidStatus
	}

	if app.Type != models.ExchangeType {
		return "", errors.New("only exchange applications can create shipping orders")
	}

	if app.PriceDifference != 0 && !app.PriceDifferenceHandled {
		return "", ErrPriceDifferenceNotHandled
	}

	shippingOrder := &models.ShippingOrder{
		ID:               s.generateID(),
		ApplicationID:    app.ID,
		OrderID:          app.OrderID,
		UserID:           app.UserID,
		NewSKU:           app.NewSKU,
		NewSpecification: app.NewSpecification,
		CreateTime:       time.Now(),
	}

	s.store.SaveShippingOrder(shippingOrder)

	app.ShippingOrderID = shippingOrder.ID
	app.Status = models.StatusProcessing
	app.UpdateTime = time.Now()
	s.store.SaveApplication(app)

	return shippingOrder.ID, nil
}

func (s *Service) HandlePriceDifference(req common.HandlePriceDifferenceRequest) (string, float64, string, error) {
	app, exists := s.store.GetApplication(req.ApplicationID)
	if !exists {
		return "", 0, "", ErrApplicationNotFound
	}

	if app.Status != models.StatusApproved {
		return "", 0, "", ErrInvalidStatus
	}

	if app.Type != models.ExchangeType {
		return "", 0, "", errors.New("only exchange applications need price difference handling")
	}

	if app.PriceDifferenceHandled {
		return "", 0, "", errors.New("price difference already handled")
	}

	if app.PriceDifference == 0 {
		app.PriceDifferenceHandled = true
		app.UpdateTime = time.Now()
		s.store.SaveApplication(app)
		return "no_action", 0, "", nil
	}

	priceDifference := app.PriceDifference
	var action string
	var refundID string

	if priceDifference > 0 {
		action = "refund_difference"
		refundAmount := priceDifference

		refundRecord := &models.RefundRecord{
			ID:            s.generateID(),
			ApplicationID: app.ID,
			OrderID:       app.OrderID,
			UserID:        app.UserID,
			Amount:        refundAmount,
			TotalRefund:   refundAmount,
			CreateTime:    time.Now(),
		}

		s.store.SaveRefundRecord(refundRecord)

		app.PriceDifferenceRefundID = refundRecord.ID
		refundID = refundRecord.ID
	} else {
		action = "require_payment"
		priceDifference = -priceDifference
	}

	app.PriceDifferenceHandled = true
	app.UpdateTime = time.Now()
	s.store.SaveApplication(app)

	return action, priceDifference, refundID, nil
}

func (s *Service) CompleteApplication(applicationID string) error {
	app, exists := s.store.GetApplication(applicationID)
	if !exists {
		return ErrApplicationNotFound
	}

	if app.Status != models.StatusProcessing {
		return ErrInvalidStatus
	}

	app.Status = models.StatusCompleted
	app.UpdateTime = time.Now()
	s.store.SaveApplication(app)

	return nil
}

func (s *Service) GetApplication(applicationID string) (*models.ReturnExchangeApplication, error) {
	app, exists := s.store.GetApplication(applicationID)
	if !exists {
		return nil, ErrApplicationNotFound
	}
	return app, nil
}

func (s *Service) ListAllApplications() []models.ReturnExchangeApplication {
	return s.store.ListApplications()
}

func (s *Service) ListUserApplications(userID string) []models.ReturnExchangeApplication {
	return s.store.ListApplicationsByUserID(userID)
}

func (s *Service) CreateMockOrder() string {
	orderID := s.generateID()
	order := &models.Order{
		ID:            orderID,
		UserID:        "user_001",
		SKU:           "SKU_12345",
		Specification: "红色 / L码",
		Price:         299.0,
		ShippingFee:   15.0,
		TotalAmount:   314.0,
		CreateTime:    time.Now(),
	}
	s.store.SaveOrder(order)
	fmt.Printf("Created mock order: ID=%s, Price=%.2f, Total=%.2f\n", orderID, order.Price, order.TotalAmount)
	return orderID
}
