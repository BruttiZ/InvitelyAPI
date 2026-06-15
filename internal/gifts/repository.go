package gifts

import (
	"context"
	"database/sql"
)

type Repository interface {
	Create(ctx context.Context, gift Gift, tenantID string) (Gift, error)
	Delete(ctx context.Context, id string, tenantID string) error
	ListByEvent(ctx context.Context, eventID string, tenantID string) ([]Gift, error)
	Update(ctx context.Context, gift Gift, tenantID string) (Gift, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) ListByEvent(ctx context.Context, eventID string, tenantID string) ([]Gift, error) {
	query := `
		select g.id, g.event_id, g.name, g.description, g.price, g.url, g.reserved, g.reserved_by, g.created_at, g.updated_at
		from event_gifts g
		join events e on e.id = g.event_id
		where g.event_id = $1 and e.tenant_id = $2
		order by g.created_at desc`
	rows, err := r.db.QueryContext(ctx, query, eventID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	gifts := make([]Gift, 0)
	for rows.Next() {
		var gift Gift
		if err := rows.Scan(&gift.ID, &gift.EventID, &gift.Name, &gift.Description, &gift.Price, &gift.URL, &gift.Reserved, &gift.ReservedBy, &gift.CreatedAt, &gift.UpdatedAt); err != nil {
			return nil, err
		}
		gifts = append(gifts, gift)
	}

	return gifts, rows.Err()
}

func (r *PostgresRepository) Create(ctx context.Context, gift Gift, tenantID string) (Gift, error) {
	query := `
		insert into event_gifts (id, event_id, name, description, price, url, reserved, reserved_by)
		select $1, $2, $3, $4, $5, $6, $7, $8
		where exists(select 1 from events where id = $2 and tenant_id = $9)
		returning id, event_id, name, description, price, url, reserved, reserved_by, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, gift.ID, gift.EventID, gift.Name, gift.Description, gift.Price, gift.URL, gift.Reserved, gift.ReservedBy, tenantID).
		Scan(&gift.ID, &gift.EventID, &gift.Name, &gift.Description, &gift.Price, &gift.URL, &gift.Reserved, &gift.ReservedBy, &gift.CreatedAt, &gift.UpdatedAt)
	return gift, err
}

func (r *PostgresRepository) Update(ctx context.Context, gift Gift, tenantID string) (Gift, error) {
	query := `
		update event_gifts g
		set name = $3,
			description = $4,
			price = $5,
			url = $6,
			reserved = $7,
			reserved_by = $8,
			updated_at = now()
		from events e
		where g.id = $1 and g.event_id = e.id and e.tenant_id = $2
		returning g.id, g.event_id, g.name, g.description, g.price, g.url, g.reserved, g.reserved_by, g.created_at, g.updated_at`
	err := r.db.QueryRowContext(ctx, query, gift.ID, tenantID, gift.Name, gift.Description, gift.Price, gift.URL, gift.Reserved, gift.ReservedBy).
		Scan(&gift.ID, &gift.EventID, &gift.Name, &gift.Description, &gift.Price, &gift.URL, &gift.Reserved, &gift.ReservedBy, &gift.CreatedAt, &gift.UpdatedAt)
	return gift, err
}

func (r *PostgresRepository) Delete(ctx context.Context, id string, tenantID string) error {
	result, err := r.db.ExecContext(ctx, `
		delete from event_gifts g
		using events e
		where g.id = $1 and g.event_id = e.id and e.tenant_id = $2`, id, tenantID)
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
