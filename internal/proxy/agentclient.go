package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ResolveResponse is the response from agent GET /internal/resolve.
type ResolveResponse struct {
	AccessToken   string         `json:"access_token"`
	Claims        map[string]any `json:"claims"`
	TenantContext map[string]any `json:"tenant_context"`
}

// AgentClient calls the authsentinel-agent for session resolution.
type AgentClient struct {
	baseURL    string
	cookieName string
	httpClient *http.Client
}

// NewAgentClient creates a client for the agent at baseURL (e.g. http://authsentinel-agent:8080).
func NewAgentClient(baseURL, cookieName string) *AgentClient {
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &AgentClient{
		baseURL:    baseURL,
		cookieName: cookieName,
		httpClient: &http.Client{},
	}
}

// Resolve calls GET /internal/resolve with the session cookie and returns session data.
func (c *AgentClient) Resolve(ctx context.Context, sessionCookie string) (*ResolveResponse, error) {
	url := c.baseURL + "/internal/resolve"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if sessionCookie != "" {
		req.Header.Set("Cookie", c.cookieName+"="+sessionCookie)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("agent resolve: status %d: %s", resp.StatusCode, string(body))
	}
	var out ResolveResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetCookieFromResponse copies Set-Cookie headers from the agent response to the proxy response.
func SetCookieFromResponse(dst, src http.Header) {
	for _, v := range src["Set-Cookie"] {
		dst.Add("Set-Cookie", v)
	}
}
