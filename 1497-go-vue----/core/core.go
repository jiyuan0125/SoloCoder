package core

import (
	"marketplace/common"
)

type Service struct {
	Store       *Store
	User        *UserService
	Item        *ItemService
	Negotiation *NegotiationService
	Order       *OrderService
}

func NewService() *Service {
	store := NewStore()
	userSvc := NewUserService(store)
	itemSvc := NewItemService(store, userSvc)
	orderSvc := NewOrderService(store, itemSvc)
	negotiationSvc := NewNegotiationService(store, itemSvc)
	negotiationSvc.SetOrderService(orderSvc)

	return &Service{
		Store:       store,
		User:        userSvc,
		Item:        itemSvc,
		Negotiation: negotiationSvc,
		Order:       orderSvc,
	}
}

func (s *Service) RegisterUser(username string, role common.UserRole) (*common.User, error) {
	return s.User.Register(username, role)
}

func (s *Service) GetUser(id string) (*common.User, error) {
	return s.User.Get(id)
}

func (s *Service) CreateItem(sellerID string, req common.CreateItemRequest) (*common.Item, error) {
	return s.Item.Create(sellerID, req)
}

func (s *Service) ApproveItem(adminID, itemID string, approved bool, reason string) (*common.Item, error) {
	return s.Item.Approve(adminID, itemID, approved, reason)
}

func (s *Service) SearchItems(req common.SearchItemsRequest) []*common.Item {
	return s.Item.Search(req.Keyword, req.Category, req.MinPrice, req.MaxPrice, req.Condition)
}

func (s *Service) GetItem(id string) (*common.Item, error) {
	return s.Item.Get(id)
}

func (s *Service) RemoveItem(sellerID, itemID string) error {
	return s.Item.Remove(sellerID, itemID)
}

func (s *Service) MakeOffer(buyerID, itemID string, price int64) (*common.Negotiation, error) {
	return s.Negotiation.MakeOffer(buyerID, itemID, price)
}

func (s *Service) SellerRespond(sellerID, negotiationID string, action string, counterPrice int64) (*common.Negotiation, error) {
	return s.Negotiation.SellerRespond(sellerID, negotiationID, action, counterPrice)
}

func (s *Service) BuyerRespond(buyerID, negotiationID string, action string, counterPrice int64) (*common.Negotiation, error) {
	return s.Negotiation.BuyerRespond(buyerID, negotiationID, action, counterPrice)
}

func (s *Service) GetNegotiation(id string) (*common.Negotiation, error) {
	return s.Negotiation.Get(id)
}

func (s *Service) PayOrder(buyerID, orderID string) (*common.Order, error) {
	return s.Order.Pay(buyerID, orderID)
}

func (s *Service) ShipOrder(sellerID, orderID, logisticsNo string) (*common.Order, error) {
	return s.Order.Ship(sellerID, orderID, logisticsNo)
}

func (s *Service) ConfirmReceiveOrder(buyerID, orderID string) (*common.Order, error) {
	return s.Order.ConfirmReceive(buyerID, orderID)
}

func (s *Service) GetOrder(id string) (*common.Order, error) {
	return s.Order.Get(id)
}

func (s *Service) ProcessTimeouts() {
	s.Order.ProcessTimeouts()
}
