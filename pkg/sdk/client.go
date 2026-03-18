package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// AgentClient is a server-side HTTP client for the AuthSentinel agent's public endpoints.
// Use it to build login/logout URLs and to call /session and /refresh with the user's session cookie
// (e.g. from a BFF that forwards the cookie to the agent).
type AgentClient struct {
	BaseURL    string
	CookieName string
	HTTPClient *http.Client
}

// NewAgentClient returns a client for the agent at baseURL (e.g. https://auth.example.com).
// cookieName must match the agent's configured session cookie name.
func NewAgentClient(baseURL, cookieName string) *AgentClient {
	baseURL = strings.TrimSuffix(baseURL, "/")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	return &AgentClient{
		BaseURL:    baseURL,
		CookieName: cookieName,
		HTTPClient: &http.Client{},
	}
}

// GetLoginURL returns the agent login URL; redirect the user here to start the login flow.
// returnURL is where the user should land after successful login (validated by the agent).
func (c *AgentClient) GetLoginURL(returnURL string) string {
	q := url.Values{}
	if returnURL != "" {
		q.Set("redirect_to", returnURL)
	}
	return c.BaseURL + "/login?" + q.Encode()
}

// GetLogoutURL returns the agent logout URL; redirect the user here to log out.
// redirectTo is the post-logout redirect (validated by the agent).
func (c *AgentClient) GetLogoutURL(redirectTo string) string {
	q := url.Values{}
	if redirectTo != "" {
		q.Set("redirect_to", redirectTo)
	}
	return c.BaseURL + "/logout?" + q.Encode()
}

// SessionResponse is the JSON shape returned by GET /session (is_authenticated, user).
// SetCookie is populated from the response Set-Cookie header when the agent issues a new session cookie.
type SessionResponse struct {
	IsAuthenticated bool        `json:"is_authenticated"`
	User            *SessionUser `json:"user,omitempty"`
	SetCookie       string     `json:"-"` // Filled from response header when present
}

// SessionUser is the user object in /session and /me.
type SessionUser struct {
	Sub               string         `json:"sub"`
	Email             string         `json:"email,omitempty"`
	PreferredUsername string         `json:"preferred_username,omitempty"`
	Name              string         `json:"name,omitempty"`
	Roles             []string       `json:"roles,omitempty"`
	Groups            []string       `json:"groups,omitempty"`
	IsAdmin           bool           `json:"is_admin,omitempty"`
	TenantContext     map[string]any `json:"tenant_context,omitempty"`
	Claims            map[string]any `json:"claims,omitempty"`
}

// GetSession calls GET /session with the given session cookie and returns session state.
// If the agent returns a new Set-Cookie (e.g. after refresh), it is set on out.SetCookie for the caller to forward.
func (c *AgentClient) GetSession(ctx context.Context, sessionCookie string) (*SessionResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/session", nil)
	if err != nil {
		return nil, err
	}
	if sessionCookie != "" {
		req.Header.Set("Cookie", c.CookieName+"="+sessionCookie)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("agent session: status %d: %s", resp.StatusCode, string(body))
	}
	var out SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if setCookie := resp.Header.Get("Set-Cookie"); setCookie != "" {
		out.SetCookie = setCookie
	}
	return &out, nil
}

// RefreshResult is the result of a refresh call; SetCookie is set when the agent issues a new session cookie.
type RefreshResult struct {
	Refreshed bool
	SetCookie string
}

// Refresh calls GET /refresh with the given session cookie to refresh the session.
// Returns Refreshed and SetCookie when the agent issues a new cookie.
func (c *AgentClient) Refresh(ctx context.Context, sessionCookie string) (*RefreshResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/refresh", nil)
	if err != nil {
		return nil, err
	}
	if sessionCookie != "" {
		req.Header.Set("Cookie", c.CookieName+"="+sessionCookie)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusUnauthorized {
		return &RefreshResult{Refreshed: false}, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("agent refresh: status %d: %s", resp.StatusCode, string(body))
	}
	setCookie := resp.Header.Get("Set-Cookie")
	return &RefreshResult{Refreshed: setCookie != "", SetCookie: setCookie}, nil
}
