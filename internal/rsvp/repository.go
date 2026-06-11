package rsvp

import (
	"context"
	"database/sql"
)

type Repository interface {
	Upsert(ctx context.Context, response RSVP) (RSVP, error)
	FindByGuest(ctx context.Context, guestID string) (RSVP, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Upsert(ctx context.Context, response RSVP) (RSVP, error) {
	query := `
		insert into rsvps (id, guest_id, event_id, status)
		values ($1, $2, $3, $4)
		on conflict (guest_id) do update set
			status = excluded.status,
			updated_at = now()
		returning id, guest_id, event_id, status, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, response.ID, response.GuestID, response.EventID, response.Status).
		Scan(&response.ID, &response.GuestID, &response.EventID, &response.Status, &response.CreatedAt, &response.UpdatedAt)
	if err == nil {
		_, err = r.db.ExecContext(ctx, `update guests set status = $1, updated_at = now() where id = $2`, response.Status, response.GuestID)
	}
	return response, err
}

func (r *PostgresRepository) FindByGuest(ctx context.Context, guestID string) (RSVP, error) {
	var response RSVP
	query := `
		select id, guest_id, event_id, status, created_at, updated_at
		from rsvps
		where guest_id = $1`
	err := r.db.QueryRowContext(ctx, query, guestID).
		Scan(&response.ID, &response.GuestID, &response.EventID, &response.Status, &response.CreatedAt, &response.UpdatedAt)
	return response, err
}
