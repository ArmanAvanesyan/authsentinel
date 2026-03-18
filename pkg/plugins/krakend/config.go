package krakend

// Config is the configuration shape for the AuthSentinel KrakenD adapter.
// It aligns with schemas/plugins/integration/krakend.schema.json.
// The adapter acts as an endpoint/auth middleware bridge; all adapters call
// the same proxy/policy/token runtime and share principal decision output.
type Config struct {
	// Endpoint is the KrakenD endpoint pattern this plugin protects (e.g. "/api/*").
	Endpoint string `json:"endpoint"`

	// UpstreamURL is the upstream to forward allowed requests to.
	UpstreamURL string `json:"upstream_url"`

	// RequireAuth, when true, denies unauthenticated requests with 401.
	RequireAuth bool `json:"require_auth,omitempty"`
}
