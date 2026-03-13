package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testEngine struct{}

func (testEngine) Handle(_ context.Context, _ *Request) (*Response, error) {
	return &Response{Allow: true}, nil
}

func TestNormalizeRequestPopulatesFields(t *testing.T) {
	req := NormalizeRequest(
		"http",
		"GET",
		"/foo",
		map[string]string{"X-Test": "1"},
		map[string]string{"cookie": "value"},
		[]byte("body"),
	)

	if req.Protocol != "http" || req.Method != "GET" || req.Path != "/foo" {
		t.Fatalf("unexpected basic fields: %#v", req)
	}
	if req.Headers["X-Test"] != "1" {
		t.Fatalf("expected header X-Test=1, got %q", req.Headers["X-Test"])
	}
	if req.Cookies["cookie"] != "value" {
		t.Fatalf("expected cookie value, got %q", req.Cookies["cookie"])
	}
	if string(req.Body) != "body" {
		t.Fatalf("expected body 'body', got %q", string(req.Body))
	}
}

func TestMiddlewareCallsNextHandler(t *testing.T) {
	engine := testEngine{}
	mw := Middleware(engine)

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})

	handler := mw(next)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/foo", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatalf("expected next handler to be called")
	}
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}
