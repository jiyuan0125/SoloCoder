package core

import (
	"regexp"
	"strings"
	"time"

	"community-activity-platform/pkg/common"
)

type ParticipantService struct {
	storage *Storage
}

func NewParticipantService(storage *Storage) *ParticipantService {
	return &ParticipantService{storage: storage}
}

func (s *ParticipantService) Register(req common.RegisterParticipantRequest) (*common.Participant, error) {
	req.Phone = strings.TrimSpace(req.Phone)
	req.RoomNumber = strings.TrimSpace(req.RoomNumber)
	req.Name = strings.TrimSpace(req.Name)

	if req.Phone == "" {
		return nil, common.ErrMissingRequiredField
	}
	if req.RoomNumber == "" {
		return nil, common.ErrMissingRequiredField
	}
	if req.Name == "" {
		return nil, common.ErrMissingRequiredField
	}

	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)
	if !phoneRegex.MatchString(req.Phone) {
		return nil, common.ErrInvalidPhoneNumber
	}

	if s.storage.PhoneExists(req.Phone) {
		return nil, common.ErrPhoneAlreadyRegistered
	}

	if s.storage.RoomExists(req.RoomNumber) {
		return nil, common.ErrRoomNumberAlreadyInUse
	}

	participant := &common.Participant{
		ID:          generateID(),
		Phone:       req.Phone,
		PhoneLast4:  getPhoneLast4(req.Phone),
		RoomNumber:  req.RoomNumber,
		Name:        req.Name,
		Age:         req.Age,
		CreateTime:  time.Now(),
	}

	s.storage.StoreParticipant(participant)
	return participant, nil
}

func (s *ParticipantService) GetByID(id string) (*common.Participant, error) {
	p, ok := s.storage.GetParticipant(id)
	if !ok {
		return nil, common.ErrParticipantNotFound
	}
	return p, nil
}

func (s *ParticipantService) GetByPhone(phone string) (*common.Participant, error) {
	p, ok := s.storage.GetParticipantByPhone(phone)
	if !ok {
		return nil, common.ErrParticipantNotFound
	}
	return p, nil
}
