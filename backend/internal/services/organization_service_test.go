package services

import (
	"strings"
	"testing"
)

func TestToSlug(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"My Org Name", "my-org-name"},
		{"CTF_Event 2025", "ctf-event-2025"},
		{"hello world", "hello-world"},
		{"  leading trailing  ", "leading-trailing"},
		{"UPPERCASE", "uppercase"},
		{"dots.and.dashes", "dots-and-dashes"},
	}
	for _, tc := range cases {
		got := toSlug(tc.input)
		if got != tc.expected {
			t.Errorf("toSlug(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestGenerateSecureToken(t *testing.T) {
	tok, err := generateSecureToken(16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 16 bytes → 32 hex chars
	if len(tok) != 32 {
		t.Errorf("expected 32 hex chars, got %d", len(tok))
	}

	// Two tokens should not be equal
	tok2, _ := generateSecureToken(16)
	if tok == tok2 {
		t.Error("two generated tokens should not be equal")
	}
}

func TestAPIKeyFormat(t *testing.T) {
	// Simulate the key generation that CreateOrganization uses
	rawToken, err := generateSecureToken(16)
	if err != nil {
		t.Fatal(err)
	}
	apiKey := "ra_org_" + rawToken
	prefix := apiKey[:12]

	if !strings.HasPrefix(apiKey, "ra_org_") {
		t.Errorf("API key should start with 'ra_org_', got %q", apiKey)
	}
	if len(prefix) != 12 {
		t.Errorf("API key prefix should be 12 chars, got %d", len(prefix))
	}
}

func TestEventTokenFormat(t *testing.T) {
	rawToken, err := generateSecureToken(16)
	if err != nil {
		t.Fatal(err)
	}
	eventToken := "evt_" + rawToken
	prefix := eventToken[:8]

	if !strings.HasPrefix(eventToken, "evt_") {
		t.Errorf("event token should start with 'evt_', got %q", eventToken)
	}
	if len(prefix) != 8 {
		t.Errorf("event token prefix should be 8 chars, got %d", len(prefix))
	}
}
