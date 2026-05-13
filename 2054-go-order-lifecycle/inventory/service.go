package inventory

import (
	"fmt"

	"order-lifecycle/repository"
)

type InventoryService struct {
	store *repository.Store
}

func NewInventoryService(store *repository.Store) *InventoryService {
	return &InventoryService{store: store}
}

func (s *InventoryService) GetAvailable(productID string) (int, error) {
	return s.store.GetInventory(productID), nil
}

func (s *InventoryService) Deduct(productID string, quantity int) error {
	current := s.store.GetInventory(productID)
	if current < quantity {
		return fmt.Errorf("库存不足")
	}
	s.store.SetInventory(productID, current-quantity)
	return nil
}

func (s *InventoryService) Restore(productID string, quantity int) error {
	current := s.store.GetInventory(productID)
	s.store.SetInventory(productID, current+quantity)
	return nil
}

func (s *InventoryService) AddInventory(productID string, quantity int) error {
	s.store.AddInventory(productID, quantity)
	return nil
}
