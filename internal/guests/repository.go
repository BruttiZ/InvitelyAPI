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
		insert into guests (id, event_id, name, email, phone, status, party_size, max_companions, invite_token)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		returning id, event_id, name, email, phone, status, coalesce(party_size, 1), coalesce(max_companions, 5), coalesce(invite_token, ''), last_seen_at, created_at, updated_at`
	var lastSeenAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, guest.ID, guest.EventID, guest.Name, guest.Email, guest.Phone, guest.Status, guest.PartySize, guest.MaxCompanions, guest.InviteToken).
		Scan(&guest.ID, &guest.EventID, &guest.Name, &guest.Email, &guest.Phone, &guest.Status, &guest.PartySize, &guest.MaxCompanions, &guest.InviteToken, &lastSeenAt, &guest.CreatedAt, &guest.UpdatedAt)
	if lastSeenAt.Valid {
		guest.LastSeenAt = &lastSeenAt.Time
	}
	return guest, err
}

func (r *PostgresRepository) EventBelongsToTenant(ctx context.Context, eventID string, tenantID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `select exists(select 1 from events where id = $1 and tenant_id = $2)`, eventID, tenantID).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) ListByEvent(ctx context.Context, eventID string, tenantID string) ([]Guest, error) {
	query := `
		select
			g.id,
			g.event_id,
			g.name,
			g.email,
			g.phone,
			g.status,
			coalesce(g.party_size, 1),
			coalesce(g.max_companions, 5),
			coalesce(g.invite_token, ''),
			g.last_seen_at,
			g.created_at,
			g.updated_at,
			r.id,
			r.status,
			coalesce(r.companions, 0),
			coalesce(r.message, ''),
			coalesce(r.source, ''),
			r.created_at,
			r.updated_at
		from guests g
		join events e on e.id = g.event_id
		left join rsvps r on r.guest_id = g.id
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
		var rsvpID, rsvpStatus, rsvpMessage, rsvpSource sql.NullString
		var rsvpCompanions sql.NullInt64
		var lastSeenAt, rsvpCreatedAt, rsvpUpdatedAt sql.NullTime
		if err := rows.Scan(
			&guest.ID,
			&guest.EventID,
			&guest.Name,
			&guest.Email,
			&guest.Phone,
			&guest.Status,
			&guest.PartySize,
			&guest.MaxCompanions,
			&guest.InviteToken,
			&lastSeenAt,
			&guest.CreatedAt,
			&guest.UpdatedAt,
			&rsvpID,
			&rsvpStatus,
			&rsvpCompanions,
			&rsvpMessage,
			&rsvpSource,
			&rsvpCreatedAt,
			&rsvpUpdatedAt,
		); err != nil {
			return nil, err
		}
		if lastSeenAt.Valid {
			guest.LastSeenAt = &lastSeenAt.Time
		}
		if rsvpID.Valid {
			guest.RSVP = &GuestRSVP{
				ID:         rsvpID.String,
				Status:     rsvpStatus.String,
				Companions: int(rsvpCompanions.Int64),
				Message:    rsvpMessage.String,
				Source:     rsvpSource.String,
				CreatedAt:  rsvpCreatedAt.Time,
				UpdatedAt:  rsvpUpdatedAt.Time,
			}
		}
		guests = append(guests, guest)
	}

	return guests, rows.Err()
}
