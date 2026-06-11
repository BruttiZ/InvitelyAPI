package dashboard

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

func (s *Service) Overview(ctx context.Context, tenantID string) (Overview, error) {
	var overview Overview
	query := `
		select
			(select count(*) from events where tenant_id = $1) as total_events,
			(select count(*) from guests g join events e on e.id = g.event_id where e.tenant_id = $1) as total_guests,
			(select count(*) from rsvps r join events e on e.id = r.event_id where e.tenant_id = $1 and r.status = 'confirmed') as total_confirmed`
	err := s.db.QueryRowContext(ctx, query, tenantID).Scan(&overview.TotalEvents, &overview.TotalGuests, &overview.TotalConfirmed)
	return overview, err
}
