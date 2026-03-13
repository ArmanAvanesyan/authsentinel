package sdk

import (
	"context"
	"net/http"
	"testing"
)

func TestGraphQLContextKeyValue(t *testing.T) {
	if GraphQLContextKey == "" {
		t.Fatalf("expected non-empty GraphQLContextKey")
	}
}

func TestPrincipalAliasAssignable(t *testing.T) {
	p := Principal{
		Subject: "user-123",
	}
	if p.Subject != "user-123" {
		t.Fatalf("expected subject user-123, got %q", p.Subject)
	}

	base := &p
	if base.Subject != "user-123" {
		t.Fatalf("alias did not behave as token.Principal")
	}
}

func TestIdentityFromGRPCContextDefault(t *testing.T) {
	got, err := IdentityFromGRPCContext(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil principal from default implementation, got %#v", got)
	}
}

func TestIdentityFromHTTPRequestDefault(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	got, err := IdentityFromHTTPRequest(req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil principal from default implementation, got %#v", got)
	}
}
