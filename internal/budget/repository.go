package budget

import (
	"context"
	"database/sql"
)

type Repository interface {
	Create(ctx context.Context, item Item, tenantID string) (Item, error)
	Delete(ctx context.Context, id string, tenantID string) error
	EventBelongsToTenant(ctx context.Context, eventID string, tenantID string) (bool, error)
	ListByEvent(ctx context.Context, eventID string, tenantID string) ([]Item, error)
	Update(ctx context.Context, item Item, tenantID string) (Item, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) EventBelongsToTenant(ctx context.Context, eventID string, tenantID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `select exists(select 1 from events where id = $1 and tenant_id = $2)`, eventID, tenantID).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) ListByEvent(ctx context.Context, eventID string, tenantID string) ([]Item, error) {
	query := `
		select b.id, b.event_id, b.description, b.category, b.amount, b.paid, b.created_at, b.updated_at
		from event_budget_items b
		join events e on e.id = b.event_id
		where b.event_id = $1 and e.tenant_id = $2
		order by b.created_at desc`
	rows, err := r.db.QueryContext(ctx, query, eventID, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.EventID, &item.Description, &item.Category, &item.Amount, &item.Paid, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *PostgresRepository) Create(ctx context.Context, item Item, tenantID string) (Item, error) {
	query := `
		insert into event_budget_items (id, event_id, description, category, amount, paid)
		select $1, $2, $3, $4, $5, $6
		where exists(select 1 from events where id = $2 and tenant_id = $7)
		returning id, event_id, description, category, amount, paid, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, item.ID, item.EventID, item.Description, item.Category, item.Amount, item.Paid, tenantID).
		Scan(&item.ID, &item.EventID, &item.Description, &item.Category, &item.Amount, &item.Paid, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *PostgresRepository) Update(ctx context.Context, item Item, tenantID string) (Item, error) {
	query := `
		update event_budget_items b
		set description = $3,
			category = $4,
			amount = $5,
			paid = $6,
			updated_at = now()
		from events e
		where b.id = $1 and b.event_id = e.id and e.tenant_id = $2
		returning b.id, b.event_id, b.description, b.category, b.amount, b.paid, b.created_at, b.updated_at`
	err := r.db.QueryRowContext(ctx, query, item.ID, tenantID, item.Description, item.Category, item.Amount, item.Paid).
		Scan(&item.ID, &item.EventID, &item.Description, &item.Category, &item.Amount, &item.Paid, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r *PostgresRepository) Delete(ctx context.Context, id string, tenantID string) error {
	result, err := r.db.ExecContext(ctx, `
		delete from event_budget_items b
		using events e
		where b.id = $1 and b.event_id = e.id and e.tenant_id = $2`, id, tenantID)
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
