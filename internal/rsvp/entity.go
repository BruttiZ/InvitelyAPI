package rsvp

import "time"

type RSVP struct {
	ID        string    `json:"id"`
	GuestID   string    `json:"guest_id"`
	EventID   string    `json:"event_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
