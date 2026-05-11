package core

type Service struct {
	storage *Storage
}

func NewService() *Service {
	return &Service{
		storage: NewStorage(),
	}
}

func NewServiceWithStorage(storage *Storage) *Service {
	return &Service{
		storage: storage,
	}
}
