package dashboard

type Overview struct {
	TotalEvents    int `json:"total_events"`
	TotalGuests    int `json:"total_guests"`
	TotalConfirmed int `json:"total_confirmed"`
}
