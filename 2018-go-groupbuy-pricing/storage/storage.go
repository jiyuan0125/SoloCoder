package storage

import (
	"encoding/json"
	"fmt"
	"groupbuy/models"
	"os"
	"path/filepath"
	"sync"
)

type DataStore struct {
	Users           map[string]models.User           `json:"users"`
	Activities      map[string]models.Activity       `json:"activities"`
	Orders          map[string]models.Order          `json:"orders"`
	RefundRecords   map[string]models.RefundRecord   `json:"refund_records"`
	WithdrawalRecords map[string]models.WithdrawalRecord `json:"withdrawal_records"`
}

type Storage struct {
	filePath string
	mu       sync.RWMutex
}

func NewStorage(filePath string) *Storage {
	return &Storage{filePath: filePath}
}

func (s *Storage) ensureFile() error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return s.save(&DataStore{
			Users:              make(map[string]models.User),
			Activities:         make(map[string]models.Activity),
			Orders:             make(map[string]models.Order),
			RefundRecords:      make(map[string]models.RefundRecord),
			WithdrawalRecords:  make(map[string]models.WithdrawalRecord),
		})
	}
	return nil
}

func (s *Storage) load() (*DataStore, error) {
	if err := s.ensureFile(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}
	var store DataStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}
	if store.Users == nil {
		store.Users = make(map[string]models.User)
	}
	if store.Activities == nil {
		store.Activities = make(map[string]models.Activity)
	}
	if store.Orders == nil {
		store.Orders = make(map[string]models.Order)
	}
	if store.RefundRecords == nil {
		store.RefundRecords = make(map[string]models.RefundRecord)
	}
	if store.WithdrawalRecords == nil {
		store.WithdrawalRecords = make(map[string]models.WithdrawalRecord)
	}
	return &store, nil
}

func (s *Storage) save(store *DataStore) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *Storage) ReadTransaction(fn func(store *DataStore) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	store, err := s.load()
	if err != nil {
		return err
	}
	return fn(store)
}

func (s *Storage) WriteTransaction(fn func(store *DataStore) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	store, err := s.load()
	if err != nil {
		return err
	}
	if err := fn(store); err != nil {
		return err
	}
	return s.save(store)
}

func (s *Storage) GetOrCreateUser(userID string) (*models.User, error) {
	var user *models.User
	err := s.WriteTransaction(func(store *DataStore) error {
		if u, exists := store.Users[userID]; exists {
			user = &u
			return nil
		}
		u := models.User{UserID: userID, Balance: 0}
		store.Users[userID] = u
		user = &u
		return nil
	})
	return user, err
}

func (s *Storage) GetUser(userID string) (*models.User, error) {
	var user *models.User
	err := s.ReadTransaction(func(store *DataStore) error {
		if u, exists := store.Users[userID]; exists {
			user = &u
			return nil
		}
		return fmt.Errorf("用户不存在: %s", userID)
	})
	return user, err
}

func (s *Storage) SaveActivity(activity models.Activity) error {
	return s.WriteTransaction(func(store *DataStore) error {
		store.Activities[activity.ActivityID] = activity
		return nil
	})
}

func (s *Storage) GetActivity(activityID string) (*models.Activity, error) {
	var activity *models.Activity
	err := s.ReadTransaction(func(store *DataStore) error {
		if a, exists := store.Activities[activityID]; exists {
			activity = &a
			return nil
		}
		return fmt.Errorf("活动不存在: %s", activityID)
	})
	return activity, err
}

func (s *Storage) ListActivities() ([]models.Activity, error) {
	var activities []models.Activity
	err := s.ReadTransaction(func(store *DataStore) error {
		for _, a := range store.Activities {
			activities = append(activities, a)
		}
		return nil
	})
	return activities, err
}

func (s *Storage) SaveOrder(order models.Order) error {
	return s.WriteTransaction(func(store *DataStore) error {
		store.Orders[order.OrderID] = order
		return nil
	})
}

func (s *Storage) GetOrder(orderID string) (*models.Order, error) {
	var order *models.Order
	err := s.ReadTransaction(func(store *DataStore) error {
		if o, exists := store.Orders[orderID]; exists {
			order = &o
			return nil
		}
		return fmt.Errorf("订单不存在: %s", orderID)
	})
	return order, err
}

func (s *Storage) GetOrdersByActivity(activityID string) ([]models.Order, error) {
	var orders []models.Order
	err := s.ReadTransaction(func(store *DataStore) error {
		for _, o := range store.Orders {
			if o.ActivityID == activityID {
				orders = append(orders, o)
			}
		}
		return nil
	})
	return orders, err
}

func (s *Storage) GetOrdersByUser(userID string) ([]models.Order, error) {
	var orders []models.Order
	err := s.ReadTransaction(func(store *DataStore) error {
		for _, o := range store.Orders {
			if o.UserID == userID {
				orders = append(orders, o)
			}
		}
		return nil
	})
	return orders, err
}

func (s *Storage) HasUserParticipated(activityID, userID string) bool {
	orders, _ := s.GetOrdersByActivity(activityID)
	for _, o := range orders {
		if o.UserID == userID && o.Status != "cancelled" {
			return true
		}
	}
	return false
}

func (s *Storage) SaveRefundRecord(record models.RefundRecord) error {
	return s.WriteTransaction(func(store *DataStore) error {
		store.RefundRecords[record.RefundID] = record
		return nil
	})
}

func (s *Storage) GetRefundRecordsByUser(userID string) ([]models.RefundRecord, error) {
	var records []models.RefundRecord
	err := s.ReadTransaction(func(store *DataStore) error {
		for _, r := range store.RefundRecords {
			if r.UserID == userID {
				records = append(records, r)
			}
		}
		return nil
	})
	return records, err
}

func (s *Storage) SaveWithdrawalRecord(record models.WithdrawalRecord) error {
	return s.WriteTransaction(func(store *DataStore) error {
		store.WithdrawalRecords[record.WithdrawalID] = record
		return nil
	})
}

func (s *Storage) GetWithdrawalRecordsByUser(userID string) ([]models.WithdrawalRecord, error) {
	var records []models.WithdrawalRecord
	err := s.ReadTransaction(func(store *DataStore) error {
		for _, w := range store.WithdrawalRecords {
			if w.UserID == userID {
				records = append(records, w)
			}
		}
		return nil
	})
	return records, err
}

func (s *Storage) UpdateUserBalance(userID string, delta int64) error {
	return s.WriteTransaction(func(store *DataStore) error {
		u, exists := store.Users[userID]
		if !exists {
			return fmt.Errorf("用户不存在: %s", userID)
		}
		u.Balance += delta
		store.Users[userID] = u
		return nil
	})
}

func (s *Storage) UpdateOrderStatus(orderID, status string) error {
	return s.WriteTransaction(func(store *DataStore) error {
		o, exists := store.Orders[orderID]
		if !exists {
			return fmt.Errorf("订单不存在: %s", orderID)
		}
		o.Status = status
		store.Orders[orderID] = o
		return nil
	})
}
