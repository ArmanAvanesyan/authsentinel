package policy

import (
	"context"
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

func TestDecisionObligations(t *stdtesting.T) {
	dec := &Decision{
		Allow:       true,
		StatusCode:  200,
		Obligations: map[string]any{"set_header_X_User": "alice"},
	}
	if dec.Obligations == nil || dec.Obligations["set_header_X_User"] != "alice" {
		t.Fatalf("expected obligations set_header_X_User=alice, got %#v", dec.Obligations)
	}
}

func TestWASMRuntimeFallbackNoModule(t *stdtesting.T) {
	// No module loaded -> fallback deny
	w := NewWASMRuntime(DefaultFallbackDeny)
	dec, err := w.Evaluate(context.Background(), Input{})
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if dec.Allow || dec.StatusCode != 503 {
		t.Fatalf("expected fallback deny 503, got Allow=%v StatusCode=%d", dec.Allow, dec.StatusCode)
	}
	// Fallback allow
	w2 := NewWASMRuntime(DefaultFallbackAllow)
	dec2, _ := w2.Evaluate(context.Background(), Input{})
	if !dec2.Allow || dec2.StatusCode != 200 {
		t.Fatalf("expected fallback allow 200, got Allow=%v StatusCode=%d", dec2.Allow, dec2.StatusCode)
	}
}

// TestPolicyEvaluate_WithBundle is a placeholder for when Rego/WASM bundle loading is implemented.
// Intended behavior: load a minimal bundle (e.g. allow GET /public, deny otherwise), call Evaluate
// with Input{Method: "GET", Path: "/public"} and assert Allow=true, StatusCode=200; with Path: "/admin"
// assert Allow=false and appropriate StatusCode/Reason. Remove the skip and implement when bundle loader exists.
func TestPolicyEvaluate_WithBundle(t *stdtesting.T) {
	t.Skip("policy bundle loading not yet implemented; when Rego/WASM bundle loader exists, load minimal bundle and assert Evaluate returns expected Allow/StatusCode/Headers/Obligations for given Input")
}
