package core

import (
	"errors"
	"time"
)

var (
	ErrMemberLevelNotFound = errors.New("member level not found")
	ErrMemberNotFound      = errors.New("member not found")
	ErrInvalidDiscountRate = errors.New("invalid discount rate, must be between 0 and 1")
)

func (s *Store) CreateMemberLevel(name string, discountRate float64) (string, error) {
	if discountRate < 0 || discountRate > 1 {
		return "", ErrInvalidDiscountRate
	}

	s.Lock()
	defer s.Unlock()

	id := s.idGen.Generate()
	level := &MemberLevel{
		ID:           id,
		Name:         name,
		DiscountRate: discountRate,
		CreatedAt:    time.Now(),
	}

	s.memberLevels[id] = level
	return id, nil
}

func (s *Store) GetMemberLevel(id string) (*MemberLevel, error) {
	s.RLock()
	defer s.RUnlock()

	level, exists := s.memberLevels[id]
	if !exists {
		return nil, ErrMemberLevelNotFound
	}

	copy := *level
	return &copy, nil
}

func (s *Store) CreateMember(name, phone, memberLevelID string) (string, error) {
	s.Lock()
	defer s.Unlock()

	if _, exists := s.memberLevels[memberLevelID]; !exists {
		return "", ErrMemberLevelNotFound
	}

	id := s.idGen.Generate()
	member := &Member{
		ID:            id,
		Name:          name,
		Phone:         phone,
		MemberLevelID: memberLevelID,
		CreatedAt:     time.Now(),
	}

	s.members[id] = member
	return id, nil
}

func (s *Store) GetMember(id string) (*Member, error) {
	s.RLock()
	defer s.RUnlock()

	member, exists := s.members[id]
	if !exists {
		return nil, ErrMemberNotFound
	}

	copy := *member
	return &copy, nil
}
