package reminders

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	stdmail "net/mail"
	"net/smtp"
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

type SenderConfig struct {
	APIKey       string
	FromName     string
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPKey      string
}

func NewSender(config SenderConfig) EmailSender {
	if strings.TrimSpace(config.SMTPKey) != "" || strings.HasPrefix(strings.TrimSpace(config.APIKey), "xsmtpsib-") {
		smtpKey := strings.TrimSpace(config.SMTPKey)
		if smtpKey == "" {
			smtpKey = strings.TrimSpace(config.APIKey)
		}

		return NewBrevoSMTPSender(config.SMTPHost, config.SMTPPort, config.SMTPUsername, smtpKey, config.FromName)
	}

	return NewBrevoSender(config.APIKey, config.FromName)
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

type BrevoSMTPSender struct {
	host     string
	port     string
	username string
	key      string
	fromName string
}

func NewBrevoSMTPSender(host string, port string, username string, key string, fromName string) *BrevoSMTPSender {
	host = strings.TrimSpace(host)
	if host == "" {
		host = "smtp-relay.brevo.com"
	}
	port = strings.TrimSpace(port)
	if port == "" {
		port = "587"
	}
	fromName = strings.TrimSpace(fromName)
	if fromName == "" {
		fromName = "Invitely"
	}

	return &BrevoSMTPSender{
		host:     host,
		port:     port,
		username: strings.TrimSpace(username),
		key:      strings.TrimSpace(key),
		fromName: fromName,
	}
}

func (s *BrevoSMTPSender) SendReminder(ctx context.Context, message EmailMessage) (string, error) {
	if s.username == "" || s.key == "" {
		return "", ErrEmailProviderUnavailable
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	fromName := valueOrFallback(message.FromName, s.fromName)
	fromHeader := formatAddress(fromName, message.FromEmail)
	toHeader := strings.Join(message.Recipients, ", ")
	raw := strings.Join([]string{
		"From: " + fromHeader,
		"To: " + toHeader,
		"Subject: " + encodeHeader(message.Subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		htmlEmail(message.Message),
	}, "\r\n")

	addr := s.host + ":" + s.port
	auth := smtp.PlainAuth("", s.username, s.key, s.host)
	if err := smtp.SendMail(addr, auth, message.FromEmail, message.Recipients, []byte(raw)); err != nil {
		return "", ProviderError{StatusCode: 0, Body: err.Error()}
	}

	return "", nil
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

func formatAddress(name string, email string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "<" + email + ">"
	}
	return (&stdmail.Address{Name: name, Address: email}).String()
}

func encodeHeader(value string) string {
	value = strings.ReplaceAll(strings.ReplaceAll(value, "\r", ""), "\n", "")
	return mime.QEncoding.Encode("UTF-8", value)
}
