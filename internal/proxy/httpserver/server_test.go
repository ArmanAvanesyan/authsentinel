package httpserver

import (
	"testing"

	"github.com/ArmanAvanesyan/authsentinel/internal/proxy"
	"github.com/ArmanAvanesyan/authsentinel/internal/proxy/config"
	"github.com/ArmanAvanesyan/authsentinel/pkg/policy"
)

func TestNewReturnsServerWithHandler(t *testing.T) {
	cfg := &config.Config{
		UpstreamURL:     "http://localhost:8002",
		ProxyPathPrefix: "/graphql",
		AgentURL:        "http://localhost:8080",
		CookieName:      "test",
	}
	client := proxy.NewAgentClient(cfg.AgentURL, cfg.CookieName)
	s := New(cfg, client, policy.NewWASMRuntime(policy.DefaultFallbackAllow), nil, nil, nil, nil, nil)
	if s == nil {
		t.Fatalf("expected non-nil Server")
	}
	if s.Handler() == nil {
		t.Fatalf("expected non-nil Handler")
	}
}
