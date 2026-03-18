package service

import (
	"context"
	"testing"
	"time"

	"github.com/ArmanAvanesyan/authsentinel/internal/agent/config"
	"github.com/ArmanAvanesyan/authsentinel/pkg/agent"
	"github.com/ArmanAvanesyan/authsentinel/pkg/cookie"
	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginapi"
	"github.com/ArmanAvanesyan/authsentinel/pkg/session"
	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
)

type inMemorySessionStore struct {
	data map[string]*session.Session
}

func newInMemorySessionStore() *inMemorySessionStore {
	return &inMemorySessionStore{data: make(map[string]*session.Session)}
}

func (s *inMemorySessionStore) Get(ctx context.Context, id string) (*session.Session, error) {
	return s.data[id], nil
}

func (s *inMemorySessionStore) Set(ctx context.Context, id string, sess *session.Session, ttlSeconds int) error {
	s.data[id] = sess
	return nil
}

func (s *inMemorySessionStore) Delete(ctx context.Context, id string) error {
	delete(s.data, id)
	return nil
}

type inMemoryPKCEStore struct {
	data map[string]*session.PKCEState
}

func newInMemoryPKCEStore() *inMemoryPKCEStore {
	return &inMemoryPKCEStore{data: make(map[string]*session.PKCEState)}
}

func (s *inMemoryPKCEStore) Get(ctx context.Context, state string) (*session.PKCEState, error) {
	return s.data[state], nil
}

func (s *inMemoryPKCEStore) Set(ctx context.Context, state string, p *session.PKCEState, ttlSeconds int) error {
	s.data[state] = p
	return nil
}

func (s *inMemoryPKCEStore) Delete(ctx context.Context, state string) error {
	delete(s.data, state)
	return nil
}

type inMemoryRefreshLockStore struct{}

func (s *inMemoryRefreshLockStore) Obtain(ctx context.Context, sessionID string, ttlSeconds int) (bool, error) {
	return true, nil
}
func (s *inMemoryRefreshLockStore) Release(ctx context.Context, sessionID string) error { return nil }

type mockProvider struct {
	authCalls     int
	authState     string
	authChallenge string
	authNonce     string

	exchangeCalls       int
	exchangeCode        string
	exchangeVerifier    string
	exchangeRedirectURI string

	refreshCalls    int
	refreshToken    string

	endCalls            int
	endIDTokenHint     string
	endPostLogoutURI   string
}

func (m *mockProvider) Descriptor() pluginapi.PluginDescriptor {
	return pluginapi.PluginDescriptor{ID: "provider:mock", Kind: pluginapi.PluginKindProvider, Name: "mock"}
}
func (m *mockProvider) Health(ctx context.Context) pluginapi.PluginHealth {
	return pluginapi.PluginHealth{State: pluginapi.PluginStateHealthy}
}
func (m *mockProvider) AuthorizationURL(ctx context.Context, state string, codeChallenge string, nonce string, extraParams map[string]string) (string, error) {
	m.authCalls++
	m.authState = state
	m.authChallenge = codeChallenge
	m.authNonce = nonce
	_ = extraParams
	return "https://idp.example/auth?state=" + state, nil
}
func (m *mockProvider) ExchangeCode(ctx context.Context, code string, codeVerifier string, redirectURI string) (*pluginapi.ProviderTokens, error) {
	m.exchangeCalls++
	m.exchangeCode = code
	m.exchangeVerifier = codeVerifier
	m.exchangeRedirectURI = redirectURI
	// Return empty ID token so LoginEnd deterministically fails token validation in tests.
	return &pluginapi.ProviderTokens{
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
		IDToken:      "",
		ExpiresIn:    0,
	}, nil
}
func (m *mockProvider) Refresh(ctx context.Context, refreshToken string) (*pluginapi.ProviderTokens, error) {
	m.refreshCalls++
	m.refreshToken = refreshToken
	return &pluginapi.ProviderTokens{
		AccessToken:  "access-2",
		RefreshToken: refreshToken,
		IDToken:      "",
		ExpiresIn:    3600,
	}, nil
}
func (m *mockProvider) EndSessionURL(ctx context.Context, idTokenHint, postLogoutRedirectURI string) (string, error) {
	m.endCalls++
	m.endIDTokenHint = idTokenHint
	m.endPostLogoutURI = postLogoutRedirectURI
	return "https://idp.example/logout?ok=1", nil
}

func newTestConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := &config.Config{
		OIDCIssuer:            "https://issuer.example",
		OIDCRedirectURI:       "https://app.example.com/callback",
		OIDCClientID:          "client-1",
		RedisURL:              "redis://localhost:6379",
		CookieSigningSecret:  "test-cookie-signing-secret-32bytes",
		AppBaseURL:            "https://app.example.com",
		HTTPPort:              "8080",
		CookieName:            "test_session",
	}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("config.Validate: %v", err)
	}
	return cfg
}

func TestService_LoginStart_UsesProviderAuthorizationURL(t *testing.T) {
	ctx := context.Background()
	cfg := newTestConfig(t)

	sessions := newInMemorySessionStore()
	pkce := newInMemoryPKCEStore()
	refreshLock := &inMemoryRefreshLockStore{}
	cookieManager := cookie.NewSignedManager(cfg.CookieSigningSecret)

	prov := &mockProvider{}
	svc, err := New(cfg, sessions, pkce, refreshLock, cookieManager, token.JWKSSource(nil), prov, nil)
	if err != nil {
		t.Fatalf("service.New: %v", err)
	}

	resp, err := svc.LoginStart(ctx, agent.LoginStartRequest{RedirectTo: "/welcome"})
	if err != nil {
		t.Fatalf("LoginStart: %v", err)
	}
	if prov.authCalls != 1 {
		t.Fatalf("expected 1 AuthorizationURL call, got %d", prov.authCalls)
	}
	if prov.authState == "" || prov.authChallenge == "" || prov.authNonce == "" {
		t.Fatalf("expected non-empty PKCE/state values, got state=%q challenge=%q nonce=%q", prov.authState, prov.authChallenge, prov.authNonce)
	}
	if resp.RedirectURL == "" {
		t.Fatalf("expected redirect URL")
	}

	st, err := pkce.Get(ctx, prov.authState)
	if err != nil {
		t.Fatalf("pkce.Get: %v", err)
	}
	if st == nil {
		t.Fatalf("expected pkce state to be stored")
	}
	if st.CodeChallenge != prov.authChallenge {
		t.Fatalf("code_challenge mismatch: got %q want %q", st.CodeChallenge, prov.authChallenge)
	}
	if st.Nonce != prov.authNonce {
		t.Fatalf("nonce mismatch: got %q want %q", st.Nonce, prov.authNonce)
	}
	if st.RedirectTo != "https://app.example.com/welcome" {
		t.Fatalf("unexpected redirect_to: got %q", st.RedirectTo)
	}
}

func TestService_LoginEnd_UsesProviderExchangeCode(t *testing.T) {
	ctx := context.Background()
	cfg := newTestConfig(t)

	sessions := newInMemorySessionStore()
	pkce := newInMemoryPKCEStore()
	refreshLock := &inMemoryRefreshLockStore{}
	cookieManager := cookie.NewSignedManager(cfg.CookieSigningSecret)

	prov := &mockProvider{}
	svc, err := New(cfg, sessions, pkce, refreshLock, cookieManager, token.JWKSSource(nil), prov, nil)
	if err != nil {
		t.Fatalf("service.New: %v", err)
	}

	state := "state-123"
	if err := pkce.Set(ctx, state, &session.PKCEState{
		State:         state,
		CodeVerifier: "verifier-123",
		Nonce:         "nonce-123",
		RedirectTo:   "https://app.example.com/welcome",
	}, cfg.SessionPKCETTLSeconds); err != nil {
		t.Fatalf("pkce.Set: %v", err)
	}

	resp, err := svc.LoginEnd(ctx, agent.LoginEndRequest{Code: "auth-code-1", State: state})
	if err != nil {
		t.Fatalf("LoginEnd: %v", err)
	}
	if prov.exchangeCalls != 1 {
		t.Fatalf("expected 1 ExchangeCode call, got %d", prov.exchangeCalls)
	}
	if prov.exchangeCode != "auth-code-1" {
		t.Fatalf("unexpected exchange code: %q", prov.exchangeCode)
	}
	if prov.exchangeVerifier != "verifier-123" {
		t.Fatalf("unexpected code_verifier: %q", prov.exchangeVerifier)
	}
	if prov.exchangeRedirectURI != cfg.OIDCRedirectURI {
		t.Fatalf("unexpected redirectURI: got %q want %q", prov.exchangeRedirectURI, cfg.OIDCRedirectURI)
	}

	// Token validation must fail (mock returns empty ID token), so service should clear cookie.
	if !resp.ClearCookie {
		t.Fatalf("expected ClearCookie=true on failure")
	}
	if resp.RedirectURL != cfg.AppBaseURL+cfg.LoginErrorRedirectPath {
		t.Fatalf("unexpected redirectURL: got %q want %q", resp.RedirectURL, cfg.AppBaseURL+cfg.LoginErrorRedirectPath)
	}
	if len(sessions.data) != 0 {
		t.Fatalf("expected no session created, got %d sessions", len(sessions.data))
	}
	st, _ := pkce.Get(ctx, state)
	if st != nil {
		t.Fatalf("expected pkce state to be deleted on LoginEnd")
	}
}

func TestService_Refresh_UsesProviderRefresh(t *testing.T) {
	ctx := context.Background()
	cfg := newTestConfig(t)

	sessions := newInMemorySessionStore()
	pkce := newInMemoryPKCEStore()
	refreshLock := &inMemoryRefreshLockStore{}
	cookieManager := cookie.NewSignedManager(cfg.CookieSigningSecret)

	// Pre-seed a session that needs refresh.
	sessID := "sess-1"
	sessions.data[sessID] = &session.Session{
		ID:           sessID,
		AccessToken: "access-1",
		RefreshToken: "refresh-1",
		ExpiresAt:   time.Now().Unix() + int64(cfg.SessionRefreshEarlySeconds/2),
		Claims:       map[string]any{"sub": "u1"},
	}

	prov := &mockProvider{}
	svc, err := New(cfg, sessions, pkce, refreshLock, cookieManager, token.JWKSSource(nil), prov, nil)
	if err != nil {
		t.Fatalf("service.New: %v", err)
	}

	cookieVal, err := cookieManager.Encode(sessID)
	if err != nil {
		t.Fatalf("cookie.Encode: %v", err)
	}

	resp, err := svc.Refresh(ctx, agent.RefreshRequest{SessionCookie: cookieVal})
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if !resp.Refreshed {
		t.Fatalf("expected refreshed=true")
	}
	if prov.refreshCalls != 1 {
		t.Fatalf("expected 1 Refresh call, got %d", prov.refreshCalls)
	}
	if prov.refreshToken != "refresh-1" {
		t.Fatalf("unexpected refresh token: %q", prov.refreshToken)
	}
	updated, _ := sessions.Get(ctx, sessID)
	if updated == nil || updated.AccessToken != "access-2" {
		t.Fatalf("expected access token to be updated, got %#v", updated)
	}
}

func TestService_Logout_UsesProviderEndSessionURL(t *testing.T) {
	ctx := context.Background()
	cfg := newTestConfig(t)

	sessions := newInMemorySessionStore()
	pkce := newInMemoryPKCEStore()
	refreshLock := &inMemoryRefreshLockStore{}
	cookieManager := cookie.NewSignedManager(cfg.CookieSigningSecret)

	// Pre-seed a session for logout.
	sessID := "sess-1"
	sessions.data[sessID] = &session.Session{
		ID:        sessID,
		IDToken:  "id-token-hint",
		Claims:    map[string]any{"sub": "u1"},
	}

	prov := &mockProvider{}
	svc, err := New(cfg, sessions, pkce, refreshLock, cookieManager, token.JWKSSource(nil), prov, nil)
	if err != nil {
		t.Fatalf("service.New: %v", err)
	}

	cookieVal, err := cookieManager.Encode(sessID)
	if err != nil {
		t.Fatalf("cookie.Encode: %v", err)
	}

	resp, err := svc.Logout(ctx, agent.LogoutRequest{
		SessionCookie: cookieVal,
		RedirectTo:    "/loggedout",
	})
	if err != nil {
		t.Fatalf("Logout: %v", err)
	}
	if prov.endCalls != 1 {
		t.Fatalf("expected 1 EndSessionURL call, got %d", prov.endCalls)
	}
	if prov.endIDTokenHint != "id-token-hint" {
		t.Fatalf("unexpected id_token_hint: %q", prov.endIDTokenHint)
	}
	if prov.endPostLogoutURI != "https://app.example.com/loggedout" {
		t.Fatalf("unexpected postLogoutRedirectURI: %q", prov.endPostLogoutURI)
	}
	if resp.ClearCookie != true {
		t.Fatalf("expected ClearCookie=true")
	}
	if resp.RedirectURL == "" {
		t.Fatalf("expected redirect URL")
	}
	// Logout should delete session.
	deleted, _ := sessions.Get(ctx, sessID)
	if deleted != nil {
		t.Fatalf("expected session deletion on logout")
	}
}

