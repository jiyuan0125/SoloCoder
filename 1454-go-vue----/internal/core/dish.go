package core

import (
	"canteen/internal/common"
	"time"

	"github.com/google/uuid"
)

func (s *CanteenService) CreateDish(name string, price int, nutrition common.NutritionInfo) *common.Dish {
	s.dishesMu.Lock()
	defer s.dishesMu.Unlock()

	now := time.Now()
	dish := &common.Dish{
		ID:        uuid.New().String(),
		Name:      name,
		Price:     price,
		Nutrition: nutrition,
		IsOnSale:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.dishes[dish.ID] = dish
	return dish
}

func (s *CanteenService) GetDish(id string) (*common.Dish, error) {
	s.dishesMu.RLock()
	defer s.dishesMu.RUnlock()

	dish, exists := s.dishes[id]
	if !exists {
		return nil, ErrDishNotFound
	}
	return dish, nil
}

func (s *CanteenService) ListDishes(onlyOnSale bool) []*common.Dish {
	s.dishesMu.RLock()
	defer s.dishesMu.RUnlock()

	dishes := make([]*common.Dish, 0)
	for _, d := range s.dishes {
		if onlyOnSale && !d.IsOnSale {
			continue
		}
		dishes = append(dishes, d)
	}
	return dishes
}

func (s *CanteenService) UpdateDish(id string, name *string, price *int, nutrition *common.NutritionInfo) (*common.Dish, error) {
	s.dishesMu.Lock()
	defer s.dishesMu.Unlock()

	dish, exists := s.dishes[id]
	if !exists {
		return nil, ErrDishNotFound
	}

	if name != nil {
		dish.Name = *name
	}
	if price != nil {
		dish.Price = *price
	}
	if nutrition != nil {
		dish.Nutrition = *nutrition
	}

	dish.UpdatedAt = time.Now()
	return dish, nil
}

func (s *CanteenService) SetDishOnSale(id string, onSale bool) error {
	s.dishesMu.Lock()
	defer s.dishesMu.Unlock()

	dish, exists := s.dishes[id]
	if !exists {
		return ErrDishNotFound
	}

	dish.IsOnSale = onSale
	dish.UpdatedAt = time.Now()
	return nil
}

func (s *CanteenService) snapshotDish(dish *common.Dish) common.DishSnapshot {
	return common.DishSnapshot{
		ID:        dish.ID,
		Name:      dish.Name,
		Price:     dish.Price,
		Nutrition: dish.Nutrition,
	}
}
