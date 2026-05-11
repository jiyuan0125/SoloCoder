package core

import (
	"sync"
	"recycling/api"
)

type CategoryStore struct {
	mu        sync.RWMutex
	categories map[string]*api.Category
}

func NewCategoryStore() *CategoryStore {
	cs := &CategoryStore{
		categories: make(map[string]*api.Category),
	}
	cs.initDefaultCategories()
	return cs
}

func (cs *CategoryStore) initDefaultCategories() {
	presets := []struct {
		id        string
		name      string
		parentID  string
		price     float64
	}{
		{"paper", "废纸类", "", 0},
		{"paper-newspaper", "报纸", "paper", 1.50},
		{"paper-magazine", "杂志", "paper", 1.20},
		{"paper-box", "纸箱", "paper", 1.20},
		{"paper-office", "办公用纸", "paper", 1.80},
		{"plastic", "塑料类", "", 0},
		{"plastic-pet", "PET瓶", "plastic", 2.50},
		{"plastic-hdpe", "HDPE容器", "plastic", 1.80},
		{"plastic-film", "塑料薄膜", "plastic", 1.00},
		{"metal", "金属类", "", 0},
		{"metal-iron", "废铁", "metal", 3.00},
		{"metal-copper", "废铜", "metal", 25.00},
		{"metal-aluminum", "废铝", "metal", 12.00},
		{"metal-stainless", "不锈钢", "metal", 8.00},
		{"glass", "玻璃类", "", 0},
		{"glass-clear", "透明玻璃", "glass", 0.80},
		{"glass-colored", "有色玻璃", "glass", 0.50},
		{"electronic", "电子废弃物", "", 0},
		{"electronic-phone", "旧手机", "electronic", 50.00},
		{"electronic-computer", "旧电脑", "electronic", 80.00},
		{"electronic-battery", "电池", "electronic", 2.00},
	}

	for _, p := range presets {
		cs.categories[p.id] = &api.Category{
			ID:       p.id,
			Name:     p.name,
			ParentID: p.parentID,
			Price:    p.price,
		}
	}
}

func (cs *CategoryStore) GetAll() []api.Category {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	result := make([]api.Category, 0, len(cs.categories))
	for _, c := range cs.categories {
		result = append(result, *c)
	}
	return result
}

func (cs *CategoryStore) GetByID(id string) (*api.Category, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	c, exists := cs.categories[id]
	if !exists {
		return nil, false
	}
	return &api.Category{
		ID:       c.ID,
		Name:     c.Name,
		ParentID: c.ParentID,
		Price:    c.Price,
	}, true
}

func (cs *CategoryStore) UpdatePrice(id string, newPrice float64) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	c, exists := cs.categories[id]
	if !exists {
		return false
	}
	c.Price = newPrice
	return true
}
