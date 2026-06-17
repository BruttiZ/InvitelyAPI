package rsvp

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestPublicRSVPByInviteToken(t *testing.T) {
	repo := newFakePublicRepository()
	event := repo.addEvent("casamento", "published")
	guest := repo.addGuest(event.id, "Ana", "ana@example.com", "token-ana", 2)

	service := NewService(repo)
	response, err := service.SubmitPublic(context.Background(), "casamento", "127.0.0.1", PublicSubmitRequest{
		Name:        "Ana Maria",
		Email:       "ANA@example.com",
		Status:      "accepted",
		Companions:  1,
		Message:     "Confirmo!",
		InviteToken: "token-ana",
	})
	if err != nil {
		t.Fatalf("SubmitPublic returned error: %v", err)
	}

	if response.GuestID != guest.id || response.Status != "accepted" || response.Companions != 1 || response.Source != "public_link" {
		t.Fatalf("unexpected response: %#v", response)
	}
	updated := repo.guests[guest.id]
	if updated.name != "Ana Maria" || updated.email != "ana@example.com" || updated.partySize != 2 || updated.status != "accepted" {
		t.Fatalf("guest was not updated from invite token RSVP: %#v", updated)
	}
}

func TestPublicRSVPByEmailCreatesGuest(t *testing.T) {
	repo := newFakePublicRepository()
	event := repo.addEvent("publico", "active")
	service := NewService(repo)

	response, err := service.SubmitPublic(context.Background(), "publico", "127.0.0.2", PublicSubmitRequest{
		Name:       "Bruno",
		Email:      "BRUNO@example.com",
		Status:     "accepted",
		Companions: 0,
	})
	if err != nil {
		t.Fatalf("SubmitPublic returned error: %v", err)
	}

	guest := repo.guests[response.GuestID]
	if guest.eventID != event.id || guest.email != "bruno@example.com" || guest.maxCompanions != 5 || guest.metadataSource != "public_event_link" {
		t.Fatalf("guest was not created with public defaults: %#v", guest)
	}
}

func TestPublicRSVPSecondResponseUpdatesWithoutDuplicate(t *testing.T) {
	repo := newFakePublicRepository()
	repo.addEvent("repetido", "published")
	service := NewService(repo)

	first, err := service.SubmitPublic(context.Background(), "repetido", "127.0.0.3", PublicSubmitRequest{
		Name: "Carla", Email: "carla@example.com", Status: "accepted", Companions: 1,
	})
	if err != nil {
		t.Fatalf("first SubmitPublic returned error: %v", err)
	}
	second, err := service.SubmitPublic(context.Background(), "repetido", "127.0.0.3", PublicSubmitRequest{
		Name: "Carla Souza", Email: "CARLA@example.com", Status: "declined", Companions: 0,
	})
	if err != nil {
		t.Fatalf("second SubmitPublic returned error: %v", err)
	}

	if first.GuestID != second.GuestID || len(repo.guests) != 1 || len(repo.rsvps) != 1 {
		t.Fatalf("expected one guest and one RSVP, got guests=%d rsvps=%d", len(repo.guests), len(repo.rsvps))
	}
	if repo.guests[first.GuestID].status != "declined" || repo.guests[first.GuestID].partySize != 1 {
		t.Fatalf("guest was not updated by second RSVP: %#v", repo.guests[first.GuestID])
	}
}

func TestPublicRSVPDeclined(t *testing.T) {
	repo := newFakePublicRepository()
	repo.addEvent("recusa", "published")
	service := NewService(repo)

	response, err := service.SubmitPublic(context.Background(), "recusa", "127.0.0.4", PublicSubmitRequest{
		Name: "Diego", Email: "diego@example.com", Status: "declined",
	})
	if err != nil {
		t.Fatalf("SubmitPublic returned error: %v", err)
	}
	if response.Status != "declined" || repo.guests[response.GuestID].status != "declined" {
		t.Fatalf("declined RSVP was not persisted: response=%#v guest=%#v", response, repo.guests[response.GuestID])
	}
}

func TestPublicRSVPBlocksCompanionsAboveLimit(t *testing.T) {
	repo := newFakePublicRepository()
	event := repo.addEvent("limite", "published")
	repo.addGuest(event.id, "Eva", "eva@example.com", "token-eva", 1)
	service := NewService(repo)

	_, err := service.SubmitPublic(context.Background(), "limite", "127.0.0.5", PublicSubmitRequest{
		Name: "Eva", Email: "eva@example.com", Status: "accepted", Companions: 2,
	})
	if !errors.Is(err, ErrTooManyCompanions) {
		t.Fatalf("expected ErrTooManyCompanions, got %v", err)
	}
}

func TestPublicRSVPBlocksMissingOrUnpublishedEvent(t *testing.T) {
	repo := newFakePublicRepository()
	repo.addEvent("rascunho", "draft")
	service := NewService(repo)

	_, err := service.SubmitPublic(context.Background(), "inexistente", "127.0.0.6", PublicSubmitRequest{
		Name: "Fabio", Email: "fabio@example.com", Status: "accepted",
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

	_, err = service.SubmitPublic(context.Background(), "rascunho", "127.0.0.7", PublicSubmitRequest{
		Name: "Fabio", Email: "fabio@example.com", Status: "accepted",
	})
	if !errors.Is(err, ErrPublicRSVPClosed) {
		t.Fatalf("expected ErrPublicRSVPClosed, got %v", err)
	}
}

func TestPublicRSVPMetricsReflectResponses(t *testing.T) {
	repo := newFakePublicRepository()
	repo.addEvent("metricas", "published")
	service := NewService(repo)

	requests := []PublicSubmitRequest{
		{Name: "Gabi", Email: "gabi@example.com", Status: "accepted"},
		{Name: "Hugo", Email: "hugo@example.com", Status: "declined"},
	}
	for i, request := range requests {
		if _, err := service.SubmitPublic(context.Background(), "metricas", "10.0.0."+string(rune('1'+i)), request); err != nil {
			t.Fatalf("SubmitPublic returned error: %v", err)
		}
	}
	eventID := repo.eventsBySlug["metricas"].id
	repo.addGuest(eventID, "Iris", "iris@example.com", "token-iris", 5)

	accepted, declined, invited := repo.metrics(eventID)
	if accepted != 1 || declined != 1 || invited != 1 {
		t.Fatalf("unexpected metrics accepted=%d declined=%d invited=%d", accepted, declined, invited)
	}
}

type fakePublicRepository struct {
	eventsBySlug map[string]fakeEvent
	guests       map[string]fakeGuest
	guestByEmail map[string]string
	guestByToken map[string]string
	rsvps        map[string]RSVP
	nextID       int
}

type fakeEvent struct {
	id     string
	status string
	endsAt *time.Time
}

type fakeGuest struct {
	id             string
	eventID        string
	name           string
	email          string
	status         string
	partySize      int
	maxCompanions  int
	inviteToken    string
	metadataSource string
}

func newFakePublicRepository() *fakePublicRepository {
	return &fakePublicRepository{
		eventsBySlug: make(map[string]fakeEvent),
		guests:       make(map[string]fakeGuest),
		guestByEmail: make(map[string]string),
		guestByToken: make(map[string]string),
		rsvps:        make(map[string]RSVP),
	}
}

func (r *fakePublicRepository) Upsert(ctx context.Context, response RSVP) (RSVP, error) {
	r.rsvps[response.GuestID] = response
	return response, nil
}

func (r *fakePublicRepository) FindByGuest(ctx context.Context, guestID string) (RSVP, error) {
	response, ok := r.rsvps[guestID]
	if !ok {
		return RSVP{}, sql.ErrNoRows
	}
	return response, nil
}

func (r *fakePublicRepository) SubmitPublic(ctx context.Context, slug string, ip string, request PublicSubmitRequest) (RSVP, error) {
	event, ok := r.eventsBySlug[slug]
	if !ok {
		return RSVP{}, sql.ErrNoRows
	}
	if event.status != "published" && event.status != "active" {
		return RSVP{}, ErrPublicRSVPClosed
	}
	if event.endsAt != nil && event.endsAt.Before(time.Now()) {
		return RSVP{}, ErrPublicRSVPClosed
	}

	guestID := ""
	if request.InviteToken != "" {
		guestID = r.guestByToken[event.id+"|"+request.InviteToken]
		if guestID == "" {
			return RSVP{}, sql.ErrNoRows
		}
	} else {
		guestID = r.guestByEmail[event.id+"|"+request.Email]
	}

	if guestID == "" {
		guest := r.addGuest(event.id, request.Name, request.Email, "generated-token-"+r.next(), 5)
		guest.metadataSource = "public_event_link"
		r.guests[guest.id] = guest
		guestID = guest.id
	}

	guest := r.guests[guestID]
	if request.Companions > guest.maxCompanions {
		return RSVP{}, ErrTooManyCompanions
	}
	guest.name = request.Name
	guest.email = request.Email
	guest.status = request.Status
	guest.partySize = 1 + request.Companions
	r.guests[guest.id] = guest
	r.guestByEmail[event.id+"|"+request.Email] = guest.id

	response := RSVP{
		ID:         "rsvp-" + guest.id,
		EventID:    event.id,
		GuestID:    guest.id,
		Status:     request.Status,
		Companions: request.Companions,
		Message:    request.Message,
		Source:     "public_link",
	}
	r.rsvps[guest.id] = response
	return response, nil
}

func (r *fakePublicRepository) addEvent(slug string, status string) fakeEvent {
	event := fakeEvent{id: "event-" + r.next(), status: status}
	r.eventsBySlug[slug] = event
	return event
}

func (r *fakePublicRepository) addGuest(eventID string, name string, email string, token string, maxCompanions int) fakeGuest {
	guest := fakeGuest{
		id:            "guest-" + r.next(),
		eventID:       eventID,
		name:          name,
		email:         email,
		status:        "invited",
		partySize:     1,
		maxCompanions: maxCompanions,
		inviteToken:   token,
	}
	r.guests[guest.id] = guest
	r.guestByEmail[eventID+"|"+email] = guest.id
	r.guestByToken[eventID+"|"+token] = guest.id
	return guest
}

func (r *fakePublicRepository) metrics(eventID string) (accepted int, declined int, invited int) {
	for _, guest := range r.guests {
		if guest.eventID != eventID {
			continue
		}
		switch guest.status {
		case "accepted", "confirmed":
			accepted++
		case "declined":
			declined++
		case "invited", "pending":
			invited++
		}
	}
	return accepted, declined, invited
}

func (r *fakePublicRepository) next() string {
	r.nextID++
	return string(rune('0' + r.nextID))
}
