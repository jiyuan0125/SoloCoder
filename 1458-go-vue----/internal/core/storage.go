package core

import (
	"merchant-mgmt-system/pkg/common"
	"sync"
)

type Storage struct {
	mu        sync.RWMutex
	customers map[string]*common.Customer
	follows   map[string]*common.FollowRecord
	contracts map[string]*common.Contract

	customerByPhone map[string][]string
}

func NewStorage() *Storage {
	return &Storage{
		customers:       make(map[string]*common.Customer),
		follows:         make(map[string]*common.FollowRecord),
		contracts:       make(map[string]*common.Contract),
		customerByPhone: make(map[string][]string),
	}
}

func (s *Storage) SaveCustomer(c *common.Customer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.customers[c.ID] = c
	key := c.CustomerName + "|" + c.ContactPhone
	s.customerByPhone[key] = []string{c.ID}
}

func (s *Storage) GetCustomer(id string) (*common.Customer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.customers[id]
	return c, ok
}

func (s *Storage) GetAllCustomers() []*common.Customer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*common.Customer, 0, len(s.customers))
	for _, c := range s.customers {
		result = append(result, c)
	}
	return result
}

func (s *Storage) FindCustomerByNamePhone(name, phone string) (*common.Customer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := name + "|" + phone
	ids, ok := s.customerByPhone[key]
	if !ok || len(ids) == 0 {
		return nil, false
	}
	c, ok := s.customers[ids[0]]
	return c, ok
}

func (s *Storage) SaveFollow(f *common.FollowRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.follows[f.ID] = f
}

func (s *Storage) GetFollow(id string) (*common.FollowRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	f, ok := s.follows[id]
	return f, ok
}

func (s *Storage) GetFollowsByCustomer(customerID string) []*common.FollowRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*common.FollowRecord{}
	for _, f := range s.follows {
		if f.CustomerID == customerID {
			result = append(result, f)
		}
	}
	return result
}

func (s *Storage) GetOpenFollowsByCustomer(customerID string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, f := range s.follows {
		if f.CustomerID == customerID && f.Status == common.FollowStatusOpen {
			count++
		}
	}
	return count
}

func (s *Storage) GetAllFollows() []*common.FollowRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*common.FollowRecord, 0, len(s.follows))
	for _, f := range s.follows {
		result = append(result, f)
	}
	return result
}

func (s *Storage) SaveContract(c *common.Contract) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.contracts[c.ID] = c
}

func (s *Storage) GetAllContracts() []*common.Contract {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*common.Contract, 0, len(s.contracts))
	for _, c := range s.contracts {
		result = append(result, c)
	}
	return result
}

func (s *Storage) GetContractsByCustomer(customerID string) []*common.Contract {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*common.Contract{}
	for _, c := range s.contracts {
		if c.CustomerID == customerID {
			result = append(result, c)
		}
	}
	return result
}
