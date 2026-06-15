package apidocs

import (
	"encoding/json"
	"testing"
)

func TestSwaggerJSON(t *testing.T) {
	document, err := SwaggerJSON()
	if err != nil {
		t.Fatalf("SwaggerJSON() returned error: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(document), &parsed); err != nil {
		t.Fatalf("SwaggerJSON() returned invalid JSON: %v", err)
	}

	if parsed["openapi"] != "3.0.3" {
		t.Fatalf("unexpected openapi version: %v", parsed["openapi"])
	}
}
