package reminders

import (
	"context"
	"database/sql"
	"encoding/json"
)

type Repository interface {
	CreateCampaign(ctx context.Context, campaign Campaign) (Campaign, error)
	UpdateCampaignStatus(ctx context.Context, id string, status string, providerMessageID string, errorMessage string) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateCampaign(ctx context.Context, campaign Campaign) (Campaign, error) {
	recipients, err := json.Marshal(campaign.Recipients)
	if err != nil {
		return Campaign{}, err
	}

	query := `
		insert into event_reminder_campaigns (
			id, event_id, from_email, recipients, recipient_count, subject, message, status
		)
		values ($1, $2, $3, $4::json, $5, $6, $7, $8)
		returning id, event_id, from_email, recipients, recipient_count, subject, message, status, provider_message_id, error_message, created_at, updated_at`

	var rawRecipients []byte
	err = r.db.QueryRowContext(ctx, query,
		campaign.ID,
		campaign.EventID,
		campaign.FromEmail,
		string(recipients),
		campaign.RecipientCount,
		campaign.Subject,
		campaign.Message,
		campaign.Status,
	).Scan(
		&campaign.ID,
		&campaign.EventID,
		&campaign.FromEmail,
		&rawRecipients,
		&campaign.RecipientCount,
		&campaign.Subject,
		&campaign.Message,
		&campaign.Status,
		&campaign.ProviderMessageID,
		&campaign.ErrorMessage,
		&campaign.CreatedAt,
		&campaign.UpdatedAt,
	)
	if err != nil {
		return Campaign{}, err
	}

	if err := json.Unmarshal(rawRecipients, &campaign.Recipients); err != nil {
		return Campaign{}, err
	}

	return campaign, nil
}

func (r *PostgresRepository) UpdateCampaignStatus(ctx context.Context, id string, status string, providerMessageID string, errorMessage string) error {
	result, err := r.db.ExecContext(ctx, `
		update event_reminder_campaigns
		set status = $2,
			provider_message_id = $3,
			error_message = $4,
			updated_at = now()
		where id = $1`, id, status, providerMessageID, errorMessage)
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
