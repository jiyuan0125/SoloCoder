package main

type UserService struct {
	storage *Storage
}

func NewUserService(storage *Storage) *UserService {
	return &UserService{storage: storage}
}

func (s *UserService) GetUserClaims(userID string) []*WarrantyClaim {
	return s.storage.GetClaimsByUserID(userID)
}

func (s *UserService) GetClaimByID(claimID string) (*WarrantyClaim, bool) {
	return s.storage.GetClaimByID(claimID)
}
