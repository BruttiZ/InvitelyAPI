package rsvp

type SubmitRequest struct {
	GuestID string `json:"guest_id"`
	EventID string `json:"event_id"`
	Status  string `json:"status"`
}

type PublicSubmitRequest struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Status      string `json:"status"`
	Companions  int    `json:"companions"`
	Message     string `json:"message"`
	InviteToken string `json:"invite_token"`
}
