// Package integration: proxy integration tests (mock agent, mock upstream).
package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/ArmanAvanesyan/authsentinel/internal/proxy"
	"github.com/ArmanAvanesyan/authsentinel/internal/proxy/config"
	"github.com/ArmanAvanesyan/authsentinel/internal/proxy/httpserver"
)

// mockAgentServer returns an httptest server that responds to GET /internal/resolve with 200 and principal claims.
func mockAgentServer(t *testing.T, cookieName string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/internal/resolve", func(w http.ResponseWriter, r *http.Request) {
		cookieVal := ""
		if c, _ := r.Cookie(cookieName); c != nil {
			cookieVal = c.Value
		}
		if cookieVal == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":   "at",
			"claims":         map[string]any{"sub": "test-user", "email": "u@example.com"},
			"tenant_context": map[string]any{},
		})
	})
	return httptest.NewServer(mux)
}

// mockUpstreamServer returns an httptest server that records the last request for assertion.
func mockUpstreamServer(t *testing.T) (*httptest.Server, func() *http.Request) {
	t.Helper()
	var lastReq *http.Request
	var mu sync.Mutex
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		lastReq = r
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	srv := httptest.NewServer(mux)
	getLast := func() *http.Request {
		mu.Lock()
		defer mu.Unlock()
		return lastReq
	}
	return srv, getLast
}

func TestProxy_AuthenticatedRequest_ForwardsToUpstreamWithHeaders(t *testing.T) {
	agentSrv := mockAgentServer(t, "test_session")
	defer agentSrv.Close()
	upstreamSrv, getLastRequest := mockUpstreamServer(t)
	defer upstreamSrv.Close()

	cfg := &config.Config{
		UpstreamURL:     upstreamSrv.URL,
		ProxyPathPrefix: "/graphql",
		AgentURL:        agentSrv.URL,
		CookieName:      "test_session",
		RequireAuth:     true,
	}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("config: %v", err)
	}
	client := proxy.NewAgentClient(cfg.AgentURL, cfg.CookieName)
	proxySrv := httpserver.New(cfg, client, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/graphql", nil)
	req.Header.Set("Cookie", "test_session=any-session-value")
	rr := httptest.NewRecorder()
	proxySrv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", rr.Code, rr.Body.String())
	}
	last := getLastRequest()
	if last == nil {
		t.Fatal("upstream did not receive request")
	}
	if last.Header.Get("X-User-Id") != "test-user" {
		t.Errorf("expected X-User-Id=test-user, got %q", last.Header.Get("X-User-Id"))
	}
}

// TestProxy_PolicyBundleEnforcement is a placeholder for when policy bundle loading is implemented.
// Intended: load a small Rego/WASM bundle, run proxy engine with it, assert one allow and one deny for known inputs.
func TestProxy_PolicyBundleEnforcement(t *testing.T) {
	t.Skip("policy bundle loading not yet implemented; add integration test when bundle loader exists")
}

func TestProxy_UnauthenticatedRequest_RequireAuth_Returns401(t *testing.T) {
	agentSrv := mockAgentServer(t, "test_session")
	defer agentSrv.Close()
	upstreamSrv, _ := mockUpstreamServer(t)
	defer upstreamSrv.Close()

	cfg := &config.Config{
		UpstreamURL:     upstreamSrv.URL,
		ProxyPathPrefix: "/graphql",
		AgentURL:        agentSrv.URL,
		CookieName:      "test_session",
		RequireAuth:     true,
	}
	cfg.ApplyDefaults()
	_ = cfg.Validate()
	client := proxy.NewAgentClient(cfg.AgentURL, cfg.CookieName)
	proxySrv := httpserver.New(cfg, client, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/graphql", nil)
	rr := httptest.NewRecorder()
	proxySrv.Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
