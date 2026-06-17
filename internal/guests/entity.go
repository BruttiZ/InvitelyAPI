package guests

import "time"

type Guest struct {
	ID            string     `json:"id"`
	EventID       string     `json:"event_id"`
	Name          string     `json:"name"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone"`
	Status        string     `json:"status"`
	PartySize     int        `json:"party_size"`
	MaxCompanions int        `json:"max_companions"`
	InviteToken   string     `json:"invite_token,omitempty"`
	LastSeenAt    *time.Time `json:"last_seen_at,omitempty"`
	RSVP          *GuestRSVP `json:"rsvp,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type GuestRSVP struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	Companions int       `json:"companions"`
	Message    string    `json:"message"`
	Source     string    `json:"source"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
