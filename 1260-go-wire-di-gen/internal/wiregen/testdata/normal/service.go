package normal

import "context"

type Service struct {
	repo *Repository
}

func (s *Service) Process(ctx context.Context) string {
	return s.repo.GetData()
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}
