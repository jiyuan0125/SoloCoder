package core

import (
	"errors"
	"marketplace/common"
	"time"
)

type OrderService struct {
	store   *Store
	itemSvc *ItemService
}

func NewOrderService(store *Store, itemSvc *ItemService) *OrderService {
	return &OrderService{
		store:   store,
		itemSvc: itemSvc,
	}
}

func calculateServiceFee(price int64) int64 {
	if price < common.MinFeeThreshold {
		return 0
	}
	feeInYuan := float64(price) / 100.0 * common.ServiceFeeRate
	feeInJiao := feeInYuan * 10
	roundedJiao := int64(feeInJiao + 0.5)
	return roundedJiao * 10
}

func (s *OrderService) CreateFromNegotiation(n *common.Negotiation) (*common.Order, error) {
	if n.Status != common.NegotiationStatusAccepted {
		return nil, errors.New("议价未被接受，无法创建订单")
	}

	item, err := s.itemSvc.Get(n.ItemID)
	if err != nil {
		return nil, err
	}

	if item.Status == common.ItemStatusSold {
		return nil, errors.New("商品已售出")
	}

	now := time.Now()
	order := &common.Order{
		ID:              generateID(),
		ItemID:          n.ItemID,
		BuyerID:         n.BuyerID,
		SellerID:        n.SellerID,
		NegotiationID:   n.ID,
		Price:           n.FinalPrice,
		ServiceFee:      calculateServiceFee(n.FinalPrice),
		Status:          common.OrderStatusPendingPayment,
		CreatedAt:       now,
		StatusUpdatedAt: now,
	}

	item.Status = common.ItemStatusSold
	s.store.SaveItem(item)
	s.store.SaveOrder(order)

	return order, nil
}

func (s *OrderService) Pay(buyerID, orderID string) (*common.Order, error) {
	order, ok := s.store.GetOrder(orderID)
	if !ok {
		return nil, errors.New("订单不存在")
	}

	if order.BuyerID != buyerID {
		return nil, errors.New("无权限操作此订单")
	}

	if order.Status != common.OrderStatusPendingPayment {
		return nil, errors.New("订单不在待付款状态")
	}

	s.checkTimeout(order)
	if order.Status == common.OrderStatusCancelled {
		return order, errors.New("订单已超时取消")
	}

	now := time.Now()
	order.Status = common.OrderStatusPaid
	order.PaidAt = &now
	order.StatusUpdatedAt = now

	s.store.SaveOrder(order)
	return order, nil
}

func (s *OrderService) Ship(sellerID, orderID, logisticsNo string) (*common.Order, error) {
	if logisticsNo == "" {
		return nil, errors.New("物流单号不能为空")
	}

	order, ok := s.store.GetOrder(orderID)
	if !ok {
		return nil, errors.New("订单不存在")
	}

	if order.SellerID != sellerID {
		return nil, errors.New("无权限操作此订单")
	}

	if order.Status != common.OrderStatusPaid {
		return nil, errors.New("订单不在待发货状态")
	}

	now := time.Now()
	order.Status = common.OrderStatusShipped
	order.LogisticsNo = logisticsNo
	order.ShippedAt = &now
	order.StatusUpdatedAt = now

	s.store.SaveOrder(order)
	return order, nil
}

func (s *OrderService) ConfirmReceive(buyerID, orderID string) (*common.Order, error) {
	order, ok := s.store.GetOrder(orderID)
	if !ok {
		return nil, errors.New("订单不存在")
	}

	if order.BuyerID != buyerID {
		return nil, errors.New("无权限操作此订单")
	}

	if order.Status != common.OrderStatusShipped {
		return nil, errors.New("订单不在待收货状态")
	}

	now := time.Now()
	order.Status = common.OrderStatusCompleted
	order.CompletedAt = &now
	order.StatusUpdatedAt = now

	s.store.SaveOrder(order)
	return order, nil
}

func (s *OrderService) checkTimeout(order *common.Order) {
	if order.Status == common.OrderStatusPendingPayment {
		if time.Since(order.StatusUpdatedAt) > common.PaymentTimeout {
			order.Status = common.OrderStatusCancelled
			order.StatusUpdatedAt = time.Now()

			item, _ := s.store.GetItem(order.ItemID)
			if item != nil {
				item.Status = common.ItemStatusOnSale
				s.store.SaveItem(item)
			}

			s.store.SaveOrder(order)
		}
	} else if order.Status == common.OrderStatusShipped {
		if time.Since(order.StatusUpdatedAt) > common.ConfirmReceiptTimeout {
			now := time.Now()
			order.Status = common.OrderStatusCompleted
			order.CompletedAt = &now
			order.StatusUpdatedAt = now
			s.store.SaveOrder(order)
		}
	}
}

func (s *OrderService) ProcessTimeouts() {
	orders := s.store.ListOrders()
	for _, order := range orders {
		if order.Status == common.OrderStatusPendingPayment || order.Status == common.OrderStatusShipped {
			s.checkTimeout(order)
		}
	}
}

func (s *OrderService) Get(id string) (*common.Order, error) {
	order, ok := s.store.GetOrder(id)
	if !ok {
		return nil, errors.New("订单不存在")
	}
	return order, nil
}
