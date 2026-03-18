package httpserver

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/authsentinel/internal/agent/config"
	"github.com/ArmanAvanesyan/authsentinel/pkg/agent"
)

type mockAgentService struct{}

func (m *mockAgentService) Session(_ context.Context, _ agent.SessionRequest) (*agent.SessionResponse, error) {
	return &agent.SessionResponse{}, nil
}
func (m *mockAgentService) LoginStart(_ context.Context, _ agent.LoginStartRequest) (*agent.LoginStartResponse, error) {
	return nil, nil
}
func (m *mockAgentService) LoginEnd(_ context.Context, _ agent.LoginEndRequest) (*agent.LoginEndResponse, error) {
	return nil, nil
}
func (m *mockAgentService) Refresh(_ context.Context, _ agent.RefreshRequest) (*agent.RefreshResponse, error) {
	return nil, nil
}
func (m *mockAgentService) Logout(_ context.Context, _ agent.LogoutRequest) (*agent.LogoutResponse, error) {
	return nil, nil
}

func TestNewReturnsServerWithHandler(t *testing.T) {
	cfg := &config.Config{HTTPPort: "8080", CookieName: "test", SessionTTLSeconds: 3600}
	s := New(&mockAgentService{}, cfg, nil)
	if s == nil {
		t.Fatalf("expected non-nil Server")
	}
	if s.Handler() == nil {
		t.Fatalf("expected non-nil Handler")
	}
}
