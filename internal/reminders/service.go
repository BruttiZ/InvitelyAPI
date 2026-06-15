package reminders

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"invitely-api/internal/common"
	"invitely-api/internal/events"
	"invitely-api/pkg/uuid"
)

const maxRecipients = 200

type Service struct {
	repository       Repository
	eventsRepository events.Repository
	emailSender      EmailSender
}

func NewService(repository Repository, eventsRepository events.Repository, emailSender EmailSender) *Service {
	return &Service{repository: repository, eventsRepository: eventsRepository, emailSender: emailSender}
}

func (s *Service) Send(ctx context.Context, tenantID string, eventID string, request SendRequest) (SendResponse, []ValidationIssue, error) {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return SendResponse{}, nil, errors.New("event id is required")
	}

	event, err := s.eventsRepository.FindByID(ctx, eventID)
	if err != nil {
		return SendResponse{}, nil, err
	}
	if event.TenantID != tenantID {
		return SendResponse{}, nil, common.ErrForbidden
	}

	normalized, issues := validateRequest(request)
	if len(issues) > 0 {
		return SendResponse{}, issues, nil
	}
	if !s.emailSender.Configured() {
		return SendResponse{}, nil, ErrEmailProviderUnavailable
	}

	id, err := uuid.New()
	if err != nil {
		return SendResponse{}, nil, err
	}

	campaign, err := s.repository.CreateCampaign(ctx, Campaign{
		ID:             id,
		EventID:        eventID,
		FromEmail:      normalized.FromEmail,
		Recipients:     normalized.Recipients,
		RecipientCount: len(normalized.Recipients),
		Subject:        normalized.Subject,
		Message:        normalized.Message,
		Status:         "queued",
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return SendResponse{}, nil, err
		}
		return SendResponse{}, nil, err
	}

	s.dispatchAsync(campaign)

	return SendResponse{
		CampaignID: campaign.ID,
		EventID:    campaign.EventID,
		Queued:     campaign.RecipientCount,
		Status:     campaign.Status,
	}, nil, nil
}

func (s *Service) dispatchAsync(campaign Campaign) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		providerMessageID, err := s.emailSender.SendReminder(ctx, EmailMessage{
			FromEmail:  campaign.FromEmail,
			Recipients: campaign.Recipients,
			Subject:    campaign.Subject,
			Message:    campaign.Message,
		})
		if err != nil {
			if updateErr := s.repository.UpdateCampaignStatus(ctx, campaign.ID, "failed", "", err.Error()); updateErr != nil {
				log.Printf("failed to update reminder campaign %s after provider error: %v", campaign.ID, updateErr)
			}
			log.Printf("failed to send reminder campaign %s: %v", campaign.ID, err)
			return
		}

		if err := s.repository.UpdateCampaignStatus(ctx, campaign.ID, "sent", providerMessageID, ""); err != nil {
			log.Printf("failed to mark reminder campaign %s as sent: %v", campaign.ID, err)
		}
	}()
}

func validateRequest(request SendRequest) (SendRequest, []ValidationIssue) {
	issues := make([]ValidationIssue, 0)

	fromEmail, ok := normalizeEmail(request.FromEmail)
	if !ok {
		issues = append(issues, ValidationIssue{Field: "from_email", Message: "from_email must be a valid email"})
	}

	if len(request.Recipients) == 0 {
		issues = append(issues, ValidationIssue{Field: "recipients", Message: "recipients must contain at least one email"})
	}
	if len(request.Recipients) > maxRecipients {
		issues = append(issues, ValidationIssue{Field: "recipients", Message: "recipients must contain at most 200 emails"})
	}

	recipients := make([]string, 0, len(request.Recipients))
	seen := make(map[string]struct{}, len(request.Recipients))
	for index, recipient := range request.Recipients {
		email, ok := normalizeEmail(recipient)
		if !ok {
			issues = append(issues, ValidationIssue{Field: "recipients", Message: "recipients contains an invalid email at index " + strconv.Itoa(index)})
			continue
		}
		if _, exists := seen[email]; exists {
			continue
		}
		seen[email] = struct{}{}
		recipients = append(recipients, email)
	}

	subject := strings.TrimSpace(request.Subject)
	if subject == "" {
		issues = append(issues, ValidationIssue{Field: "subject", Message: "subject is required"})
	}

	message := strings.TrimSpace(request.Message)
	if message == "" {
		issues = append(issues, ValidationIssue{Field: "message", Message: "message is required"})
	}

	return SendRequest{
		FromEmail:  fromEmail,
		Recipients: recipients,
		Subject:    subject,
		Message:    message,
	}, issues
}

func normalizeEmail(value string) (string, bool) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "", false
	}

	address, err := mail.ParseAddress(value)
	if err != nil || address.Address != value {
		return "", false
	}

	return value, true
}
