package events

import "errors"

func ValidateCreate(req CreateEventRequest) error {
	if req.Title == "" {
		return errors.New("title is required")
	}
	if !req.EndsAt.IsZero() && req.EndsAt.Before(req.StartsAt) {
		return errors.New("ends_at must be after starts_at")
	}
	return nil
}
