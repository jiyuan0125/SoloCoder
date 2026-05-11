package core

import (
	"errors"
	"fmt"
	"marketplace/common"
	"sync"
	"time"
)

type NegotiationService struct {
	store     *Store
	itemSvc   *ItemService
	orderSvc  *OrderService
	itemLocks map[string]*sync.Mutex
	locksMu   sync.Mutex
}

func NewNegotiationService(store *Store, itemSvc *ItemService) *NegotiationService {
	return &NegotiationService{
		store:     store,
		itemSvc:   itemSvc,
		itemLocks: make(map[string]*sync.Mutex),
	}
}

func (s *NegotiationService) SetOrderService(os *OrderService) {
	s.orderSvc = os
}

func (s *NegotiationService) getNegotiationLock(nID string) *sync.Mutex {
	s.locksMu.Lock()
	defer s.locksMu.Unlock()
	if lock, ok := s.itemLocks[nID]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	s.itemLocks[nID] = lock
	return lock
}

func (s *NegotiationService) MakeOffer(buyerID, itemID string, price int64) (*common.Negotiation, error) {
	if buyerID == "" {
		return nil, errors.New("买家ID不能为空")
	}
	if itemID == "" {
		return nil, errors.New("商品ID不能为空")
	}
	if price <= 0 {
		return nil, errors.New("出价必须大于0")
	}

	lock := s.itemSvc.getItemLock(itemID)
	lock.Lock()
	defer lock.Unlock()

	item, err := s.itemSvc.Get(itemID)
	if err != nil {
		return nil, err
	}
	if item.Status != common.ItemStatusOnSale {
		return nil, errors.New("商品不在在售状态")
	}
	if item.SellerID == buyerID {
		return nil, errors.New("不能对自己的商品出价")
	}

	minPrice := item.Price / 2
	if price < minPrice {
		return nil, errors.New(fmt.Sprintf("出价不能低于卖家标价的50%%（%d分）", minPrice))
	}

	now := time.Now()
	n := &common.Negotiation{
		ID:           generateID(),
		ItemID:       itemID,
		BuyerID:      buyerID,
		SellerID:     item.SellerID,
		Status:       common.NegotiationStatusActive,
		Rounds: []common.NegotiationRound{
			{
				Round: 1,
				Role:  common.RoleBuyer,
				Price: price,
				Time:  now,
			},
		},
		CreatedAt:    now,
		UpdatedAt:    now,
		LastActiveAt: now,
	}

	s.store.SaveNegotiation(n)
	item.Status = common.ItemStatusNegotiating
	s.store.SaveItem(item)

	return n, nil
}

func (s *NegotiationService) checkTimeout(n *common.Negotiation) bool {
	if n.Status != common.NegotiationStatusActive {
		return false
	}
	if time.Since(n.LastActiveAt) > common.NegotiationTimeout {
		n.Status = common.NegotiationStatusTimeout
		s.store.SaveNegotiation(n)

		item, _ := s.store.GetItem(n.ItemID)
		if item != nil && !s.store.HasActiveNegotiationForItem(n.ItemID) {
			item.Status = common.ItemStatusOnSale
			s.store.SaveItem(item)
		}
		return true
	}
	return false
}

func (s *NegotiationService) SellerRespond(sellerID, negotiationID string, action string, counterPrice int64) (*common.Negotiation, error) {
	if action != "accept" && action != "reject" && action != "counter" {
		return nil, errors.New("无效的操作类型，只能是 accept、reject 或 counter")
	}

	n, ok := s.store.GetNegotiation(negotiationID)
	if !ok {
		return nil, errors.New("议价不存在")
	}

	if n.SellerID != sellerID {
		return nil, errors.New("无权限操作此议价")
	}

	if n.Status != common.NegotiationStatusActive {
		return nil, errors.New("议价已结束")
	}

	if s.checkTimeout(n) {
		return n, errors.New("议价已超时")
	}

	lastRound := n.Rounds[len(n.Rounds)-1]
	if lastRound.Role != common.RoleBuyer {
		return nil, errors.New("当前不是卖家回合")
	}

	if len(n.Rounds) > common.MaxNegotiationRounds*2-1 {
		n.Status = common.NegotiationStatusClosed
		s.store.SaveNegotiation(n)
		return n, errors.New("议价轮次已达上限")
	}

	now := time.Now()

	switch action {
	case "accept":
		n.Status = common.NegotiationStatusAccepted
		n.FinalPrice = lastRound.Price
		n.LastActiveAt = now
		s.store.SaveNegotiation(n)

		if s.orderSvc != nil {
			_, err := s.orderSvc.CreateFromNegotiation(n)
			if err != nil {
				return n, err
			}
		}

	case "reject":
		n.Status = common.NegotiationStatusRejected
		n.LastActiveAt = now
		s.store.SaveNegotiation(n)

		item, _ := s.store.GetItem(n.ItemID)
		if item != nil && !s.store.HasActiveNegotiationForItem(n.ItemID) {
			item.Status = common.ItemStatusOnSale
			s.store.SaveItem(item)
		}

	case "counter":
		if counterPrice <= 0 {
			return nil, errors.New("还价金额必须大于0")
		}
		if counterPrice >= lastRound.Price {
			return nil, errors.New("还价必须低于上一轮价格")
		}

		item, err := s.itemSvc.Get(n.ItemID)
		if err != nil {
			return nil, err
		}
		minPrice := item.Price / 2
		if counterPrice < minPrice {
			return nil, errors.New(fmt.Sprintf("还价不能低于卖家标价的50%%（%d分）", minPrice))
		}

		n.Rounds = append(n.Rounds, common.NegotiationRound{
			Round: len(n.Rounds) + 1,
			Role:  common.RoleSeller,
			Price: counterPrice,
			Time:  now,
		})
		n.LastActiveAt = now
		s.store.SaveNegotiation(n)
	}

	return n, nil
}

func (s *NegotiationService) BuyerRespond(buyerID, negotiationID string, action string, counterPrice int64) (*common.Negotiation, error) {
	if action != "accept" && action != "reject" && action != "counter" {
		return nil, errors.New("无效的操作类型，只能是 accept、reject 或 counter")
	}

	n, ok := s.store.GetNegotiation(negotiationID)
	if !ok {
		return nil, errors.New("议价不存在")
	}

	if n.BuyerID != buyerID {
		return nil, errors.New("无权限操作此议价")
	}

	if n.Status != common.NegotiationStatusActive {
		return nil, errors.New("议价已结束")
	}

	if s.checkTimeout(n) {
		return n, errors.New("议价已超时")
	}

	lastRound := n.Rounds[len(n.Rounds)-1]
	if lastRound.Role != common.RoleSeller {
		return nil, errors.New("当前不是买家回合")
	}

	if len(n.Rounds) > common.MaxNegotiationRounds*2-1 {
		n.Status = common.NegotiationStatusClosed
		s.store.SaveNegotiation(n)
		return n, errors.New("议价轮次已达上限")
	}

	now := time.Now()

	switch action {
	case "accept":
		n.Status = common.NegotiationStatusAccepted
		n.FinalPrice = lastRound.Price
		n.LastActiveAt = now
		s.store.SaveNegotiation(n)

		if s.orderSvc != nil {
			_, err := s.orderSvc.CreateFromNegotiation(n)
			if err != nil {
				return n, err
			}
		}

	case "reject":
		n.Status = common.NegotiationStatusRejected
		n.LastActiveAt = now
		s.store.SaveNegotiation(n)

		item, _ := s.store.GetItem(n.ItemID)
		if item != nil && !s.store.HasActiveNegotiationForItem(n.ItemID) {
			item.Status = common.ItemStatusOnSale
			s.store.SaveItem(item)
		}

	case "counter":
		if counterPrice <= 0 {
			return nil, errors.New("还价金额必须大于0")
		}
		if counterPrice >= lastRound.Price {
			return nil, errors.New("还价必须低于上一轮价格")
		}

		item, err := s.itemSvc.Get(n.ItemID)
		if err != nil {
			return nil, err
		}
		minPrice := item.Price / 2
		if counterPrice < minPrice {
			return nil, errors.New(fmt.Sprintf("出价不能低于卖家标价的50%%（%d分）", minPrice))
		}

		n.Rounds = append(n.Rounds, common.NegotiationRound{
			Round: len(n.Rounds) + 1,
			Role:  common.RoleBuyer,
			Price: counterPrice,
			Time:  now,
		})
		n.LastActiveAt = now
		s.store.SaveNegotiation(n)
	}

	return n, nil
}

func (s *NegotiationService) Get(id string) (*common.Negotiation, error) {
	n, ok := s.store.GetNegotiation(id)
	if !ok {
		return nil, errors.New("议价不存在")
	}
	return n, nil
}
