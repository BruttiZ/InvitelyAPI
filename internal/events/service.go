package events

import (
	"context"
	"errors"
	"strings"

	"invitely-api/pkg/slug"
	"invitely-api/pkg/uuid"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, tenantID string, request CreateEventRequest) (Event, error) {
	if tenantID == "" {
		return Event{}, errors.New("tenant not found")
	}
	if strings.TrimSpace(request.Title) == "" {
		return Event{}, errors.New("title is required")
	}
	if request.StartsAt.IsZero() {
		return Event{}, errors.New("starts_at is required")
	}

	id, err := uuid.New()
	if err != nil {
		return Event{}, err
	}

	eventSlug := slug.Make(request.Title + "-" + id[:8])
	event := Event{
		ID:          id,
		TenantID:    tenantID,
		Title:       strings.TrimSpace(request.Title),
		Description: strings.TrimSpace(request.Description),
		StartsAt:    request.StartsAt,
		EndsAt:      request.EndsAt,
		Location:    strings.TrimSpace(request.Location),
		Slug:        eventSlug,
		Status:      "published",
		TemplateID:  strings.TrimSpace(request.TemplateID),
		Theme:       request.Theme,
		Image:       strings.TrimSpace(request.Image),
	}

	return s.repository.Create(ctx, event)
}

func (s *Service) List(ctx context.Context, tenantID string) ([]Event, error) {
	if tenantID == "" {
		return []Event{}, nil
	}
	return s.repository.List(ctx, tenantID)
}

func (s *Service) FindByID(ctx context.Context, id string) (Event, error) {
	return s.repository.FindByID(ctx, id)
}

func (s *Service) FindPublicBySlug(ctx context.Context, slug string) (Event, error) {
	return s.repository.FindPublicBySlug(ctx, strings.TrimSpace(slug))
}

func (s *Service) Update(ctx context.Context, tenantID string, id string, request UpdateEventRequest) (Event, error) {
	if tenantID == "" {
		return Event{}, errors.New("tenant not found")
	}
	if strings.TrimSpace(id) == "" {
		return Event{}, errors.New("event id is required")
	}
	if strings.TrimSpace(request.Title) == "" {
		return Event{}, errors.New("title is required")
	}
	if request.StartsAt.IsZero() {
		return Event{}, errors.New("starts_at is required")
	}

	event := Event{
		ID:          strings.TrimSpace(id),
		TenantID:    tenantID,
		Title:       strings.TrimSpace(request.Title),
		Description: strings.TrimSpace(request.Description),
		StartsAt:    request.StartsAt,
		EndsAt:      request.EndsAt,
		Location:    strings.TrimSpace(request.Location),
		TemplateID:  strings.TrimSpace(request.TemplateID),
		Theme:       request.Theme,
		Image:       strings.TrimSpace(request.Image),
	}

	return s.repository.Update(ctx, event)
}

func (s *Service) Delete(ctx context.Context, tenantID string, id string) error {
	if tenantID == "" {
		return errors.New("tenant not found")
	}
	if strings.TrimSpace(id) == "" {
		return errors.New("event id is required")
	}
	return s.repository.Delete(ctx, strings.TrimSpace(id), tenantID)
}
