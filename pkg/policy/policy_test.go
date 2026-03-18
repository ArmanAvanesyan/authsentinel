package policy

import (
	"context"
	"os"
	"path/filepath"
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

func TestPolicyEvaluate_WithBundle(t *stdtesting.T) {
	t.Parallel()

	dir := t.TempDir()
	p := filepath.Join(dir, "policy.rego")

	// deny-by-default based on path: allow only /public.
	regoSrc := `
package authsentinel

decision := {"allow": true, "status_code": 200, "reason": "", "headers": {}, "obligations": {}} {
  input.Path == "/public"
} else := {"allow": false, "status_code": 403, "reason": "denied by policy", "headers": {}, "obligations": {}} {
  true
}
`
	if err := os.WriteFile(p, []byte(regoSrc), 0o600); err != nil {
		t.Fatalf("write rego: %v", err)
	}

	eng := NewRegoEngine(DefaultFallbackDeny)
	if err := eng.Load(p); err != nil {
		t.Fatalf("Load: %v", err)
	}

	dec1, err := eng.Evaluate(context.Background(), Input{Protocol: "http", Method: "GET", Path: "/public"})
	if err != nil {
		t.Fatalf("Evaluate /public: %v", err)
	}
	if !dec1.Allow || dec1.StatusCode != 200 {
		t.Fatalf("expected allow 200 for /public, got Allow=%v StatusCode=%d Reason=%q", dec1.Allow, dec1.StatusCode, dec1.Reason)
	}

	dec2, err := eng.Evaluate(context.Background(), Input{Protocol: "http", Method: "GET", Path: "/admin"})
	if err != nil {
		t.Fatalf("Evaluate /admin: %v", err)
	}
	if dec2.Allow || dec2.StatusCode != 403 {
		t.Fatalf("expected deny 403 for /admin, got Allow=%v StatusCode=%d Reason=%q", dec2.Allow, dec2.StatusCode, dec2.Reason)
	}
}
