package reminders

import "time"

type Campaign struct {
	ID                string    `json:"id"`
	EventID           string    `json:"event_id"`
	FromEmail         string    `json:"from_email"`
	Recipients        []string  `json:"recipients"`
	RecipientCount    int       `json:"recipient_count"`
	Subject           string    `json:"subject"`
	Message           string    `json:"message"`
	Status            string    `json:"status"`
	ProviderMessageID string    `json:"provider_message_id"`
	ErrorMessage      string    `json:"error_message"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
