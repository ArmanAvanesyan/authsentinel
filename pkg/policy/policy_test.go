package policy

import (
	stdtesting "testing"

	asTesting "github.com/ArmanAvanesyan/authsentinel/pkg/testing"
)

func TestInputWithPrincipalAndHeaders(t *stdtesting.T) {
	principal := asTesting.NewTestPrincipal("user-123")

	in := Input{
		Protocol:         "http",
		Method:           "GET",
		Path:             "/foo",
		GraphQLOperation: "QueryFoo",
		GRPCService:      "svc.Foo",
		GRPCMethod:       "Bar",
		Principal:        principal,
		Headers: map[string]string{
			"X-Test": "1",
		},
	}

	if in.Principal == nil || in.Principal.Subject != "user-123" {
		t.Fatalf("expected principal with subject user-123, got %#v", in.Principal)
	}
	if got := in.Headers["X-Test"]; got != "1" {
		t.Fatalf("expected header X-Test=1, got %q", got)
	}
}

func TestDecisionFields(t *stdtesting.T) {
	dec := &Decision{
		Allow:      true,
		StatusCode: 200,
		Headers: map[string]string{
			"X-Policy": "ok",
		},
		Reason: "allowed",
	}

	if !dec.Allow || dec.StatusCode != 200 {
		t.Fatalf("unexpected decision core fields: %#v", dec)
	}
	if got := dec.Headers["X-Policy"]; got != "ok" {
		t.Fatalf("expected header X-Policy=ok, got %q", got)
	}
	if dec.Reason == "" {
		t.Fatalf("expected non-empty reason")
	}
}
