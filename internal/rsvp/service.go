package rsvp

import (
	"context"
	"errors"
	"html"
	"net"
	"net/mail"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"invitely-api/pkg/uuid"
)

var (
	ErrPublicRSVPClosed = errors.New("event is not available for RSVP")
	ErrTooManyCompanions = errors.New("companions exceeds guest limit")
	ErrRateLimited = errors.New("too many RSVP attempts")
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Submit(ctx context.Context, request SubmitRequest) (RSVP, error) {
	status := strings.ToLower(strings.TrimSpace(request.Status))
	if status == "confirmed" {
		status = "accepted"
	}
	if status != "accepted" && status != "declined" && status != "pending" && status != "invited" {
		return RSVP{}, errors.New("status must be accepted, declined, pending or invited")
	}
	if strings.TrimSpace(request.GuestID) == "" {
		return RSVP{}, errors.New("guest_id is required")
	}
	if strings.TrimSpace(request.EventID) == "" {
		return RSVP{}, errors.New("event_id is required")
	}

	id, err := uuid.New()
	if err != nil {
		return RSVP{}, err
	}

	return s.repository.Upsert(ctx, RSVP{
		ID:        id,
		GuestID:   strings.TrimSpace(request.GuestID),
		EventID:   strings.TrimSpace(request.EventID),
		Status:    status,
		Source:    "direct",
	})
}

func (s *Service) FindByGuest(ctx context.Context, guestID string) (RSVP, error) {
	return s.repository.FindByGuest(ctx, guestID)
}

func (s *Service) SubmitPublic(ctx context.Context, slug string, ip string, request PublicSubmitRequest) (RSVP, error) {
	request.Name = sanitizeText(request.Name, 160)
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Status = strings.ToLower(strings.TrimSpace(request.Status))
	request.Message = sanitizeText(request.Message, 1000)
	request.InviteToken = strings.TrimSpace(request.InviteToken)

	if strings.TrimSpace(slug) == "" {
		return RSVP{}, errors.New("event slug is required")
	}
	if request.Name == "" {
		return RSVP{}, errors.New("name is required")
	}
	if !validEmail(request.Email) {
		return RSVP{}, errors.New("valid email is required")
	}
	if request.Status != "accepted" && request.Status != "declined" {
		return RSVP{}, errors.New("status must be accepted or declined")
	}
	if request.Companions < 0 {
		return RSVP{}, errors.New("companions must be greater than or equal to zero")
	}

	return s.repository.SubmitPublic(ctx, strings.TrimSpace(slug), ip, request)
}

func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}

var whitespace = regexp.MustCompile(`\s+`)

func sanitizeText(value string, max int) string {
	value = strings.TrimSpace(value)
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			continue
		}
		builder.WriteRune(r)
	}
	clean := whitespace.ReplaceAllString(builder.String(), " ")
	if len(clean) > max {
		clean = clean[:max]
	}
	return html.EscapeString(strings.TrimSpace(clean))
}

func clientIPKey(ip string) string {
	ip = strings.TrimSpace(ip)
	if parsed := net.ParseIP(ip); parsed != nil {
		return parsed.String()
	}
	return ip
}

type rateLimiter struct {
	max      int
	window   time.Duration
	mu       sync.Mutex
	attempts map[string][]time.Time
}

var publicRSVPLimiter = newRateLimiter(5, 10*time.Minute)

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{max: max, window: window, attempts: make(map[string][]time.Time)}
}

func (l *rateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)
	kept := l.attempts[key][:0]
	for _, attempt := range l.attempts[key] {
		if attempt.After(cutoff) {
			kept = append(kept, attempt)
		}
	}
	if len(kept) >= l.max {
		l.attempts[key] = kept
		return false
	}
	l.attempts[key] = append(kept, now)
	return true
}
