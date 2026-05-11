package business

import (
	"errors"
	"sync"
	"time"
)

type User struct {
	ID        int64     `json:"id" xml:"id" csv:"id"`
	Name      string    `json:"name" xml:"name" csv:"name"`
	Email     string    `json:"email" xml:"email" csv:"email"`
	Age       int       `json:"age" xml:"age" csv:"age"`
	CreatedAt time.Time `json:"created_at" xml:"created_at" csv:"created_at"`
}

type UserStore struct {
	mu    sync.RWMutex
	users map[int64]*User
	next  int64
}

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

func NewUserStore() *UserStore {
	return &UserStore{
		users: make(map[int64]*User),
		next:  1,
	}
}

func (s *UserStore) Create(user *User) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newUser := &User{
		ID:        s.next,
		Name:      user.Name,
		Email:     user.Email,
		Age:       user.Age,
		CreatedAt: time.Now(),
	}
	s.users[newUser.ID] = newUser
	s.next++
	return newUser, nil
}

func (s *UserStore) Get(id int64) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserStore) Update(user *User) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.users[user.ID]
	if !ok {
		return nil, ErrUserNotFound
	}

	if user.Name != "" {
		existing.Name = user.Name
	}
	if user.Email != "" {
		existing.Email = user.Email
	}
	if user.Age != 0 {
		existing.Age = user.Age
	}

	return existing, nil
}

func (s *UserStore) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[id]; !ok {
		return ErrUserNotFound
	}
	delete(s.users, id)
	return nil
}

func (s *UserStore) List() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users
}
