package events

import (
	"context"
	"database/sql"
)

type Repository interface {
	Create(ctx context.Context, event Event) (Event, error)
	FindByID(ctx context.Context, id string) (Event, error)
	List(ctx context.Context, tenantID string) ([]Event, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, event Event) (Event, error) {
	query := `
		insert into events (id, tenant_id, title, description, starts_at, ends_at, location, slug)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
		returning id, tenant_id, title, description, starts_at, coalesce(ends_at, starts_at), location, slug, created_at, updated_at`
	var endsAt any
	if !event.EndsAt.IsZero() {
		endsAt = event.EndsAt
	}

	err := r.db.QueryRowContext(ctx, query,
		event.ID,
		event.TenantID,
		event.Title,
		event.Description,
		event.StartsAt,
		endsAt,
		event.Location,
		event.Slug,
	).Scan(&event.ID, &event.TenantID, &event.Title, &event.Description, &event.StartsAt, &event.EndsAt, &event.Location, &event.Slug, &event.CreatedAt, &event.UpdatedAt)

	return event, err
}

func (r *PostgresRepository) FindByID(ctx context.Context, id string) (Event, error) {
	var event Event
	query := `
		select id, tenant_id, title, description, starts_at, coalesce(ends_at, starts_at), location, slug, created_at, updated_at
		from events
		where id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&event.ID, &event.TenantID, &event.Title, &event.Description, &event.StartsAt, &event.EndsAt, &event.Location, &event.Slug, &event.CreatedAt, &event.UpdatedAt)
	return event, err
}

func (r *PostgresRepository) List(ctx context.Context, tenantID string) ([]Event, error) {
	query := `
		select id, tenant_id, title, description, starts_at, coalesce(ends_at, starts_at), location, slug, created_at, updated_at
		from events
		where tenant_id = $1
		order by starts_at desc`
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]Event, 0)
	for rows.Next() {
		var event Event
		if err := rows.Scan(&event.ID, &event.TenantID, &event.Title, &event.Description, &event.StartsAt, &event.EndsAt, &event.Location, &event.Slug, &event.CreatedAt, &event.UpdatedAt); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}
