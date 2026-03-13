package plugins

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/authsentinel/pkg/plugins/caddy"
	"github.com/ArmanAvanesyan/authsentinel/pkg/plugins/krakend"
	"github.com/ArmanAvanesyan/authsentinel/pkg/plugins/traefik"
	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
)

type testEngine struct{}

func (testEngine) Handle(_ context.Context, _ *proxy.Request) (*proxy.Response, error) {
	return &proxy.Response{Allow: true}, nil
}

func TestKrakendNewPluginStoresEngine(t *testing.T) {
	e := testEngine{}
	p := krakend.NewPlugin(e)
	if p == nil {
		t.Fatal("expected non-nil Plugin")
	}
}

func TestTraefikNewMiddlewareStoresEngine(t *testing.T) {
	e := testEngine{}
	m := traefik.NewMiddleware(e)
	if m == nil {
		t.Fatal("expected non-nil Middleware")
	}
}

func TestCaddyNewMiddlewareStoresEngine(t *testing.T) {
	e := testEngine{}
	m := caddy.NewMiddleware(e)
	if m == nil {
		t.Fatal("expected non-nil Middleware")
	}
}
