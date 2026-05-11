package core

import (
	"errors"
	"marketplace/common"
	"strings"
	"sync"
	"time"
)

type ItemService struct {
	store     *Store
	userSvc   *UserService
	itemLocks map[string]*sync.Mutex
	locksMu   sync.Mutex
}

func NewItemService(store *Store, userSvc *UserService) *ItemService {
	return &ItemService{
		store:     store,
		userSvc:   userSvc,
		itemLocks: make(map[string]*sync.Mutex),
	}
}

func (s *ItemService) getItemLock(itemID string) *sync.Mutex {
	s.locksMu.Lock()
	defer s.locksMu.Unlock()
	if lock, ok := s.itemLocks[itemID]; ok {
		return lock
	}
	lock := &sync.Mutex{}
	s.itemLocks[itemID] = lock
	return lock
}

func (s *ItemService) Create(sellerID string, req common.CreateItemRequest) (*common.Item, error) {
	if _, err := s.userSvc.Get(sellerID); err != nil {
		return nil, errors.New("卖家不存在")
	}

	if req.Title == "" {
		return nil, errors.New("标题不能为空")
	}
	if req.Description == "" {
		return nil, errors.New("描述不能为空")
	}
	if !isValidCategory(req.Category) {
		return nil, errors.New("无效的分类")
	}
	if !isValidCondition(req.Condition) {
		return nil, errors.New("无效的新旧程度")
	}
	if req.OriginalPrice <= 0 {
		return nil, errors.New("原价必须大于0")
	}
	if req.Price <= 0 {
		return nil, errors.New("售价必须大于0")
	}

	count := s.store.CountActiveItemsBySeller(sellerID)
	if count >= common.MaxActiveItemsPerUser {
		return nil, errors.New("在售商品数量已达上限")
	}

	item := &common.Item{
		ID:            generateID(),
		SellerID:      sellerID,
		Title:         req.Title,
		Description:   req.Description,
		Category:      req.Category,
		Condition:     req.Condition,
		OriginalPrice: req.OriginalPrice,
		Price:         req.Price,
		Status:        common.ItemStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	s.store.SaveItem(item)
	return item, nil
}

func (s *ItemService) Approve(adminID, itemID string, approved bool, reason string) (*common.Item, error) {
	if !s.userSvc.IsAdmin(adminID) {
		return nil, errors.New("无权限执行审核操作")
	}

	item, ok := s.store.GetItem(itemID)
	if !ok {
		return nil, errors.New("商品不存在")
	}

	if item.Status != common.ItemStatusPending {
		return nil, errors.New("商品不在待审核状态")
	}

	if approved {
		item.Status = common.ItemStatusOnSale
	} else {
		item.Status = common.ItemStatusRejected
	}

	s.store.SaveItem(item)
	return item, nil
}

func (s *ItemService) ListOnSale() []*common.Item {
	items := s.store.ListItems()
	result := make([]*common.Item, 0)
	for _, item := range items {
		if item.Status == common.ItemStatusOnSale {
			result = append(result, item)
		}
	}
	return result
}

func (s *ItemService) Search(keyword string, category common.Category, minPrice, maxPrice int64, condition common.Condition) []*common.Item {
	items := s.store.ListItems()
	result := make([]*common.Item, 0)
	for _, item := range items {
		if item.Status != common.ItemStatusOnSale {
			continue
		}

		if category != "" && item.Category != category {
			continue
		}
		if condition != "" && item.Condition != condition {
			continue
		}
		if minPrice > 0 && item.Price < minPrice {
			continue
		}
		if maxPrice > 0 && item.Price > maxPrice {
			continue
		}
		if keyword != "" {
			kw := strings.ToLower(keyword)
			if !strings.Contains(strings.ToLower(item.Title), kw) &&
			   !strings.Contains(strings.ToLower(item.Description), kw) {
				continue
			}
		}
		result = append(result, item)
	}
	return result
}

func (s *ItemService) Get(id string) (*common.Item, error) {
	item, ok := s.store.GetItem(id)
	if !ok {
		return nil, errors.New("商品不存在")
	}
	return item, nil
}

func (s *ItemService) Remove(sellerID, itemID string) error {
	item, ok := s.store.GetItem(itemID)
	if !ok {
		return errors.New("商品不存在")
	}
	if item.SellerID != sellerID {
		return errors.New("无权限下架此商品")
	}
	if item.Status != common.ItemStatusOnSale && item.Status != common.ItemStatusPending && item.Status != common.ItemStatusRejected {
		if s.store.HasActiveNegotiationForItem(itemID) {
			return errors.New("商品有正在进行的议价，无法下架")
		}
	}
	if item.Status == common.ItemStatusSold {
		return errors.New("商品已售出，无法下架")
	}
	item.Status = common.ItemStatusRemoved
	s.store.SaveItem(item)
	return nil
}

func isValidCategory(c common.Category) bool {
	for _, valid := range common.ValidCategories {
		if valid == c {
			return true
		}
	}
	return false
}

func isValidCondition(c common.Condition) bool {
	for _, valid := range common.ValidConditions {
		if valid == c {
			return true
		}
	}
	return false
}
