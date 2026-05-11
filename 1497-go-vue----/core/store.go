package core

import (
	"marketplace/common"
	"sync"
	"time"
)

type Store struct {
	users        map[string]*common.User
	items        map[string]*common.Item
	negotiations map[string]*common.Negotiation
	orders       map[string]*common.Order
	mu           sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		users:        make(map[string]*common.User),
		items:        make(map[string]*common.Item),
		negotiations: make(map[string]*common.Negotiation),
		orders:       make(map[string]*common.Order),
	}
}

func (s *Store) SaveUser(user *common.User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[user.ID] = user
}

func (s *Store) GetUser(id string) (*common.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

func (s *Store) SaveItem(item *common.Item) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item.UpdatedAt = time.Now()
	s.items[item.ID] = item
}

func (s *Store) GetItem(id string) (*common.Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	return item, ok
}

func (s *Store) ListItems() []*common.Item {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]*common.Item, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item)
	}
	return items
}

func (s *Store) CountActiveItemsBySeller(sellerID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, item := range s.items {
		if item.SellerID == sellerID && 
		   (item.Status == common.ItemStatusOnSale || 
		    item.Status == common.ItemStatusNegotiating || 
		    item.Status == common.ItemStatusPending) {
			count++
		}
	}
	return count
}

func (s *Store) SaveNegotiation(n *common.Negotiation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n.UpdatedAt = time.Now()
	s.negotiations[n.ID] = n
}

func (s *Store) GetNegotiation(id string) (*common.Negotiation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.negotiations[id]
	return n, ok
}

func (s *Store) ListNegotiationsByItem(itemID string) []*common.Negotiation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*common.Negotiation, 0)
	for _, n := range s.negotiations {
		if n.ItemID == itemID {
			result = append(result, n)
		}
	}
	return result
}

func (s *Store) HasActiveNegotiationForItem(itemID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, n := range s.negotiations {
		if n.ItemID == itemID && n.Status == common.NegotiationStatusActive {
			return true
		}
	}
	return false
}

func (s *Store) SaveOrder(o *common.Order) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orders[o.ID] = o
}

func (s *Store) GetOrder(id string) (*common.Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	return o, ok
}

func (s *Store) ListOrders() []*common.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*common.Order, 0, len(s.orders))
	for _, o := range s.orders {
		result = append(result, o)
	}
	return result
}
