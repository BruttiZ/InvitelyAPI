package uploads

import (
	"context"
	"io"
)

type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{storage: storage}
}

func (s *Service) Save(ctx context.Context, filename string, content io.Reader) (UploadResponse, error) {
	return s.storage.Save(ctx, filename, content)
}
