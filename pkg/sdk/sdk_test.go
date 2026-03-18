package sdk

import (
	"context"
	"net/http"
	"testing"

	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
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

func TestPrincipalFromContext(t *testing.T) {
	ctx := context.Background()
	if PrincipalFromContext(ctx) != nil {
		t.Fatal("expected nil principal in empty context")
	}
	p := &token.Principal{Subject: "sub-1"}
	ctx = WithPrincipal(ctx, p)
	if got := PrincipalFromContext(ctx); got != p {
		t.Fatalf("expected principal %p, got %p", p, got)
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

func TestIdentityFromHTTPRequestNoMiddleware(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "http://example.com", nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}
	got, err := IdentityFromHTTPRequest(req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil principal when no middleware set context, got %#v", got)
	}
}

func TestMiddlewareOptionalAuth(t *testing.T) {
	extractor := &staticExtractor{nil}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := PrincipalFromContext(r.Context())
		if p != nil {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(p.Subject))
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	})
	handler := Middleware(extractor, false)(next)
	req := mustNewRequest(http.MethodGet, "http://example.com", nil)
	rec := &responseRecorder{code: 200}
	handler.ServeHTTP(rec, req)
	if rec.code != http.StatusNoContent {
		t.Fatalf("expected 204 when extractor returns nil, got %d", rec.code)
	}
}

func TestMiddlewareWithPrincipal(t *testing.T) {
	p := &token.Principal{Subject: "alice"}
	extractor := &staticExtractor{p}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := PrincipalFromContext(r.Context())
		if got != p {
			t.Fatalf("expected principal in context, got %v", got)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(got.Subject))
	})
	handler := Middleware(extractor, false)(next)
	req := mustNewRequest(http.MethodGet, "http://example.com", nil)
	rec := &responseRecorder{code: 200}
	handler.ServeHTTP(rec, req)
	if rec.code != http.StatusOK || string(rec.body) != "alice" {
		t.Fatalf("expected 200 and body alice, got %d %q", rec.code, string(rec.body))
	}
}

func TestMiddlewareRequireAuthDeny(t *testing.T) {
	extractor := &staticExtractor{nil}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := Middleware(extractor, true)(next)
	req := mustNewRequest(http.MethodGet, "http://example.com", nil)
	rec := &responseRecorder{code: 200}
	handler.ServeHTTP(rec, req)
	if rec.code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when requireAuth and no principal, got %d", rec.code)
	}
}

func TestAgentClientURLs(t *testing.T) {
	c := NewAgentClient("https://auth.example.com", "sess")
	if got := c.GetLoginURL("https://app.example.com/callback"); got != "https://auth.example.com/login?redirect_to=https%3A%2F%2Fapp.example.com%2Fcallback" {
		t.Fatalf("unexpected login URL: %s", got)
	}
	if got := c.GetLogoutURL(""); got != "https://auth.example.com/logout?" {
		t.Fatalf("unexpected logout URL: %s", got)
	}
}

type staticExtractor struct {
	p *token.Principal
}

func (s *staticExtractor) ExtractPrincipal(_ context.Context, _ *http.Request) (*token.Principal, error) {
	return s.p, nil
}

type responseRecorder struct {
	code int
	body []byte
}

func (r *responseRecorder) Header() http.Header       { return http.Header{} }
func (r *responseRecorder) Write(b []byte) (int, error) { r.body = b; return len(b), nil }
func (r *responseRecorder) WriteHeader(code int)      { r.code = code }

func mustNewRequest(method, url string, body []byte) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	return req
}
