package events

import (
	"context"
	"database/sql"
	"encoding/json"
)

type Repository interface {
	Create(ctx context.Context, event Event) (Event, error)
	Delete(ctx context.Context, id string, tenantID string) error
	FindByID(ctx context.Context, id string) (Event, error)
	List(ctx context.Context, tenantID string) ([]Event, error)
	Update(ctx context.Context, event Event) (Event, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, event Event) (Event, error) {
	query := `
		insert into events (
			id, tenant_id, title, name, description, starts_at, ends_at, location, venue_name, address,
			slug, status, timezone, hero, content, theme, gallery, seo
		)
		values (
			$1, $2, $3, $3, $4, $5, $6, $7, $7, $7,
			$8, 'published', 'America/Sao_Paulo', $9::json, $10::json, $11::json, $12::json, $13::json
		)
		returning id, tenant_id, coalesce(title, name), description, starts_at, coalesce(ends_at, starts_at), coalesce(location, venue_name, ''), slug, created_at, updated_at`
	var endsAt any
	if !event.EndsAt.IsZero() {
		endsAt = event.EndsAt
	}
	hero := mustJSON(map[string]string{"title": event.Title})
	content := mustJSON(map[string]string{"description": event.Description})
	theme := mustJSON(map[string]string{"primary": "#8B5CF6", "accent": "#0EA5E9"})
	emptyArray := mustJSON([]string{})
	seo := mustJSON(map[string]string{"title": event.Title})

	err := r.db.QueryRowContext(ctx, query,
		event.ID,
		event.TenantID,
		event.Title,
		event.Description,
		event.StartsAt,
		endsAt,
		event.Location,
		event.Slug,
		hero,
		content,
		theme,
		emptyArray,
		seo,
	).Scan(&event.ID, &event.TenantID, &event.Title, &event.Description, &event.StartsAt, &event.EndsAt, &event.Location, &event.Slug, &event.CreatedAt, &event.UpdatedAt)

	return event, err
}

func mustJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}

	return string(data)
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

func (r *PostgresRepository) Update(ctx context.Context, event Event) (Event, error) {
	query := `
		update events
		set title = $3,
			name = $3,
			description = $4,
			starts_at = $5,
			ends_at = $6,
			location = $7,
			venue_name = $7,
			address = $7,
			updated_at = now()
		where id = $1 and tenant_id = $2
		returning id, tenant_id, coalesce(title, name), description, starts_at, coalesce(ends_at, starts_at), coalesce(location, venue_name, ''), slug, created_at, updated_at`
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
	).Scan(&event.ID, &event.TenantID, &event.Title, &event.Description, &event.StartsAt, &event.EndsAt, &event.Location, &event.Slug, &event.CreatedAt, &event.UpdatedAt)
	return event, err
}

func (r *PostgresRepository) Delete(ctx context.Context, id string, tenantID string) error {
	result, err := r.db.ExecContext(ctx, `delete from events where id = $1 and tenant_id = $2`, id, tenantID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
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
