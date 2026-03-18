package caddy

// Config is the configuration shape for the AuthSentinel Caddy adapter.
// It can be used from Caddyfile (directive args) or JSON config.
// All adapters call the same proxy/policy/token runtime; this config
// only specifies how to reach it (in-process Engine or forward-auth URL).
type Config struct {
	// AuthURL is the AuthSentinel proxy base URL when using forward-auth mode.
	// Example: "http://localhost:8081". If set, the adapter forwards each request
	// to this URL and maps the response (2xx + headers = allow, else deny).
	// If empty, the adapter expects an in-process Engine to be provided.
	AuthURL string `json:"auth_url,omitempty"`

	// UpstreamURL is the upstream to proxy to when the engine allows the request.
	// Used when the adapter performs the upstream handoff (in-process Engine).
	UpstreamURL string `json:"upstream_url,omitempty"`

	// RequireAuth, when true, causes unauthenticated requests to be denied (401).
	// Only applies when using in-process Engine; forward-auth uses proxy behavior.
	RequireAuth bool `json:"require_auth,omitempty"`
}
