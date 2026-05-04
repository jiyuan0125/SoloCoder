package store

import (
	"encoding/json"
	"os"
	"sync"
)

type ReferralRelationship struct {
	NewUserID       string `json:"new_user_id"`
	ReferrerID      string `json:"referrer_id"`
	BoundAt         int64  `json:"bound_at"`
	FirstOrderDone  bool   `json:"first_order_done"`
	OrderID         string `json:"order_id,omitempty"`
	PointsAwarded   int64  `json:"points_awarded,omitempty"`
}

type UserData struct {
	UserID               string `json:"user_id"`
	ReferralCode         string `json:"referral_code,omitempty"`
	TotalPoints          int64  `json:"total_points"`
	TotalReferrals       int64  `json:"total_referrals"`
	CompletedFirstOrders int64  `json:"completed_first_orders"`
}

type PersistentData struct {
	ReferralCodes        map[string]string         `json:"referral_codes"`
	UserToCode           map[string]string         `json:"user_to_code"`
	ReferralRelationships map[string]ReferralRelationship `json:"referral_relationships"`
	UserData             map[string]UserData       `json:"user_data"`
	RewardPointsPerOrder int64                     `json:"reward_points_per_order"`
	TotalPointsIssued    int64                     `json:"total_points_issued"`
}

type Store struct {
	mu                    sync.RWMutex
	dataFile              string
	referralCodes         map[string]string
	userToCode            map[string]string
	referralRelationships map[string]ReferralRelationship
	userData              map[string]UserData
	rewardPointsPerOrder  int64
	totalPointsIssued     int64
}

const DefaultRewardPoints = 100

func NewStore(dataFile string) (*Store, error) {
	s := &Store{
		dataFile:              dataFile,
		referralCodes:         make(map[string]string),
		userToCode:            make(map[string]string),
		referralRelationships: make(map[string]ReferralRelationship),
		userData:              make(map[string]UserData),
		rewardPointsPerOrder:  DefaultRewardPoints,
		totalPointsIssued:     0,
	}

	if err := s.load(); err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}

	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		return err
	}

	var pd PersistentData
	if err := json.Unmarshal(data, &pd); err != nil {
		return err
	}

	s.referralCodes = pd.ReferralCodes
	s.userToCode = pd.UserToCode
	s.referralRelationships = pd.ReferralRelationships
	s.userData = pd.UserData
	s.rewardPointsPerOrder = pd.RewardPointsPerOrder
	s.totalPointsIssued = pd.TotalPointsIssued

	return nil
}

func (s *Store) save() error {
	pd := PersistentData{
		ReferralCodes:         s.referralCodes,
		UserToCode:            s.userToCode,
		ReferralRelationships: s.referralRelationships,
		UserData:              s.userData,
		RewardPointsPerOrder:  s.rewardPointsPerOrder,
		TotalPointsIssued:    s.totalPointsIssued,
	}

	data, err := json.MarshalIndent(pd, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.dataFile, data, 0644)
}

func (s *Store) SaveReferralCode(userID, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	codeLower := toLower(code)
	s.referralCodes[codeLower] = userID
	s.userToCode[userID] = code

	if _, exists := s.userData[userID]; !exists {
		s.userData[userID] = UserData{
			UserID:       userID,
			ReferralCode: code,
			TotalPoints:  0,
			TotalReferrals: 0,
			CompletedFirstOrders: 0,
		}
	} else {
		ud := s.userData[userID]
		ud.ReferralCode = code
		s.userData[userID] = ud
	}

	return s.save()
}

func (s *Store) GetUserByReferralCode(code string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID, exists := s.referralCodes[toLower(code)]
	return userID, exists
}

func (s *Store) GetReferralCodeByUser(userID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	code, exists := s.userToCode[userID]
	return code, exists
}

func (s *Store) IsReferralCodeExists(code string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.referralCodes[toLower(code)]
	return exists
}

func (s *Store) BindReferral(newUserID, referrerID string, boundAt int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.referralRelationships[newUserID]; exists {
		return os.ErrExist
	}

	s.referralRelationships[newUserID] = ReferralRelationship{
		NewUserID:      newUserID,
		ReferrerID:     referrerID,
		BoundAt:        boundAt,
		FirstOrderDone: false,
	}

	referrerData := s.userData[referrerID]
	referrerData.TotalReferrals++
	s.userData[referrerID] = referrerData

	return s.save()
}

func (s *Store) GetReferralRelationship(newUserID string) (ReferralRelationship, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rel, exists := s.referralRelationships[newUserID]
	return rel, exists
}

func (s *Store) CompleteFirstOrder(userID, orderID string, points int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rel, exists := s.referralRelationships[userID]
	if !exists {
		return os.ErrNotExist
	}

	if rel.FirstOrderDone {
		return os.ErrExist
	}

	rel.FirstOrderDone = true
	rel.OrderID = orderID
	rel.PointsAwarded = points
	s.referralRelationships[userID] = rel

	referrerData := s.userData[rel.ReferrerID]
	referrerData.CompletedFirstOrders++
	referrerData.TotalPoints += points
	s.userData[rel.ReferrerID] = referrerData

	s.totalPointsIssued += points

	return s.save()
}

func (s *Store) RefundFirstOrder(userID, orderID string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	rel, exists := s.referralRelationships[userID]
	if !exists {
		return 0, os.ErrNotExist
	}

	if !rel.FirstOrderDone || rel.OrderID != orderID {
		return 0, os.ErrInvalid
	}

	pointsToDeduct := rel.PointsAwarded
	rel.FirstOrderDone = false
	rel.OrderID = ""
	rel.PointsAwarded = 0
	s.referralRelationships[userID] = rel

	referrerData := s.userData[rel.ReferrerID]
	referrerData.CompletedFirstOrders--
	if referrerData.TotalPoints >= pointsToDeduct {
		referrerData.TotalPoints -= pointsToDeduct
	} else {
		referrerData.TotalPoints = 0
	}
	s.userData[rel.ReferrerID] = referrerData

	if s.totalPointsIssued >= pointsToDeduct {
		s.totalPointsIssued -= pointsToDeduct
	} else {
		s.totalPointsIssued = 0
	}

	return pointsToDeduct, s.save()
}

func (s *Store) GetUserData(userID string) (UserData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.userData[userID]
	return data, exists
}

func (s *Store) SetRewardPoints(points int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rewardPointsPerOrder = points
	return s.save()
}

func (s *Store) GetRewardPoints() int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.rewardPointsPerOrder
}

func (s *Store) GetStats() (totalReferrals, completedFirstOrders, totalPointsIssued int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return int64(len(s.referralRelationships)), 
		   s.countCompletedFirstOrders(), 
		   s.totalPointsIssued
}

func (s *Store) countCompletedFirstOrders() int64 {
	count := int64(0)
	for _, rel := range s.referralRelationships {
		if rel.FirstOrderDone {
			count++
		}
	}
	return count
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}
