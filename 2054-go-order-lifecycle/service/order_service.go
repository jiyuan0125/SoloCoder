package service

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"order-lifecycle/inventory"
	"order-lifecycle/models"
	"order-lifecycle/payment"
	"order-lifecycle/repository"
	"order-lifecycle/statemachine"
)

var (
	ErrOrderNotFound       = fmt.Errorf("订单不存在")
	ErrInvalidTransition   = fmt.Errorf("状态流转无效")
	ErrInsufficientInventory = fmt.Errorf("库存不足")
	ErrConflict            = fmt.Errorf("并发冲突")
	ErrDuplicatePayment    = fmt.Errorf("重复支付")
	ErrRefundWindowExpired = fmt.Errorf("退货窗口期已过")
	ErrTerminalState       = fmt.Errorf("订单已终止，无法操作")
	ErrPaymentFailed       = fmt.Errorf("支付失败")
)

type OrderService struct {
	repo           *repository.OrderRepository
	inventorySvc   *inventory.InventoryService
	stateMachine   *statemachine.StateMachine
	paymentFactory *payment.GatewayFactory
	paymentLocks   sync.Map
}

func NewOrderService(
	repo *repository.OrderRepository,
	inventorySvc *inventory.InventoryService,
	stateMachine *statemachine.StateMachine,
	paymentFactory *payment.GatewayFactory,
) *OrderService {
	return &OrderService{
		repo:           repo,
		inventorySvc:   inventorySvc,
		stateMachine:   stateMachine,
		paymentFactory: paymentFactory,
		paymentLocks:   sync.Map{},
	}
}

type CreateOrderRequest struct {
	OrderID   string
	UserID    string
	ProductID string
	Quantity  int
	Price     float64
}

func (s *OrderService) CreateOrder(req *CreateOrderRequest) (*models.Order, error) {
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("数量必须大于0")
	}

	if err := s.repo.BeginTx(); err != nil {
		return nil, err
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.Rollback()
		}
	}()

	if err := s.inventorySvc.Deduct(req.ProductID, req.Quantity); err != nil {
		available, _ := s.inventorySvc.GetAvailable(req.ProductID)
		return nil, fmt.Errorf("库存不足，剩余库存: %d", available)
	}

	order := &models.Order{
		ID:          req.OrderID,
		UserID:      req.UserID,
		ProductID:   req.ProductID,
		Quantity:    req.Quantity,
		TotalAmount: req.Price * float64(req.Quantity),
		Status:      models.StatusCreated,
	}

	if err := s.repo.Create(order); err != nil {
		return nil, err
	}

	if err := s.repo.Commit(); err != nil {
		return nil, err
	}
	committed = true

	go s.notifyInventoryDeducted(order)

	return order, nil
}

func (s *OrderService) GetOrder(orderID string) (*models.Order, error) {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

func (s *OrderService) GetOrderHistory(orderID string) ([]models.OrderStatusHistory, error) {
	return s.repo.GetHistory(orderID)
}

type PayRequest struct {
	OrderID       string
	PaymentMethod models.PaymentMethod
}

func (s *OrderService) Pay(req *PayRequest) error {
	lockKey := fmt.Sprintf("payment:%s", req.OrderID)
	if _, loaded := s.paymentLocks.LoadOrStore(lockKey, struct{}{}); loaded {
		return ErrDuplicatePayment
	}
	defer s.paymentLocks.Delete(lockKey)

	order, err := s.repo.GetByID(req.OrderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}

	if order.Status != models.StatusCreated {
		if order.Status == models.StatusPaid {
			return ErrDuplicatePayment
		}
		if s.stateMachine.IsTerminal(order.Status) {
			return ErrTerminalState
		}
		return ErrInvalidTransition
	}

	if err := s.stateMachine.ValidateTransition(order.Status, models.StatusPaid); err != nil {
		return ErrInvalidTransition
	}

	gateway, err := s.paymentFactory.GetGateway(req.PaymentMethod)
	if err != nil {
		return err
	}

	paymentReq := models.PaymentRequest{
		OrderID:       req.OrderID,
		PaymentMethod: req.PaymentMethod,
		Amount:        order.TotalAmount,
	}
	result := gateway.ProcessPayment(paymentReq)

	if !result.Success {
		return fmt.Errorf("%w: %s", ErrPaymentFailed, result.Message)
	}

	if err := s.repo.BeginTx(); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.Rollback()
		}
	}()

	if err := s.repo.UpdatePaymentMethod(order, req.PaymentMethod); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(order, models.StatusCreated, models.StatusPaid, "支付成功"); err != nil {
		if strings.Contains(err.Error(), "状态不一致") {
			return ErrConflict
		}
		return err
	}

	if err := s.repo.Commit(); err != nil {
		return err
	}
	committed = true

	go s.notifyWarehousePick(order)

	return nil
}

func (s *OrderService) Ship(orderID string) (string, error) {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		return "", err
	}
	if order == nil {
		return "", ErrOrderNotFound
	}

	if s.stateMachine.IsTerminal(order.Status) {
		return "", ErrTerminalState
	}

	if err := s.stateMachine.ValidateTransition(order.Status, models.StatusShipped); err != nil {
		return "", ErrInvalidTransition
	}

	trackingNumber := generateTrackingNumber()

	if err := s.repo.BeginTx(); err != nil {
		return "", err
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.Rollback()
		}
	}()

	if err := s.repo.UpdateTrackingNumber(order, trackingNumber); err != nil {
		return "", err
	}

	if err := s.repo.UpdateStatus(order, models.StatusPaid, models.StatusShipped, "已发货"); err != nil {
		if strings.Contains(err.Error(), "状态不一致") {
			return "", ErrConflict
		}
		return "", err
	}

	if err := s.repo.Commit(); err != nil {
		return "", err
	}
	committed = true

	go s.notifyLogisticsTracking(order)

	return trackingNumber, nil
}

func (s *OrderService) Deliver(orderID string) error {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}

	if s.stateMachine.IsTerminal(order.Status) {
		return ErrTerminalState
	}

	if err := s.stateMachine.ValidateTransition(order.Status, models.StatusDelivered); err != nil {
		return ErrInvalidTransition
	}

	if err := s.repo.BeginTx(); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.Rollback()
		}
	}()

	if err := s.repo.UpdateDeliveredAt(order); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(order, models.StatusShipped, models.StatusDelivered, "已签收"); err != nil {
		if strings.Contains(err.Error(), "状态不一致") {
			return ErrConflict
		}
		return err
	}

	if err := s.repo.Commit(); err != nil {
		return err
	}
	committed = true

	return nil
}

func (s *OrderService) Complete(orderID string) error {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}

	if s.stateMachine.IsTerminal(order.Status) {
		return ErrTerminalState
	}

	if err := s.stateMachine.ValidateTransition(order.Status, models.StatusCompleted); err != nil {
		return ErrInvalidTransition
	}

	if err := s.repo.UpdateStatus(order, models.StatusDelivered, models.StatusCompleted, "订单完成"); err != nil {
		if strings.Contains(err.Error(), "状态不一致") {
			return ErrConflict
		}
		return err
	}

	return nil
}

func (s *OrderService) RequestRefund(orderID string) error {
	order, err := s.repo.GetByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}

	if order.Status != models.StatusDelivered {
		if s.stateMachine.IsTerminal(order.Status) {
			return ErrTerminalState
		}
		return ErrInvalidTransition
	}

	if order.DeliveredAt != nil {
		if time.Since(*order.DeliveredAt) > 7*24*time.Hour {
			return ErrRefundWindowExpired
		}
	}

	if err := s.repo.BeginTx(); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.Rollback()
		}
	}()

	if err := s.repo.UpdateStatus(order, models.StatusDelivered, models.StatusRefunding, "申请退款"); err != nil {
		if strings.Contains(err.Error(), "状态不一致") {
			return ErrConflict
		}
		return err
	}

	if err := s.inventorySvc.Restore(order.ProductID, order.Quantity); err != nil {
		return err
	}

	if order.PaymentMethod != nil {
		gateway, gerr := s.paymentFactory.GetGateway(*order.PaymentMethod)
		if gerr == nil {
			gerr = gateway.RefundPayment(orderID, order.TotalAmount)
			if gerr != nil {
				return gerr
			}
		}
	}

	if err := s.repo.UpdateStatus(order, models.StatusRefunding, models.StatusRefunded, "退款完成"); err != nil {
		if strings.Contains(err.Error(), "状态不一致") {
			return ErrConflict
		}
		return err
	}

	if err := s.repo.Commit(); err != nil {
		return err
	}
	committed = true

	return nil
}

func (s *OrderService) CancelExpiredOrders() error {
	orders, err := s.repo.GetExpiredOrders()
	if err != nil {
		return err
	}

	for _, order := range orders {
		if err = s.cancelOrder(order); err != nil {
			fmt.Printf("自动取消订单 %s 失败: %v\n", order.ID, err)
		}
	}
	return nil
}

func (s *OrderService) cancelOrder(order *models.Order) error {
	if err := s.repo.BeginTx(); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			s.repo.Rollback()
		}
	}()

	if err := s.inventorySvc.Restore(order.ProductID, order.Quantity); err != nil {
		return err
	}

	if err := s.repo.UpdateStatus(order, models.StatusCreated, models.StatusCancelled, "超时未支付自动取消"); err != nil {
		if strings.Contains(err.Error(), "状态不一致") {
			s.repo.Rollback()
			return nil
		}
		return err
	}

	if err := s.repo.Commit(); err != nil {
		return err
	}
	committed = true

	return nil
}

func (s *OrderService) AddInventory(productID string, quantity int) error {
	return s.inventorySvc.AddInventory(productID, quantity)
}

func generateTrackingNumber() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 2)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("TRK%s%s%08d", string(b), time.Now().Format("20060102"), rand.Intn(100000000))
}

func (s *OrderService) notifyInventoryDeducted(order *models.Order) {
	fmt.Printf("[事件通知] 库存扣减: 订单 %s, 商品 %s, 数量 %d\n", order.ID, order.ProductID, order.Quantity)
}

func (s *OrderService) notifyWarehousePick(order *models.Order) {
	fmt.Printf("[事件通知] 通知仓库拣货: 订单 %s\n", order.ID)
}

func (s *OrderService) notifyLogisticsTracking(order *models.Order) {
	if order.TrackingNumber != nil {
		fmt.Printf("[事件通知] 通知物流跟踪: 订单 %s, 物流单号 %s\n", order.ID, *order.TrackingNumber)
	}
}
