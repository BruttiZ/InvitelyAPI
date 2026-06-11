package guests

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

func (s *Service) Create(ctx context.Context, tenantID string, request CreateGuestRequest) (Guest, error) {
	if strings.TrimSpace(request.EventID) == "" {
		return Guest{}, errors.New("event_id is required")
	}
	if strings.TrimSpace(request.Name) == "" {
		return Guest{}, errors.New("name is required")
	}
	if strings.TrimSpace(request.Email) == "" {
		return Guest{}, errors.New("email is required")
	}
	ok, err := s.repository.EventBelongsToTenant(ctx, strings.TrimSpace(request.EventID), tenantID)
	if err != nil {
		return Guest{}, err
	}
	if !ok {
		return Guest{}, errors.New("event not found")
	}

	id, err := uuid.New()
	if err != nil {
		return Guest{}, err
	}

	guest := Guest{
		ID:      id,
		EventID: strings.TrimSpace(request.EventID),
		Name:    strings.TrimSpace(request.Name),
		Email:   strings.ToLower(strings.TrimSpace(request.Email)),
		Phone:   strings.TrimSpace(request.Phone),
		Status:  "invited",
	}

	return s.repository.Create(ctx, guest)
}

func (s *Service) ListByEvent(ctx context.Context, tenantID string, eventID string) ([]Guest, error) {
	if strings.TrimSpace(eventID) == "" {
		return []Guest{}, nil
	}
	return s.repository.ListByEvent(ctx, eventID, tenantID)
}
