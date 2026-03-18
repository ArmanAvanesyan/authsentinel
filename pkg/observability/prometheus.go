package observability

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMetrics implements Metrics using Prometheus counters and gauges.
type PrometheusMetrics struct {
	authDecisions    *prometheus.CounterVec
	jwksCache        *prometheus.CounterVec
	sessionStoreOps  *prometheus.CounterVec
	pluginHealth     *prometheus.GaugeVec
}

// NewPrometheusMetrics registers and returns Prometheus metrics for AuthSentinel.
// The returned handler serves /metrics; expose it on your metrics port or path.
func NewPrometheusMetrics(reg *prometheus.Registry) (*PrometheusMetrics, http.Handler) {
	if reg == nil {
		reg = prometheus.NewRegistry()
	}
	authDecisions := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "authsentinel_auth_decisions_total", Help: "Total auth decisions by result and status code."},
		[]string{"result", "status_code"},
	)
	jwksCache := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "authsentinel_jwks_cache_operations_total", Help: "JWKS cache hits and misses by issuer."},
		[]string{"issuer", "result"},
	)
	sessionStoreOps := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "authsentinel_session_store_operations_total", Help: "Session store get/set operations by operation and result."},
		[]string{"operation", "result"},
	)
	pluginHealth := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: "authsentinel_plugin_health_state", Help: "Plugin health state (1=healthy, 0.5=degraded, 0=unhealthy/stopped)."},
		[]string{"plugin_id"},
	)
	reg.MustRegister(authDecisions, jwksCache, sessionStoreOps, pluginHealth)
	m := &PrometheusMetrics{
		authDecisions:   authDecisions,
		jwksCache:       jwksCache,
		sessionStoreOps: sessionStoreOps,
		pluginHealth:    pluginHealth,
	}
	handler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	return m, handler
}

func (p *PrometheusMetrics) AuthDecision(allow bool, statusCode int) {
	result := "deny"
	if allow {
		result = "allow"
	}
	p.authDecisions.WithLabelValues(result, strconv.Itoa(statusCode)).Inc()
}

func (p *PrometheusMetrics) JWKSCacheHit(issuer string) {
	p.jwksCache.WithLabelValues(issuer, "hit").Inc()
}

func (p *PrometheusMetrics) JWKSCacheMiss(issuer string) {
	p.jwksCache.WithLabelValues(issuer, "miss").Inc()
}

func (p *PrometheusMetrics) SessionStoreOp(operation string, success bool) {
	result := "error"
	if success {
		result = "ok"
	}
	p.sessionStoreOps.WithLabelValues(operation, result).Inc()
}

var healthStateValue = map[string]float64{
	"healthy":  1,
	"degraded": 0.5,
	"stopped":  0,
	"":         0,
}

func (p *PrometheusMetrics) PluginHealthState(pluginID string, state string) {
	v := healthStateValue[state]
	if v == 0 && state != "" && state != "stopped" {
		v = 0
	}
	p.pluginHealth.WithLabelValues(pluginID).Set(v)
}
