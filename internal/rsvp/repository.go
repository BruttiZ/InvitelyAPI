package rsvp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"invitely-api/pkg/uuid"
)

type Repository interface {
	Upsert(ctx context.Context, response RSVP) (RSVP, error)
	FindByGuest(ctx context.Context, guestID string) (RSVP, error)
	SubmitPublic(ctx context.Context, slug string, ip string, request PublicSubmitRequest) (RSVP, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Upsert(ctx context.Context, response RSVP) (RSVP, error) {
	query := `
		insert into rsvps (id, guest_id, event_id, status, companions, message, source)
		values ($1, $2, $3, $4, $5, $6, $7)
		on conflict (guest_id) do update set
			status = excluded.status,
			companions = excluded.companions,
			message = excluded.message,
			source = excluded.source,
			updated_at = now()
		returning id, guest_id, event_id, status, companions, coalesce(message, ''), coalesce(source, ''), created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, response.ID, response.GuestID, response.EventID, response.Status, response.Companions, response.Message, response.Source).
		Scan(&response.ID, &response.GuestID, &response.EventID, &response.Status, &response.Companions, &response.Message, &response.Source, &response.CreatedAt, &response.UpdatedAt)
	if err == nil {
		_, err = r.db.ExecContext(ctx, `update guests set status = $1, party_size = $2, updated_at = now(), last_seen_at = now() where id = $3`, response.Status, 1+response.Companions, response.GuestID)
	}
	return response, err
}

func (r *PostgresRepository) FindByGuest(ctx context.Context, guestID string) (RSVP, error) {
	var response RSVP
	query := `
		select id, guest_id, event_id, status, coalesce(companions, 0), coalesce(message, ''), coalesce(source, ''), created_at, updated_at
		from rsvps
		where guest_id = $1`
	err := r.db.QueryRowContext(ctx, query, guestID).
		Scan(&response.ID, &response.GuestID, &response.EventID, &response.Status, &response.Companions, &response.Message, &response.Source, &response.CreatedAt, &response.UpdatedAt)
	return response, err
}

func (r *PostgresRepository) SubmitPublic(ctx context.Context, slug string, ip string, request PublicSubmitRequest) (RSVP, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return RSVP{}, err
	}
	defer tx.Rollback()

	var eventID, status string
	var endsAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		select id, coalesce(status, 'draft'), ends_at
		from events
		where slug = $1
		for update`, slug).Scan(&eventID, &status, &endsAt)
	if err != nil {
		return RSVP{}, err
	}
	if !eventAllowsRSVP(status, endsAt) {
		return RSVP{}, ErrPublicRSVPClosed
	}
	if !publicRSVPLimiter.Allow(strings.Join([]string{clientIPKey(ip), request.Email, eventID}, "|")) {
		return RSVP{}, ErrRateLimited
	}

	guest, err := r.findPublicGuest(ctx, tx, eventID, request)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return RSVP{}, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		if strings.TrimSpace(request.InviteToken) != "" {
			return RSVP{}, sql.ErrNoRows
		}
		guest, err = r.createPublicGuest(ctx, tx, eventID, request)
		if err != nil {
			return RSVP{}, err
		}
	}
	if request.Companions > guest.MaxCompanions {
		return RSVP{}, ErrTooManyCompanions
	}

	responseID, err := uuid.New()
	if err != nil {
		return RSVP{}, err
	}

	var response RSVP
	err = tx.QueryRowContext(ctx, `
		insert into rsvps (id, guest_id, event_id, status, companions, message, source)
		values ($1, $2, $3, $4, $5, $6, 'public_link')
		on conflict (guest_id) do update set
			event_id = excluded.event_id,
			status = excluded.status,
			companions = excluded.companions,
			message = excluded.message,
			source = excluded.source,
			updated_at = now()
		returning id, guest_id, event_id, status, companions, coalesce(message, ''), coalesce(source, ''), created_at, updated_at`,
		responseID,
		guest.ID,
		eventID,
		request.Status,
		request.Companions,
		request.Message,
	).Scan(&response.ID, &response.GuestID, &response.EventID, &response.Status, &response.Companions, &response.Message, &response.Source, &response.CreatedAt, &response.UpdatedAt)
	if err != nil {
		return RSVP{}, err
	}

	_, err = tx.ExecContext(ctx, `
		update guests
		set name = $1,
			email = $2,
			status = $3,
			party_size = $4,
			last_seen_at = now(),
			updated_at = now()
		where id = $5`,
		request.Name,
		request.Email,
		request.Status,
		1+request.Companions,
		guest.ID,
	)
	if err != nil {
		return RSVP{}, err
	}

	if _, err := tx.ExecContext(ctx, `delete from public_event_cache where event_id = $1`, eventID); err != nil {
		return RSVP{}, err
	}

	if err := tx.Commit(); err != nil {
		return RSVP{}, err
	}

	return response, nil
}

type publicGuest struct {
	ID            string
	MaxCompanions int
}

func (r *PostgresRepository) findPublicGuest(ctx context.Context, tx *sql.Tx, eventID string, request PublicSubmitRequest) (publicGuest, error) {
	var guest publicGuest
	if strings.TrimSpace(request.InviteToken) != "" {
		err := tx.QueryRowContext(ctx, `
			select id, coalesce(max_companions, 5)
			from guests
			where event_id = $1 and invite_token = $2
			for update`, eventID, strings.TrimSpace(request.InviteToken)).Scan(&guest.ID, &guest.MaxCompanions)
		return guest, err
	}

	err := tx.QueryRowContext(ctx, `
		select id, coalesce(max_companions, 5)
		from guests
		where event_id = $1 and lower(email) = $2
		for update`, eventID, request.Email).Scan(&guest.ID, &guest.MaxCompanions)
	return guest, err
}

func (r *PostgresRepository) createPublicGuest(ctx context.Context, tx *sql.Tx, eventID string, request PublicSubmitRequest) (publicGuest, error) {
	id, err := uuid.New()
	if err != nil {
		return publicGuest{}, err
	}
	token, err := uuid.New()
	if err != nil {
		return publicGuest{}, err
	}
	metadata, err := json.Marshal(map[string]string{"source": "public_event_link"})
	if err != nil {
		return publicGuest{}, err
	}

	var guest publicGuest
	err = tx.QueryRowContext(ctx, `
		insert into guests (id, event_id, name, email, status, party_size, max_companions, invite_token, metadata)
		values ($1, $2, $3, $4, 'invited', 1, 5, $5, $6::jsonb)
		on conflict (event_id, email) do update set
			name = excluded.name,
			updated_at = now()
		returning id, coalesce(max_companions, 5)`,
		id,
		eventID,
		request.Name,
		request.Email,
		token,
		string(metadata),
	).Scan(&guest.ID, &guest.MaxCompanions)
	return guest, err
}

func eventAllowsRSVP(status string, endsAt sql.NullTime) bool {
	normalized := strings.ToLower(strings.TrimSpace(status))
	if normalized != "published" && normalized != "active" {
		return false
	}
	return !endsAt.Valid || endsAt.Time.After(time.Now())
}
