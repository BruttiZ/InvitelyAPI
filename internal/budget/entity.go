package budget

import "time"

type Item struct {
	ID          string    `json:"id"`
	EventID     string    `json:"event_id"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Amount      float64   `json:"amount"`
	Paid        bool      `json:"paid"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
