package token

import (
	"testing"
	"time"
)

func TestPrincipalFields(t *testing.T) {
	now := time.Now()
	p := &Principal{
		Subject:   "user-123",
		Scopes:    []string{"read", "write"},
		Roles:     []string{"admin"},
		Claims:    map[string]any{"foo": "bar"},
		ExpiresAt: now,
	}

	if p.Subject != "user-123" {
		t.Errorf("expected Subject %q, got %q", "user-123", p.Subject)
	}
	if len(p.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(p.Scopes))
	}
	if len(p.Roles) != 1 || p.Roles[0] != "admin" {
		t.Errorf("unexpected roles: %#v", p.Roles)
	}
	if got := p.Claims["foo"]; got != "bar" {
		t.Errorf("expected claim foo=bar, got %#v", got)
	}
	if !p.ExpiresAt.Equal(now) {
		t.Errorf("expected ExpiresAt %v, got %v", now, p.ExpiresAt)
	}
}
