package grpc

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
	gogrpc "google.golang.org/grpc"
)

type testEngine struct{}

func (testEngine) Handle(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
	return &proxy.Response{Allow: true}, nil
}

func TestToProxyRequestUsesHeadersAndBody(t *testing.T) {
	headers := map[string]string{"X-Test": "1"}
	body := []byte("payload")

	req := ToProxyRequest("/svc.Foo/Bar", headers, body)
	if req == nil {
		t.Fatal("expected non-nil proxy.Request")
	}
	if req.Headers["X-Test"] != "1" {
		t.Fatalf("expected header X-Test=1, got %q", req.Headers["X-Test"])
	}
	if string(req.Body) != "payload" {
		t.Fatalf("expected body payload, got %q", string(req.Body))
	}
}

func TestUnaryServerInterceptorCallsHandler(t *testing.T) {
	engine := testEngine{}
	interceptor := UnaryServerInterceptor(engine)

	called := false
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		return "ok", nil
	}

	info := &gogrpc.UnaryServerInfo{FullMethod: "/svc.Foo/Bar"}
	res, err := interceptor(context.Background(), "in", info, handler)
	if err != nil {
		t.Fatalf("unexpected error from interceptor: %v", err)
	}
	if res != "ok" || !called {
		t.Fatalf("expected handler to be called and return ok, got res=%v called=%v", res, called)
	}
}

func TestStreamServerInterceptorCallsHandler(t *testing.T) {
	engine := testEngine{}
	interceptor := StreamServerInterceptor(engine)

	called := false
	handler := func(srv interface{}, ss gogrpc.ServerStream) error {
		called = true
		return nil
	}

	info := &gogrpc.StreamServerInfo{FullMethod: "/svc.Foo/Bar"}
	err := interceptor(nil, nil, info, handler)
	if err != nil {
		t.Fatalf("unexpected error from stream interceptor: %v", err)
	}
	if !called {
		t.Fatalf("expected stream handler to be called")
	}
}
