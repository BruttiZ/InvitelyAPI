package rsvp

type SubmitRequest struct {
	GuestID string `json:"guest_id"`
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}
