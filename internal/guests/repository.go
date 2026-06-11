package guests

import (
	"context"
	"database/sql"
)

type Repository interface {
	Create(ctx context.Context, guest Guest) (Guest, error)
	EventBelongsToTenant(ctx context.Context, eventID string, tenantID string) (bool, error)
	ListByEvent(ctx context.Context, eventID string, tenantID string) ([]Guest, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, guest Guest) (Guest, error) {
	query := `
		insert into guests (id, event_id, name, email, phone, status)
		values ($1, $2, $3, $4, $5, $6)
		returning id, event_id, name, email, phone, status, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, guest.ID, guest.EventID, guest.Name, guest.Email, guest.Phone, guest.Status).
		Scan(&guest.ID, &guest.EventID, &guest.Name, &guest.Email, &guest.Phone, &guest.Status, &guest.CreatedAt, &guest.UpdatedAt)
	return guest, err
}

func (r *PostgresRepository) EventBelongsToTenant(ctx context.Context, eventID string, tenantID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `select exists(select 1 from events where id = $1 and tenant_id = $2)`, eventID, tenantID).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) ListByEvent(ctx context.Context, eventID string, tenantID string) ([]Guest, error) {
	query := `
		select g.id, g.event_id, g.name, g.email, g.phone, g.status, g.created_at, g.updated_at
		from guests g
		join events e on e.id = g.event_id
		where g.event_id = $1 and e.tenant_id = $2
		order by g.created_at desc`
	rows, err := r.db.QueryContext(ctx, query, eventID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	guests := make([]Guest, 0)
	for rows.Next() {
		var guest Guest
		if err := rows.Scan(&guest.ID, &guest.EventID, &guest.Name, &guest.Email, &guest.Phone, &guest.Status, &guest.CreatedAt, &guest.UpdatedAt); err != nil {
			return nil, err
		}
		guests = append(guests, guest)
	}

	return guests, rows.Err()
}
