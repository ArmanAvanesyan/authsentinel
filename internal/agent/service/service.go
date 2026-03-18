package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ArmanAvanesyan/authsentinel/internal/agent/config"
	"github.com/ArmanAvanesyan/authsentinel/internal/agent/oidc"
	"github.com/ArmanAvanesyan/authsentinel/pkg/agent"
	"github.com/ArmanAvanesyan/authsentinel/pkg/cookie"
	"github.com/ArmanAvanesyan/authsentinel/pkg/observability"
	"github.com/ArmanAvanesyan/authsentinel/pkg/pluginapi"
	"github.com/ArmanAvanesyan/authsentinel/pkg/session"
	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
)

// Service implements agent.Service.
type Service struct {
	cfg         *config.Config
	provider    pluginapi.ProviderPlugin
	jwks        token.JWKSSource
	sessions    session.SessionStore
	pkce        session.PKCEStore
	refreshLock session.RefreshLockStore
	cookie      cookie.Manager
	cookieOpts  cookie.CookieOptions
	tracer      observability.Tracer
}

// New creates an agent Service.
func New(
	cfg *config.Config,
	sessions session.SessionStore,
	pkce session.PKCEStore,
	refreshLock session.RefreshLockStore,
	cookieManager cookie.Manager,
	jwks token.JWKSSource,
	provider pluginapi.ProviderPlugin,
	tracer observability.Tracer,
) (*Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if provider == nil {
		return nil, fmt.Errorf("provider is required")
	}
	opts := cookie.CookieOptions{
		Path:     "/",
		Domain:   cfg.CookieDomain,
		Secure:   bool(cfg.CookieSecure),
		HTTPOnly: true,
		SameSite: cfg.CookieSameSite,
		MaxAge:   cfg.SessionTTLSeconds,
	}
	if tracer == nil {
		tracer = observability.NopTracer{}
	}
	return &Service{
		cfg:         cfg,
		provider:    provider,
		jwks:        jwks,
		sessions:    sessions,
		pkce:        pkce,
		refreshLock: refreshLock,
		cookie:      cookieManager,
		cookieOpts:  opts,
		tracer:      tracer,
	}, nil
}

// Session implements agent.Service.
func (s *Service) Session(ctx context.Context, req agent.SessionRequest) (*agent.SessionResponse, error) {
	if req.SessionCookie == "" {
		return &agent.SessionResponse{IsAuthenticated: false}, nil
	}
	var sessionID string
	if err := s.cookie.Decode(req.SessionCookie, &sessionID); err != nil {
		return &agent.SessionResponse{IsAuthenticated: false}, nil
	}
	sess, err := s.sessions.Get(ctx, sessionID)
	if err != nil || sess == nil {
		return &agent.SessionResponse{IsAuthenticated: false}, nil
	}
	user := sessionToUser(sess)
	return &agent.SessionResponse{
		IsAuthenticated: true,
		User:            user,
	}, nil
}

// LoginStart implements agent.Service.
func (s *Service) LoginStart(ctx context.Context, req agent.LoginStartRequest) (*agent.LoginStartResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx, "agent.login_start", "redirect_to", req.RedirectTo)
	defer span.End()

	redirectTo := ValidateRedirect(req.RedirectTo, s.cfg.AppBaseURL, s.cfg.AllowedRedirectOrigins, s.cfg.AllowedRedirectPaths)
	if redirectTo == "" && req.RedirectTo != "" {
		redirectTo = s.cfg.AppBaseURL
	}
	if redirectTo == "" {
		redirectTo = s.cfg.AppBaseURL
	}
	_, pkceSpan := s.tracer.StartSpan(ctx, "agent.pkce_generate")
	verifier, challenge, nonce, err := oidc.GeneratePKCE()
	if err != nil {
		pkceSpan.End()
		return nil, err
	}
	state, err := oidc.GenerateState()
	if err != nil {
		pkceSpan.End()
		return nil, err
	}
	pkceSpan.End()
	_, pkceStoreSpan := s.tracer.StartSpan(ctx, "agent.pkce_store_set")
	err = s.pkce.Set(ctx, state, &session.PKCEState{
		State:         state,
		CodeVerifier:  verifier,
		CodeChallenge: challenge,
		Nonce:         nonce,
		RedirectTo:    redirectTo,
	}, s.cfg.SessionPKCETTLSeconds)
	if err != nil {
		pkceStoreSpan.End()
		return nil, err
	}
	pkceStoreSpan.End()
	_, authURLSpan := s.tracer.StartSpan(ctx, "agent.oidc_authorization_url")
	authURL, err := s.provider.AuthorizationURL(ctx, state, challenge, nonce, nil)
	if err != nil {
		authURLSpan.End()
		return nil, err
	}
	authURLSpan.End()
	return &agent.LoginStartResponse{RedirectURL: authURL}, nil
}

// LoginEnd implements agent.Service.
func (s *Service) LoginEnd(ctx context.Context, req agent.LoginEndRequest) (*agent.LoginEndResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx, "agent.login_end", "state_present", req.State != "", "code_present", req.Code != "")
	defer span.End()

	if req.Error != "" {
		redirectURL := s.cfg.AppBaseURL + s.cfg.LoginErrorRedirectPath
		return &agent.LoginEndResponse{RedirectURL: redirectURL, ClearCookie: true}, nil
	}
	if req.State == "" || req.Code == "" {
		redirectURL := s.cfg.AppBaseURL + s.cfg.LoginErrorRedirectPath
		return &agent.LoginEndResponse{RedirectURL: redirectURL, ClearCookie: true}, nil
	}
	_, pkceGetSpan := s.tracer.StartSpan(ctx, "agent.pkce_get")
	p, err := s.pkce.Get(ctx, req.State)
	if err != nil || p == nil {
		pkceGetSpan.End()
		redirectURL := s.cfg.AppBaseURL + s.cfg.LoginErrorRedirectPath
		return &agent.LoginEndResponse{RedirectURL: redirectURL, ClearCookie: true}, nil
	}
	pkceGetSpan.End()
	_ = s.pkce.Delete(ctx, req.State)
	_, exchangeSpan := s.tracer.StartSpan(ctx, "agent.oidc_exchange_code")
	tr, err := s.provider.ExchangeCode(ctx, req.Code, p.CodeVerifier, s.cfg.OIDCRedirectURI)
	if err != nil {
		exchangeSpan.End()
		redirectURL := s.cfg.AppBaseURL + s.cfg.LoginErrorRedirectPath
		return &agent.LoginEndResponse{RedirectURL: redirectURL, ClearCookie: true}, nil
	}
	exchangeSpan.End()
	audience := s.cfg.OIDCAudience
	if audience == "" {
		audience = s.cfg.OIDCClientID
	}
	_, validateSpan := s.tracer.StartSpan(ctx, "agent.token_validate_id_token")
	principal, err := token.ValidateIDToken(ctx, tr.IDToken, s.jwks, s.cfg.OIDCIssuer, audience, p.Nonce)
	if err != nil {
		validateSpan.End()
		redirectURL := s.cfg.AppBaseURL + s.cfg.LoginErrorRedirectPath
		return &agent.LoginEndResponse{RedirectURL: redirectURL, ClearCookie: true}, nil
	}
	validateSpan.End()
	expiresAt := time.Now().Unix() + int64(tr.ExpiresIn)
	if tr.ExpiresIn <= 0 {
		expiresAt = time.Now().Add(24 * time.Hour).Unix()
	}
	claims := principal.Claims
	if claims == nil {
		claims = make(map[string]any)
	}
	claims["sub"] = principal.Subject
	if principal.Roles != nil {
		claims["roles"] = principal.Roles
	}
	sessID, err := generateSessionID()
	if err != nil {
		return nil, err
	}
	sess := &session.Session{
		ID:           sessID,
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		IDToken:      tr.IDToken,
		ExpiresAt:    expiresAt,
		Claims:       claims,
	}
	_, sessionSetSpan := s.tracer.StartSpan(ctx, "agent.session_store_set")
	err = s.sessions.Set(ctx, sessID, sess, s.cfg.SessionTTLSeconds)
	if err != nil {
		sessionSetSpan.End()
		return nil, err
	}
	sessionSetSpan.End()
	if s.cfg.PostLoginWebhookURL != "" {
		_ = s.callPostLoginWebhook(ctx, s.cfg.PostLoginWebhookURL, sessID, principal.Subject, getClaimString(claims, "email"), claims, req.Host)
	}
	_, cookieEncodeSpan := s.tracer.StartSpan(ctx, "agent.cookie_encode_session_id")
	cookieValue, err := s.cookie.Encode(sessID)
	if err != nil {
		cookieEncodeSpan.End()
		return nil, err
	}
	cookieEncodeSpan.End()
	redirectURL := ValidateRedirect(p.RedirectTo, s.cfg.AppBaseURL, s.cfg.AllowedRedirectOrigins, s.cfg.AllowedRedirectPaths)
	if redirectURL == "" {
		redirectURL = s.cfg.AppBaseURL
	}
	return &agent.LoginEndResponse{
		RedirectURL:    redirectURL,
		SetCookieValue: cookieValue,
	}, nil
}

func getClaimString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func (s *Service) callPostLoginWebhook(ctx context.Context, url, sessionID, subject, email string, claims map[string]any, host string) error {
	body := map[string]any{
		"session_id": sessionID,
		"subject":    subject,
		"email":      email,
		"claims":     claims,
		"host":       host,
	}
	data, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook status %d", resp.StatusCode)
	}
	return nil
}

// Refresh implements agent.Service.
func (s *Service) Refresh(ctx context.Context, req agent.RefreshRequest) (*agent.RefreshResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx, "agent.refresh")
	defer span.End()

	if req.SessionCookie == "" {
		return nil, fmt.Errorf("no session cookie")
	}
	var sessionID string
	if err := s.cookie.Decode(req.SessionCookie, &sessionID); err != nil {
		return nil, fmt.Errorf("invalid cookie")
	}
	_, sessionGetSpan := s.tracer.StartSpan(ctx, "agent.session_store_get")
	sess, err := s.sessions.Get(ctx, sessionID)
	if err != nil || sess == nil {
		sessionGetSpan.End()
		return nil, fmt.Errorf("session not found")
	}
	sessionGetSpan.End()
	if sess.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token")
	}
	now := time.Now()
	if !sess.NeedsRefresh(now, s.cfg.SessionRefreshEarlySeconds) {
		return &agent.RefreshResponse{}, nil
	}
	_, lockSpan := s.tracer.StartSpan(ctx, "agent.refresh_lock_obtain")
	ok, err := s.refreshLock.Obtain(ctx, sessionID, s.cfg.SessionRefreshLockTTLSeconds)
	if err != nil || !ok {
		lockSpan.End()
		return &agent.RefreshResponse{}, nil
	}
	lockSpan.End()
	defer func() { _ = s.refreshLock.Release(ctx, sessionID) }()
	_, providerRefreshSpan := s.tracer.StartSpan(ctx, "agent.oidc_refresh")
	tr, err := s.provider.Refresh(ctx, sess.RefreshToken)
	if err != nil {
		providerRefreshSpan.End()
		return nil, fmt.Errorf("refresh failed: %w", err)
	}
	providerRefreshSpan.End()
	expiresAt := time.Now().Unix() + int64(tr.ExpiresIn)
	if tr.ExpiresIn <= 0 {
		expiresAt = time.Now().Add(24 * time.Hour).Unix()
	}
	sess.AccessToken = tr.AccessToken
	sess.ExpiresAt = expiresAt
	if tr.RefreshToken != "" {
		sess.RefreshToken = tr.RefreshToken
	}
	if tr.IDToken != "" {
		sess.IDToken = tr.IDToken
		aud := s.cfg.OIDCAudience
		if aud == "" {
			aud = s.cfg.OIDCClientID
		}
		_, validateSpan := s.tracer.StartSpan(ctx, "agent.token_validate_id_token_refresh")
		principal, err := token.ValidateIDToken(ctx, tr.IDToken, s.jwks, s.cfg.OIDCIssuer, aud, "")
		if err == nil && principal.Claims != nil {
			sess.Claims = principal.Claims
		}
		validateSpan.End()
	}
	_, sessionSetSpan := s.tracer.StartSpan(ctx, "agent.session_store_set")
	err = s.sessions.Set(ctx, sessionID, sess, s.cfg.SessionTTLSeconds)
	if err != nil {
		sessionSetSpan.End()
		return nil, err
	}
	sessionSetSpan.End()
	_, cookieEncodeSpan := s.tracer.StartSpan(ctx, "agent.cookie_encode_session_id")
	cookieValue, err := s.cookie.Encode(sessionID)
	if err != nil {
		cookieEncodeSpan.End()
		return nil, err
	}
	cookieEncodeSpan.End()
	return &agent.RefreshResponse{
		SetCookieValue: cookieValue,
		Refreshed:      true,
	}, nil
}

// Logout implements agent.Service.
func (s *Service) Logout(ctx context.Context, req agent.LogoutRequest) (*agent.LogoutResponse, error) {
	ctx, span := s.tracer.StartSpan(ctx, "agent.logout")
	defer span.End()

	// CSRF: for POST, check Origin/Referer against allowed
	if req.Origin != "" || req.Referer != "" {
		allowed := false
		baseURL := strings.TrimSuffix(s.cfg.AppBaseURL, "/")
		for _, o := range s.cfg.AllowedRedirectOrigins {
			if req.Origin == o || req.Referer == o || req.Referer == baseURL+"/" || strings.HasPrefix(req.Referer, baseURL+"/") {
				allowed = true
				break
			}
		}
		if !allowed && (req.Origin != "" || req.Referer != "") {
			// Strict: require same origin
			if req.Origin != baseURL && req.Origin != "" {
				return nil, fmt.Errorf("csrf: origin not allowed")
			}
		}
	}
	redirectTo := ValidateRedirect(req.RedirectTo, s.cfg.AppBaseURL, s.cfg.AllowedRedirectOrigins, s.cfg.AllowedRedirectPaths)
	if redirectTo == "" {
		redirectTo = s.cfg.AppBaseURL
	}
	var sessionID string
	if req.SessionCookie != "" {
		_ = s.cookie.Decode(req.SessionCookie, &sessionID)
	}
	var idTokenHint string
	if sessionID != "" {
		_, sessGetSpan := s.tracer.StartSpan(ctx, "agent.session_store_get_logout")
		sess, _ := s.sessions.Get(ctx, sessionID)
		if sess != nil {
			idTokenHint = sess.IDToken
			_ = s.sessions.Delete(ctx, sessionID)
		}
		sessGetSpan.End()
	}
	_, endURLSpan := s.tracer.StartSpan(ctx, "agent.oidc_end_session")
	endURL, err := s.provider.EndSessionURL(ctx, idTokenHint, redirectTo)
	if err != nil {
		endURLSpan.End()
		return &agent.LogoutResponse{RedirectURL: redirectTo, ClearCookie: true}, nil
	}
	endURLSpan.End()
	if endURL == "" {
		return &agent.LogoutResponse{RedirectURL: redirectTo, ClearCookie: true}, nil
	}
	return &agent.LogoutResponse{
		RedirectURL: endURL,
		ClearCookie: true,
	}, nil
}

func sessionToUser(sess *session.Session) *agent.SessionUser {
	user := &agent.SessionUser{
		Claims: sess.Claims,
	}
	if sub, ok := sess.Claims["sub"].(string); ok {
		user.Sub = sub
	}
	if email, ok := sess.Claims["email"].(string); ok {
		user.Email = email
	}
	if u, ok := sess.Claims["preferred_username"].(string); ok {
		user.PreferredUsername = u
	}
	if name, ok := sess.Claims["name"].(string); ok {
		user.Name = name
	}
	if r, ok := sess.Claims["realm_access"].(map[string]any); ok {
		if roles, ok := r["roles"].([]interface{}); ok {
			for _, x := range roles {
				if s, ok := x.(string); ok {
					user.Roles = append(user.Roles, s)
				}
			}
		}
	}
	if user.Roles == nil {
		if r, ok := sess.Claims["roles"].([]interface{}); ok {
			for _, x := range r {
				if s, ok := x.(string); ok {
					user.Roles = append(user.Roles, s)
				}
			}
		}
	}
	if g, ok := sess.Claims["groups"].([]interface{}); ok {
		for _, x := range g {
			if s, ok := x.(string); ok {
				user.Groups = append(user.Groups, s)
			}
		}
	}
	for _, r := range user.Roles {
		if r == "admin" || r == "administrator" {
			user.IsAdmin = true
			break
		}
	}
	if sess.TenantContext != nil {
		user.TenantContext = map[string]any{
			"tenant_id":   sess.TenantContext.TenantID,
			"tenant_slug": sess.TenantContext.TenantSlug,
			"status":      sess.TenantContext.Status,
			"locale":      sess.TenantContext.Locale,
			"timezone":    sess.TenantContext.Timezone,
		}
	}
	return user
}

func generateSessionID() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Resolve returns access_token and claims for the proxy (internal use). If the session was refreshed, setCookieValue is non-empty.
func (s *Service) Resolve(ctx context.Context, sessionCookie string) (accessToken string, claims map[string]any, tenantContext map[string]any, setCookieValue string, err error) {
	if sessionCookie == "" {
		return "", nil, nil, "", fmt.Errorf("no session cookie")
	}
	var sessionID string
	if err := s.cookie.Decode(sessionCookie, &sessionID); err != nil {
		return "", nil, nil, "", fmt.Errorf("invalid cookie")
	}
	sess, err := s.sessions.Get(ctx, sessionID)
	if err != nil || sess == nil {
		return "", nil, nil, "", fmt.Errorf("session not found")
	}
	claims = sess.Claims
	if claims == nil {
		claims = make(map[string]any)
	}
	var tc map[string]any
	if sess.TenantContext != nil {
		tc = map[string]any{
			"tenant_id":   sess.TenantContext.TenantID,
			"tenant_slug": sess.TenantContext.TenantSlug,
			"status":      sess.TenantContext.Status,
			"locale":      sess.TenantContext.Locale,
			"timezone":    sess.TenantContext.Timezone,
		}
	}
	return sess.AccessToken, claims, tc, "", nil
}

// AttachTenantContext updates the session's tenant_context (Option A enrichment).
func (s *Service) AttachTenantContext(ctx context.Context, sessionID string, tc *session.TenantContext) error {
	if sessionID == "" || tc == nil {
		return fmt.Errorf("session_id and tenant_context required")
	}
	sess, err := s.sessions.Get(ctx, sessionID)
	if err != nil || sess == nil {
		return fmt.Errorf("session not found")
	}
	sess.TenantContext = tc
	return s.sessions.Set(ctx, sessionID, sess, s.cfg.SessionTTLSeconds)
}

// Ensure Service implements agent.Service.
var _ agent.Service = (*Service)(nil)
