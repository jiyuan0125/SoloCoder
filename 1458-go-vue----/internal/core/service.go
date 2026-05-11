package core

import "sync"

type Service struct {
	storage  *Storage
	createMu sync.Mutex
}

func NewService() *Service {
	return &Service{
		storage: NewStorage(),
	}
}
