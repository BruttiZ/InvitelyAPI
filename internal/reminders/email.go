package reminders

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrEmailProviderUnavailable = errors.New("email provider unavailable")

type ProviderError struct {
	StatusCode int
	Body       string
}

func (e ProviderError) Error() string {
	return fmt.Sprintf("email provider returned status %d", e.StatusCode)
}

type EmailMessage struct {
	FromEmail  string
	FromName   string
	Recipients []string
	Subject    string
	Message    string
}

type EmailSender interface {
	SendReminder(ctx context.Context, message EmailMessage) (string, error)
}

type BrevoSender struct {
	apiKey     string
	fromName   string
	httpClient *http.Client
}

func NewBrevoSender(apiKey string, fromName string) *BrevoSender {
	fromName = strings.TrimSpace(fromName)
	if fromName == "" {
		fromName = "Invitely"
	}

	return &BrevoSender{
		apiKey:   strings.TrimSpace(apiKey),
		fromName: fromName,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *BrevoSender) SendReminder(ctx context.Context, message EmailMessage) (string, error) {
	if s.apiKey == "" {
		return "", ErrEmailProviderUnavailable
	}

	payload := brevoEmailRequest{
		Sender: brevoEmailContact{
			Name:  valueOrFallback(message.FromName, s.fromName),
			Email: message.FromEmail,
		},
		To:          brevoRecipients(message.Recipients),
		Subject:     message.Subject,
		TextContent: message.Message,
		HTMLContent: htmlEmail(message.Message),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.brevo.com/v3/smtp/email", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", s.apiKey)
	req.Header.Set("content-type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", ErrEmailProviderUnavailable
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", ProviderError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	var result brevoEmailResponse
	if len(respBody) > 0 {
		_ = json.Unmarshal(respBody, &result)
	}

	return result.MessageID, nil
}

type brevoEmailRequest struct {
	Sender      brevoEmailContact   `json:"sender"`
	To          []brevoEmailContact `json:"to"`
	Subject     string              `json:"subject"`
	TextContent string              `json:"textContent"`
	HTMLContent string              `json:"htmlContent"`
}

type brevoEmailContact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type brevoEmailResponse struct {
	MessageID string `json:"messageId"`
}

func brevoRecipients(recipients []string) []brevoEmailContact {
	contacts := make([]brevoEmailContact, 0, len(recipients))
	for _, recipient := range recipients {
		contacts = append(contacts, brevoEmailContact{Email: recipient})
	}
	return contacts
}

func htmlEmail(message string) string {
	escaped := html.EscapeString(message)
	escaped = strings.ReplaceAll(escaped, "\r\n", "\n")
	escaped = strings.ReplaceAll(escaped, "\n", "<br>")
	return "<!doctype html><html><body><p>" + escaped + "</p></body></html>"
}

func valueOrFallback(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}
