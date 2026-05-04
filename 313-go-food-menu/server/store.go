package server

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/example/food-menu/common"
)

var (
	ErrDishNotFound        = errors.New("菜品不存在")
	ErrDishNameEmpty       = errors.New("菜品名称不能为空")
	ErrDishNameTooLong     = errors.New("菜品名称不能超过30个字符")
	ErrDishPriceInvalid    = errors.New("菜品价格必须大于0")
	ErrDishDescTooLong     = errors.New("菜品描述不能超过200个字符")
	ErrCategoryInvalid     = errors.New("分类名称无效")
	ErrDishNameDuplicate   = errors.New("同一分类下菜品名称不能重复")
	ErrDishNotInCategory   = errors.New("菜品不在该分类中")
)

type Store struct {
	mu           sync.RWMutex
	dishes       map[string]*common.Dish
	nameIndex    map[string]map[string]bool
	recommendInfo *common.RecommendInfo
}

func NewStore() *Store {
	return &Store{
		dishes:    make(map[string]*common.Dish),
		nameIndex: make(map[string]map[string]bool),
		recommendInfo: &common.RecommendInfo{
			Date:    time.Now().Format("2006-01-02"),
			DishIDs: []string{},
		},
	}
}

func (s *Store) validateCategory(category string) bool {
	for _, c := range common.ValidCategories {
		if c == category {
			return true
		}
	}
	return false
}

func (s *Store) validateCreateDish(req *common.CreateDishRequest) error {
	if req.Name == "" {
		return ErrDishNameEmpty
	}
	if len([]rune(req.Name)) > common.MaxNameLength {
		return ErrDishNameTooLong
	}
	if req.Price <= 0 {
		return ErrDishPriceInvalid
	}
	if len([]rune(req.Description)) > common.MaxDescriptionLength {
		return ErrDishDescTooLong
	}
	if !s.validateCategory(req.Category) {
		return ErrCategoryInvalid
	}
	return nil
}

func (s *Store) isNameDuplicateInCategory(name, category string, excludeID string) bool {
	if names, ok := s.nameIndex[category]; ok {
		for id, exists := range names {
			if id != excludeID && exists {
				dish := s.dishes[id]
				if dish != nil && strings.EqualFold(dish.Name, name) {
					return true
				}
			}
		}
	}
	return false
}

func (s *Store) CreateDish(req *common.CreateDishRequest) (*common.Dish, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.validateCreateDish(req); err != nil {
		return nil, err
	}

	if s.isNameDuplicateInCategory(req.Name, req.Category, "") {
		return nil, ErrDishNameDuplicate
	}

	now := time.Now()
	dish := &common.Dish{
		ID:          generateID(),
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		Category:    req.Category,
		Status:      common.StatusOnSale,
		IsRecommend: false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.dishes[dish.ID] = dish
	if s.nameIndex[dish.Category] == nil {
		s.nameIndex[dish.Category] = make(map[string]bool)
	}
	s.nameIndex[dish.Category][dish.ID] = true

	return dish, nil
}

func (s *Store) UpdateDish(id string, req *common.UpdateDishRequest) (*common.Dish, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dish, ok := s.dishes[id]
	if !ok {
		return nil, ErrDishNotFound
	}

	oldCategory := dish.Category
	newCategory := oldCategory
	if req.Category != "" {
		if !s.validateCategory(req.Category) {
			return nil, ErrCategoryInvalid
		}
		newCategory = req.Category
	}

	if req.Name != "" {
		if len([]rune(req.Name)) > common.MaxNameLength {
			return nil, ErrDishNameTooLong
		}
		checkCategory := newCategory
		if s.isNameDuplicateInCategory(req.Name, checkCategory, id) {
			return nil, ErrDishNameDuplicate
		}
		dish.Name = req.Name
	}

	if req.Price > 0 {
		dish.Price = req.Price
	}

	if req.Description != "" {
		if len([]rune(req.Description)) > common.MaxDescriptionLength {
			return nil, ErrDishDescTooLong
		}
		dish.Description = req.Description
	}

	if req.ImageURL != "" {
		dish.ImageURL = req.ImageURL
	}

	if newCategory != oldCategory {
		if s.isNameDuplicateInCategory(dish.Name, newCategory, id) {
			return nil, ErrDishNameDuplicate
		}

		if s.nameIndex[oldCategory] != nil {
			delete(s.nameIndex[oldCategory], id)
		}

		if s.nameIndex[newCategory] == nil {
			s.nameIndex[newCategory] = make(map[string]bool)
		}
		s.nameIndex[newCategory][id] = true

		dish.Category = newCategory
	}

	dish.UpdatedAt = time.Now()
	return dish, nil
}

func (s *Store) DeleteDish(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dish, ok := s.dishes[id]
	if !ok {
		return ErrDishNotFound
	}

	dish.Status = common.StatusOffSale
	dish.UpdatedAt = time.Now()
	return nil
}

func (s *Store) SetDishStatus(id string, status common.DishStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dish, ok := s.dishes[id]
	if !ok {
		return ErrDishNotFound
	}

	dish.Status = status
	dish.UpdatedAt = time.Now()
	return nil
}

func (s *Store) GetDish(id string) (*common.Dish, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dish, ok := s.dishes[id]
	if !ok {
		return nil, ErrDishNotFound
	}
	return dish, nil
}

func (s *Store) ListDishesByCategory(category string, onlyOnSale bool) ([]*common.Dish, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if category != "" && !s.validateCategory(category) {
		return nil, ErrCategoryInvalid
	}

	var dishes []*common.Dish
	for _, dish := range s.dishes {
		if category != "" && dish.Category != category {
			continue
		}
		if onlyOnSale && dish.Status != common.StatusOnSale {
			continue
		}
		dishes = append(dishes, dish)
	}
	return dishes, nil
}

func (s *Store) SearchDishes(category, keyword string) ([]*common.Dish, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if category != "" && !s.validateCategory(category) {
		return nil, ErrCategoryInvalid
	}

	keywordLower := strings.ToLower(keyword)
	var dishes []*common.Dish

	for _, dish := range s.dishes {
		if category != "" && dish.Category != category {
			continue
		}
		if dish.Status != common.StatusOnSale {
			continue
		}
		if keyword == "" {
			dishes = append(dishes, dish)
			continue
		}
		if strings.Contains(strings.ToLower(dish.Name), keywordLower) {
			dishes = append(dishes, dish)
		}
	}
	return dishes, nil
}

func (s *Store) GetCategorySummary() []*common.CategorySummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summaries := make([]*common.CategorySummary, 0, len(common.ValidCategories))
	for _, category := range common.ValidCategories {
		summary := &common.CategorySummary{
			Category: category,
		}
		for _, dish := range s.dishes {
			if dish.Category == category {
				summary.TotalCount++
				if dish.Status == common.StatusOnSale {
					summary.OnSaleCount++
				}
			}
		}
		summaries = append(summaries, summary)
	}
	return summaries
}

func (s *Store) SetTodayRecommend(dishIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	today := time.Now().Format("2006-01-02")

	for id := range s.dishes {
		s.dishes[id].IsRecommend = false
	}

	for _, id := range dishIDs {
		dish, ok := s.dishes[id]
		if !ok {
			return fmt.Errorf("%w: %s", ErrDishNotFound, id)
		}
		if dish.Status != common.StatusOnSale {
			return fmt.Errorf("菜品 %s 已下架，不能设置为推荐", dish.Name)
		}
	}

	for _, id := range dishIDs {
		s.dishes[id].IsRecommend = true
		s.dishes[id].UpdatedAt = time.Now()
	}

	s.recommendInfo = &common.RecommendInfo{
		Date:      today,
		DishIDs:   dishIDs,
		UpdatedAt: time.Now(),
	}

	return nil
}

func (s *Store) GetTodayRecommend() (*common.RecommendInfo, []*common.Dish) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	today := time.Now().Format("2006-01-02")

	if s.recommendInfo.Date != today {
		s.mu.RUnlock()
		s.mu.Lock()
		s.recommendInfo = &common.RecommendInfo{
			Date:    today,
			DishIDs: []string{},
		}
		for id := range s.dishes {
			s.dishes[id].IsRecommend = false
		}
		s.mu.Unlock()
		s.mu.RLock()
	}

	var dishes []*common.Dish
	for _, id := range s.recommendInfo.DishIDs {
		if dish, ok := s.dishes[id]; ok && dish.Status == common.StatusOnSale {
			dishes = append(dishes, dish)
		}
	}

	return s.recommendInfo, dishes
}

func (s *Store) GetAllDishes() map[string]*common.Dish {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*common.Dish)
	for id, dish := range s.dishes {
		result[id] = dish
	}
	return result
}

func (s *Store) LoadDishes(dishes map[string]*common.Dish) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.dishes = dishes
	s.nameIndex = make(map[string]map[string]bool)

	for id, dish := range dishes {
		if s.nameIndex[dish.Category] == nil {
			s.nameIndex[dish.Category] = make(map[string]bool)
		}
		s.nameIndex[dish.Category][id] = true
	}
}

func (s *Store) GetRecommendInfo() *common.RecommendInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.recommendInfo
}

func (s *Store) SetRecommendInfo(info *common.RecommendInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recommendInfo = info
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
