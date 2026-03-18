//go:build caddy

package caddy

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(AuthSentinelModule{})
	httpcaddyfile.RegisterHandlerDirective("authsentinel", parseCaddyfile)
}

func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m AuthSentinelModule
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return &m, err
}

// AuthSentinelModule implements forward-auth to an AuthSentinel proxy instance.
// Directive: authsentinel <auth_proxy_url>
// Example: authsentinel http://localhost:8081
type AuthSentinelModule struct {
	AuthURL string `json:"auth_url,omitempty"`
}

// CaddyModule returns the module info.
func (AuthSentinelModule) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.authsentinel",
		New: func() caddy.Module { return &AuthSentinelModule{} },
	}
}

// UnmarshalCaddyfile parses the directive. Syntax: authsentinel <url>
func (m *AuthSentinelModule) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if !d.NextArg() {
			continue
		}
		m.AuthURL = d.Val()
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler. Forwards the request to the
// AuthSentinel proxy (AuthURL); on 2xx copies authResponseHeaders and invokes next;
// otherwise writes the proxy response (deny).
func (m AuthSentinelModule) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	if m.AuthURL == "" {
		http.Error(w, "authsentinel: auth_url not configured", http.StatusInternalServerError)
		return nil
	}
	authURL := strings.TrimSuffix(m.AuthURL, "/")
	target := authURL + r.URL.Path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	u, err := url.Parse(target)
	if err != nil {
		http.Error(w, "authsentinel: invalid auth_url", http.StatusInternalServerError)
		return nil
	}
	// Build forward-auth request: same method, path, copy relevant headers
	body, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewReader(body)) // restore for next handler
	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, bytes.NewReader(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	req.URL = u
	for k, v := range r.Header {
		if len(v) > 0 && !strings.EqualFold(k, "Host") {
			req.Header.Set(k, v[0])
		}
	}
	req.Host = r.Host
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Allow: copy headers that the proxy set (principal, etc.) to the response
		for k, v := range resp.Header {
			if len(v) > 0 && shouldForwardHeader(k) {
				w.Header().Set(k, v[0])
			}
		}
		return next.ServeHTTP(w, r)
	}
	// Deny: copy status and body
	for k, v := range resp.Header {
		if len(v) > 0 {
			w.Header().Set(k, v[0])
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
	return nil
}

func shouldForwardHeader(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasPrefix(lower, "x-") || lower == "authorization"
}
