package audit

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Record(ctx context.Context, entry Entry) (Entry, error) {
	return s.repository.Create(ctx, entry)
}
