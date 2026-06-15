package reminders

import "testing"

func TestValidateRequest(t *testing.T) {
	request := SendRequest{
		FromEmail:  "Organizador@Example.com",
		Recipients: []string{"Convidado1@Example.com", "convidado1@example.com", "convidado2@example.com"},
		Subject:    "Lembrete",
		Message:    "Confirme sua presenca.",
	}

	normalized, issues := validateRequest(request)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}
	if normalized.FromEmail != "organizador@example.com" {
		t.Fatalf("expected normalized from_email, got %q", normalized.FromEmail)
	}
	if len(normalized.Recipients) != 2 {
		t.Fatalf("expected duplicate recipients to be removed, got %d", len(normalized.Recipients))
	}
}

func TestValidateRequestWithInvalidFields(t *testing.T) {
	_, issues := validateRequest(SendRequest{
		FromEmail:  "invalid",
		Recipients: []string{"also-invalid"},
	})

	if len(issues) != 4 {
		t.Fatalf("expected 4 validation issues, got %d: %v", len(issues), issues)
	}
}

func TestValidateRequestRecipientLimit(t *testing.T) {
	recipients := make([]string, 201)
	for i := range recipients {
		recipients[i] = "convidado" + string(rune('a'+i%26)) + "@example.com"
	}

	_, issues := validateRequest(SendRequest{
		FromEmail:  "organizador@example.com",
		Recipients: recipients,
		Subject:    "Lembrete",
		Message:    "Mensagem",
	})

	if len(issues) == 0 {
		t.Fatal("expected recipient limit issue")
	}
}
