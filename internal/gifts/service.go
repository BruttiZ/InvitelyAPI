package gifts

import (
	"context"
	"errors"
	"strings"

	"invitely-api/pkg/uuid"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListByEvent(ctx context.Context, tenantID string, eventID string) ([]Gift, error) {
	if strings.TrimSpace(eventID) == "" {
		return []Gift{}, nil
	}
	return s.repository.ListByEvent(ctx, strings.TrimSpace(eventID), tenantID)
}

func (s *Service) Create(ctx context.Context, tenantID string, eventID string, request CreateGiftRequest) (Gift, error) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return Gift{}, errors.New("event id is required")
	}
	if strings.TrimSpace(request.Name) == "" {
		return Gift{}, errors.New("name is required")
	}
	if request.Price < 0 {
		return Gift{}, errors.New("price must be greater than or equal to zero")
	}

	id, err := uuid.New()
	if err != nil {
		return Gift{}, err
	}

	gift := Gift{
		ID:          id,
		EventID:     eventID,
		Name:        strings.TrimSpace(request.Name),
		Description: strings.TrimSpace(request.Description),
		Price:       request.Price,
		URL:         strings.TrimSpace(request.URL),
		Reserved:    request.Reserved,
		ReservedBy:  strings.TrimSpace(request.ReservedBy),
	}

	return s.repository.Create(ctx, gift, tenantID)
}

func (s *Service) Update(ctx context.Context, tenantID string, id string, request UpdateGiftRequest) (Gift, error) {
	if strings.TrimSpace(id) == "" {
		return Gift{}, errors.New("gift id is required")
	}
	if strings.TrimSpace(request.Name) == "" {
		return Gift{}, errors.New("name is required")
	}
	if request.Price < 0 {
		return Gift{}, errors.New("price must be greater than or equal to zero")
	}

	gift := Gift{
		ID:          strings.TrimSpace(id),
		Name:        strings.TrimSpace(request.Name),
		Description: strings.TrimSpace(request.Description),
		Price:       request.Price,
		URL:         strings.TrimSpace(request.URL),
		Reserved:    request.Reserved,
		ReservedBy:  strings.TrimSpace(request.ReservedBy),
	}

	return s.repository.Update(ctx, gift, tenantID)
}

func (s *Service) Delete(ctx context.Context, tenantID string, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("gift id is required")
	}
	return s.repository.Delete(ctx, strings.TrimSpace(id), tenantID)
}
