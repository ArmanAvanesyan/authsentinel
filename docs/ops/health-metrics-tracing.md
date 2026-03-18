# Health, metrics, and tracing

AuthSentinel agent and proxy expose health endpoints, optional Prometheus metrics, and a guarded admin API. Structured logging is supported via `pkg/observability`.

## Health endpoints

Both binaries serve:

| Endpoint   | Purpose |
|-----------|---------|
| **GET /healthz** | Simple liveness; returns 200 and `ok`. Use for load balancer checks. |
| **GET /livez**  | Liveness (process is up). Returns 200 and `ok`. |
| **GET /readyz** | Readiness (ready to accept traffic). |

### Agent

- **/readyz**: If a session store (Redis) is configured, it pings Redis. On failure returns 503 and an error message. If no store is configured, returns 200.

### Proxy

- **/readyz**: Returns 200. Dependencies (e.g. agent) are not probed here to avoid cascading failure; use **/admin** (when enabled) to inspect plugin and policy status.

## Metrics

Both the **proxy** and the **agent** expose **GET /metrics** in Prometheus format using the default wiring in `cmd/proxy` and `cmd/agent`.

### Metric names

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `authsentinel_auth_decisions_total` | Counter | `result` (allow/deny), `status_code` | Per-request auth decisions. |
| `authsentinel_jwks_cache_operations_total` | Counter | `issuer`, `result` (hit/miss) | JWKS cache hits/misses (when wired). |
| `authsentinel_session_store_operations_total` | Counter | `operation` (get/set), `result` (ok/error) | Session store ops (when wired). |
| `authsentinel_plugin_health_state` | Gauge | `plugin_id` | Plugin health (1=healthy, 0.5=degraded, 0=stopped). |

Auth decision metrics are recorded by the proxy engine when `observability.Metrics` is set. JWKS cache hits/misses and session-store operations are recorded when the agent wires `observability.Metrics` into the JWKS source and Redis store (done in `cmd/agent`).

## Admin (guarded)

When **AdminSecret** is set in config, both binaries expose **GET /admin**. Requests must include the header:

```http
X-Admin-Secret: <your-admin-secret>
```

Otherwise the response is 403 Forbidden. If `admin_secret` is empty, the admin route is not registered.

### Agent GET /admin

Returns JSON:

- **config_summary**: Non-secret config (oidc_issuer, cookie_name, http_port, redis_url_set, app_base_url).
- **session_store**: `{ "status": "ok" | "error", "error"?: "..." }` from Redis Ping.

### Proxy GET /admin

Returns JSON:

- **config_summary**: upstream_url, proxy_path_prefix, require_auth, agent_url, cookie_name, http_port.
- **plugins**: Loaded plugins (id, kind, name, enabled, state).
- **plugin_health**: Per-plugin state and error from the registry.
- **policy_bundle**: `{ "loaded": bool, "bundle_path": string }` from the policy engine (when it implements `EngineWithStatus`).

Use admin to verify plugin discovery, policy bundle load status, and session store connectivity without logging secrets.

## Structured logging

`pkg/observability` provides:

- **Logger**: Interface with `Info`, `Warn`, `Error`, and `With(keyvals...)`. Use **StdLogger** for key=value output to a `log.Logger`.
- **Metrics**: Interface for auth decisions, JWKS cache, session store, and plugin health. **PrometheusMetrics** implements it and registers Prometheus counters/gauges.
- **Tracer**: Optional tracing; **NopTracer** is the default.

Wire a **Provider** (logger + metrics + tracer) in `cmd/agent` and `cmd/proxy` and pass the logger/metrics into components that support them.

## Tracing

Tracing is optional. The **Tracer** interface in `pkg/observability` allows starting spans with key-value attributes. Default is **NopTracer**.

### OTLP/OpenTelemetry (env-only)
The repo includes an OTLP-backed tracer that is enabled when `OTEL_EXPORTER_OTLP_ENDPOINT` is set:

- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP endpoint (for example `http://localhost:4317`)
- `OTEL_EXPORTER_OTLP_PROTOCOL`: optional, `grpc` (default) or `http/protobuf`
- `OTEL_SERVICE_NAME`: optional service name (default `authsentinel`)

Tracer initialization is best-effort: if OTLP env vars are missing (or initialization fails), tracing silently falls back to **NopTracer** and must not break auth flows.

## Deployment

- **Docker**: See `deployments/docker/` (Dockerfiles for agent and proxy, docker-compose, `.env.example`). Health checks can use `GET /livez` or `GET /readyz`.
- **Kubernetes**: Use `/livez` for liveness and `/readyz` for readiness. Restrict `/admin` and `/metrics` to internal or authenticated access (e.g. network policy or ingress rules). Helm charts can be added later if needed.

## Testing

- **Redis integration tests**: Tests in `internal/store/redis` (session, PKCE, refresh lock, revoked, replay) require a running Redis. Set **REDIS_URL** (e.g. `redis://localhost:6379/1`) to run them; otherwise they are skipped. In CI, set REDIS_URL or use testcontainers to start Redis for these tests.
