package rsvp

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

func (s *Service) Submit(ctx context.Context, request SubmitRequest) (RSVP, error) {
	status := strings.ToLower(strings.TrimSpace(request.Status))
	if status != "confirmed" && status != "declined" && status != "pending" {
		return RSVP{}, errors.New("status must be confirmed, declined or pending")
	}
	if strings.TrimSpace(request.GuestID) == "" {
		return RSVP{}, errors.New("guest_id is required")
	}
	if strings.TrimSpace(request.EventID) == "" {
		return RSVP{}, errors.New("event_id is required")
	}

	id, err := uuid.New()
	if err != nil {
		return RSVP{}, err
	}

	return s.repository.Upsert(ctx, RSVP{
		ID:      id,
		GuestID: strings.TrimSpace(request.GuestID),
		EventID: strings.TrimSpace(request.EventID),
		Status:  status,
	})
}

func (s *Service) FindByGuest(ctx context.Context, guestID string) (RSVP, error) {
	return s.repository.FindByGuest(ctx, guestID)
}
