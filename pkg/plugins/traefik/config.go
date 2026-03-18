package traefik

// Config is the configuration shape for the AuthSentinel Traefik adapter.
// Used when the adapter is configured via Traefik dynamic config or plugin config.
// All adapters call the same proxy/policy/token runtime.
type Config struct {
	// AuthURL is the AuthSentinel proxy base URL when using ForwardAuth mode.
	// Traefik sends the request to this URL; 2xx + auth response headers = allow,
	// otherwise the response is returned as deny.
	AuthURL string `json:"auth_url,omitempty"`

	// UpstreamURL is the upstream to proxy to when the engine allows (in-process mode).
	UpstreamURL string `json:"upstream_url,omitempty"`

	// AuthResponseHeaders lists header names to copy from the auth response to the
	// forwarded request when using ForwardAuth. Empty means forward X-* and Authorization.
	AuthResponseHeaders []string `json:"auth_response_headers,omitempty"`
}
