package analytics

import (
	"context"
	"database/sql"
)

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) EventSummary(ctx context.Context, eventID string) (EventSummary, error) {
	summary := EventSummary{EventID: eventID}
	query := `
		select
			(select count(*) from guests where event_id = $1),
			(select count(*) from guests where event_id = $1 and status in ('accepted', 'confirmed')),
			(select count(*) from guests where event_id = $1 and status = 'declined'),
			(select count(*) from guests where event_id = $1 and status in ('invited', 'pending'))`
	err := s.db.QueryRowContext(ctx, query, eventID).Scan(&summary.GuestCount, &summary.ConfirmedCount, &summary.DeclinedCount, &summary.PendingCount)
	return summary, err
}
