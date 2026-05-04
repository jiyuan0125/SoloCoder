package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Store struct {
	mu             sync.RWMutex
	Users          map[string]*User
	ReferralCodes  map[string]string
	ReferralRelations map[string]*ReferralRelation
	PointsTransactions []*PointsTransaction
	UserPoints     map[string]int64
	SystemConfig   *SystemConfig
	NextTxnID      int64
	dataFile       string
}

func NewStore(dataFile string) *Store {
	s := &Store{
		Users:           make(map[string]*User),
		ReferralCodes:   make(map[string]string),
		ReferralRelations: make(map[string]*ReferralRelation),
		PointsTransactions: make([]*PointsTransaction, 0),
		UserPoints:      make(map[string]int64),
		SystemConfig: &SystemConfig{
			RewardPointsPerFirstOrder: 100,
		},
		NextTxnID: 1,
		dataFile:   dataFile,
	}
	s.Load()
	return s
}

func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	err = os.WriteFile(s.dataFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write data file: %w", err)
	}

	return nil
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, err := os.Stat(s.dataFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return fmt.Errorf("failed to read data file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	err = json.Unmarshal(data, s)
	if err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

func (s *Store) GetUserByID(userID string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Users[userID]
}

func (s *Store) CreateUser(userID string) *User {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	user := &User{
		ID: userID,
	}
	s.Users[userID] = user
	s.Save()
	return user
}

func (s *Store) ReferralCodeExists(code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.ReferralCodes[code]
	return exists
}

func (s *Store) SetReferralCode(userID, code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	user, exists := s.Users[userID]
	if !exists {
		return
	}
	
	user.ReferralCode = code
	s.ReferralCodes[code] = userID
	s.Save()
}

func (s *Store) GetUserIDByReferralCode(code string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userID, exists := s.ReferralCodes[code]
	return userID, exists
}

func (s *Store) GetReferralRelation(refereeID string) *ReferralRelation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ReferralRelations[refereeID]
}

func (s *Store) CreateReferralRelation(referrerID, refereeID string) *ReferralRelation {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	relation := &ReferralRelation{
		ReferrerID:     referrerID,
		RefereeID:      refereeID,
		BindTime:       getCurrentTime(),
		FirstOrderPaid: false,
		RewardPoints:   0,
	}
	
	s.ReferralRelations[refereeID] = relation
	s.Save()
	return relation
}

func (s *Store) UpdateReferralRelation(relation *ReferralRelation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Save()
}

func (s *Store) GetReferralsByReferrer(referrerID string) []*ReferralRelation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var relations []*ReferralRelation
	for _, relation := range s.ReferralRelations {
		if relation.ReferrerID == referrerID {
			relations = append(relations, relation)
		}
	}
	return relations
}

func (s *Store) GetAllReferralRelations() []*ReferralRelation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var relations []*ReferralRelation
	for _, relation := range s.ReferralRelations {
		relations = append(relations, relation)
	}
	return relations
}

func (s *Store) GetUserPoints(userID string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.UserPoints[userID]
}

func (s *Store) AddPoints(userID string, amount int64, txnType PointsTransactionType, description, refOrderID string) *PointsTransaction {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.UserPoints[userID] += amount
	
	txn := &PointsTransaction{
		ID:          fmt.Sprintf("txn_%d", s.NextTxnID),
		UserID:      userID,
		Amount:      amount,
		Type:        txnType,
		Description: description,
		RefOrderID:  refOrderID,
		CreatedAt:   getCurrentTime(),
	}
	
	s.PointsTransactions = append(s.PointsTransactions, txn)
	s.NextTxnID++
	s.Save()
	
	return txn
}

func (s *Store) SubtractPoints(userID string, amount int64, txnType PointsTransactionType, description, refOrderID string) *PointsTransaction {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	currentPoints := s.UserPoints[userID]
	if currentPoints < amount {
		amount = currentPoints
	}
	
	s.UserPoints[userID] -= amount
	
	if s.UserPoints[userID] < 0 {
		s.UserPoints[userID] = 0
	}
	
	txn := &PointsTransaction{
		ID:          fmt.Sprintf("txn_%d", s.NextTxnID),
		UserID:      userID,
		Amount:      -amount,
		Type:        txnType,
		Description: description,
		RefOrderID:  refOrderID,
		CreatedAt:   getCurrentTime(),
	}
	
	s.PointsTransactions = append(s.PointsTransactions, txn)
	s.NextTxnID++
	s.Save()
	
	return txn
}

func (s *Store) GetRewardPointsPerFirstOrder() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.SystemConfig.RewardPointsPerFirstOrder
}

func (s *Store) SetRewardPointsPerFirstOrder(points int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SystemConfig.RewardPointsPerFirstOrder = points
	s.Save()
}

func (s *Store) GetPointsTransactionsByUser(userID string) []*PointsTransaction {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var transactions []*PointsTransaction
	for _, txn := range s.PointsTransactions {
		if txn.UserID == userID {
			transactions = append(transactions, txn)
		}
	}
	return transactions
}

func getCurrentTime() time.Time {
	return time.Now().UTC()
}
