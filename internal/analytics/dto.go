package analytics

type EventSummary struct {
	EventID        string `json:"event_id"`
	GuestCount     int    `json:"guest_count"`
	ConfirmedCount int    `json:"confirmed_count"`
	DeclinedCount  int    `json:"declined_count"`
	PendingCount   int    `json:"pending_count"`
}
