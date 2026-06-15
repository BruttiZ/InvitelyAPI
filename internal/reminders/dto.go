package reminders

type SendRequest struct {
	FromEmail  string   `json:"from_email"`
	Recipients []string `json:"recipients"`
	Subject    string   `json:"subject"`
	Message    string   `json:"message"`
}

type SendResponse struct {
	CampaignID string `json:"campaign_id"`
	EventID    string `json:"event_id"`
	Queued     int    `json:"queued"`
	Status     string `json:"status"`
}

type ValidationIssue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
