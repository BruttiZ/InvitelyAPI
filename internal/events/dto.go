package events

import "time"

type CreateEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	Location    string    `json:"location"`
	TemplateID  string    `json:"templateId"`
	Theme       any       `json:"theme"`
	Image       string    `json:"image"`
}

type UpdateEventRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	Location    string    `json:"location"`
	TemplateID  string    `json:"templateId"`
	Theme       any       `json:"theme"`
	Image       string    `json:"image"`
}
