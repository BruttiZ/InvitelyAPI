package events

import "time"

type Event struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	Location    string    `json:"location"`
	Slug        string    `json:"slug"`
	Status      string    `json:"status"`
	TemplateID  string    `json:"templateId"`
	Theme       any       `json:"theme,omitempty"`
	Image       string    `json:"image"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
