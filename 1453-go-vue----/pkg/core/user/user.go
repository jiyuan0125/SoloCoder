package user

import (
	"errors"
	"property-management/pkg/core"
	"property-management/pkg/core/models"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	store *core.Store
}

func NewService(store *core.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Create(username, name string, role models.Role, building, unit, phone string) (*models.User, error) {
	s.store.Lock()
	defer s.store.Unlock()

	if username == "" || name == "" {
		return nil, errors.New("username and name are required")
	}

	for _, u := range s.store.Users() {
		if u.Username == username {
			return nil, errors.New("username already exists")
		}
	}

	now := time.Now()
	user := &models.User{
		ID:        uuid.New().String(),
		Username:  username,
		Name:      name,
		Role:      role,
		Building:  building,
		Unit:      unit,
		Phone:     phone,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.store.SetUser(user.ID, user)
	return user, nil
}

func (s *Service) GetByID(id string) (*models.User, error) {
	s.store.RLock()
	defer s.store.RUnlock()

	user, ok := s.store.Users()[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *Service) GetByUsername(username string) (*models.User, error) {
	s.store.RLock()
	defer s.store.RUnlock()

	for _, user := range s.store.Users() {
		if user.Username == username {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (s *Service) ListAll() []*models.User {
	s.store.RLock()
	defer s.store.RUnlock()

	var result []*models.User
	for _, user := range s.store.Users() {
		result = append(result, user)
	}
	return result
}

func (s *Service) ListByRole(role models.Role) []*models.User {
	s.store.RLock()
	defer s.store.RUnlock()

	var result []*models.User
	for _, user := range s.store.Users() {
		if user.Role == role {
			result = append(result, user)
		}
	}
	return result
}
