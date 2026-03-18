package observability

// Metrics records operational metrics for Prometheus or similar backends.
// All methods are safe for concurrent use.
type Metrics interface {
	// AuthDecision records a single auth decision (allow or deny).
	AuthDecision(allow bool, statusCode int)
	// JWKSCacheHit records a JWKS cache hit for the given issuer.
	JWKSCacheHit(issuer string)
	// JWKSCacheMiss records a JWKS cache miss for the given issuer.
	JWKSCacheMiss(issuer string)
	// SessionStoreOp records a session store get/set with success or failure.
	SessionStoreOp(operation string, success bool)
	// PluginHealthState sets the current health state (e.g. "healthy", "degraded") for a plugin.
	PluginHealthState(pluginID string, state string)
}

// NopMetrics discards all metrics.
type NopMetrics struct{}

func (NopMetrics) AuthDecision(allow bool, statusCode int)     {}
func (NopMetrics) JWKSCacheHit(issuer string)                 {}
func (NopMetrics) JWKSCacheMiss(issuer string)               {}
func (NopMetrics) SessionStoreOp(operation string, success bool) {}
func (NopMetrics) PluginHealthState(pluginID string, state string) {}
