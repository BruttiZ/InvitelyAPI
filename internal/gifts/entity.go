package gifts

import "time"

type Gift struct {
	ID          string    `json:"id"`
	EventID     string    `json:"event_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	URL         string    `json:"url"`
	Reserved    bool      `json:"reserved"`
	ReservedBy  string    `json:"reserved_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
