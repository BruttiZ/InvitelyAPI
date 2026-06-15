package budget

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

func (s *Service) ListByEvent(ctx context.Context, tenantID string, eventID string) ([]Item, error) {
	if strings.TrimSpace(eventID) == "" {
		return []Item{}, nil
	}
	return s.repository.ListByEvent(ctx, strings.TrimSpace(eventID), tenantID)
}

func (s *Service) Create(ctx context.Context, tenantID string, eventID string, request CreateItemRequest) (Item, error) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return Item{}, errors.New("event id is required")
	}
	if strings.TrimSpace(request.Description) == "" {
		return Item{}, errors.New("description is required")
	}
	if request.Amount < 0 {
		return Item{}, errors.New("amount must be greater than or equal to zero")
	}

	id, err := uuid.New()
	if err != nil {
		return Item{}, err
	}

	item := Item{
		ID:          id,
		EventID:     eventID,
		Description: strings.TrimSpace(request.Description),
		Category:    strings.TrimSpace(request.Category),
		Amount:      request.Amount,
		Paid:        request.Paid,
	}

	return s.repository.Create(ctx, item, tenantID)
}

func (s *Service) Update(ctx context.Context, tenantID string, id string, request UpdateItemRequest) (Item, error) {
	if strings.TrimSpace(id) == "" {
		return Item{}, errors.New("budget item id is required")
	}
	if strings.TrimSpace(request.Description) == "" {
		return Item{}, errors.New("description is required")
	}
	if request.Amount < 0 {
		return Item{}, errors.New("amount must be greater than or equal to zero")
	}

	item := Item{
		ID:          strings.TrimSpace(id),
		Description: strings.TrimSpace(request.Description),
		Category:    strings.TrimSpace(request.Category),
		Amount:      request.Amount,
		Paid:        request.Paid,
	}

	return s.repository.Update(ctx, item, tenantID)
}

func (s *Service) Delete(ctx context.Context, tenantID string, id string) error {
	if strings.TrimSpace(id) == "" {
		return errors.New("budget item id is required")
	}
	return s.repository.Delete(ctx, strings.TrimSpace(id), tenantID)
}
