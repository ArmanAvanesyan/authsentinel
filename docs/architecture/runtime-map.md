# Runtime map

This document defines **exact responsibilities** and **allowed imports** per package. It supports Phase 0: a stable architectural baseline before adding features.

## Product model (frozen)

- **authsentinel-agent**: Login, callback, refresh, logout, session bootstrap; browser redirects; CSRF/state/nonce; cookie-backed sessions.
- **authsentinel-proxy**: Gateway-agnostic request pipeline, token/session validation, policy evaluation, upstream selection, header mutation, deny/error responses.
- **Shared**: Embedded policy engine (WASM), plugin host, config (via go-config), SDKs, gateway adapters.

## Package responsibilities and allowed imports

### cmd/

| Package   | Responsibility | Allowed imports |
|-----------|----------------|-----------------|
| **cmd/agent** | Wire agent binary: load config (go-config), bootstrap observability, session store, provider, HTTP router, health/readiness. | `internal/agent/*`, `internal/store/redis`, `pkg/cookie`, `pkg/token`, **go-config** (file, env, format). |
| **cmd/proxy** | Wire proxy binary: load config (go-config), token/policy engine, request pipeline, plugin registry, health/readiness, metrics. | `internal/proxy/*`, `pkg/observability`, `pkg/plugindiscovery`, `pkg/pluginregistry`, **go-config**. |

**Rule**: Only `cmd/` (and optionally `internal/` bootstrap) may depend on **go-config**. No other package may import go-config.

### internal/

| Package   | Responsibility | Allowed imports |
|-----------|----------------|-----------------|
| **internal/agent** | Agent app orchestration: HTTP handlers, callback flow, session bootstrap, error mapping, provider glue (IdP). | `internal/agent/config`, `internal/store/redis`, `pkg/agent`, `pkg/cookie`, `pkg/token`. |
| **internal/proxy** | Proxy app orchestration: route assembly, pipeline assembly, agent client (resolve), response shaping, deny/API mapping. | `internal/proxy/config`, `pkg/proxy`, `pkg/policy`, `pkg/observability`, `pkg/pluginregistry`. |
| **internal/store/redis** | Concrete session store, PKCE store, refresh lock, revocation; implements interfaces consumed by `pkg/session` and agent. | `pkg/session` (interfaces), Redis client. |

**Rule**: `internal/*` must not import other `internal/*` packages except as already present. No `internal` → `cmd`.

### pkg/ (reusable runtime)

| Package   | Responsibility | Allowed imports |
|-----------|----------------|-----------------|
| **pkg/agent** | Agent contract: LoginStart/LoginEnd, Session, Refresh, Logout request/response types and service interface. | None from authsentinel (stdlib only if needed). |
| **pkg/cookie** | Cookie model, codecs, signing, chunking, same-site/secure/domain, WriteOutCookie. | stdlib. |
| **pkg/token** | JWT validation, JWKS, claims, principal model, NormalizeClaims. | stdlib, HTTP/JWKS deps. |
| **pkg/policy** | Policy Engine interface, Input/Decision, WASM runtime (wazero), Rego placeholder, bundle loader. | `pkg/token` (Principal in Input). |
| **pkg/proxy** | Gateway-agnostic engine (Engine, DefaultEngine), Request/Response, RequestFromHTTP, WriteResponse, ProxyToUpstream, Middleware, PrincipalResolver. | `pkg/policy`, `pkg/token`, `pkg/cookie`, `pkg/observability`. |
| **pkg/session** | Session model, store interface, PKCE/refresh lock interfaces. | stdlib. |
| **pkg/sdk** | Server-side: HTTP middleware, PrincipalExtractor, context (WithPrincipal, PrincipalFromContext), GraphQL, gRPC interceptor, AgentClient, proxy adapter. | `pkg/token`, `pkg/cookie`, HTTP/gRPC. |
| **pkg/graphql** | GraphQL request parsing, principal injection. | stdlib. |
| **pkg/grpc** | gRPC metadata extraction, JWT validation, interceptor adapter. | `pkg/token`, gRPC. |
| **pkg/observability** | Logger, Metrics, Tracer, Provider; PrometheusMetrics, StdLogger, NopTracer. | stdlib, Prometheus. |
| **pkg/pluginapi** | Plugin interfaces (Plugin, PipelinePlugin, ProviderPlugin, IntegrationPlugin), PluginDescriptor, PluginState, Capability. | `pkg/policy`, `pkg/proxy`, `pkg/token`. |
| **pkg/pluginregistry** | Registration, enable/disable, resolution by capability, dependency graph. | `pkg/pluginapi`. |
| **pkg/plugindiscovery** | Built-in and manifest-based discovery (DiscoverFromDir), manifest parsing. | `pkg/pluginapi`, `pkg/pluginregistry`. |
| **pkg/pluginconfig** | Plugin config envelope, schema mapping. | stdlib. |
| **pkg/pluginhost** | Host services for plugins: logger, metrics, cache, secrets, Redis abstraction. | observability, store interfaces. |
| **pkg/plugins/caddy** | Caddy adapter: config, Handler, optional directive; forward-auth. | `pkg/proxy`, `pkg/pluginapi`. |
| **pkg/plugins/traefik** | Traefik adapter: config, Middleware.Handler, header/deny mapping. | `pkg/proxy`, `pkg/pluginapi`. |
| **pkg/plugins/krakend** | KrakenD adapter: config, Handler, endpoint/auth bridge. | `pkg/proxy`, `pkg/pluginapi`. |
| **pkg/testing** | Test fixtures and helpers. | Other `pkg/*` as needed. |

**Rule**: `pkg/*` must not import `cmd/*`, `internal/*`, or go-config. Cross-pkg imports only as in the table.

### Config loading

- **Who**: `cmd/agent` and `cmd/proxy` only.
- **How**: go-config (file + env, JSON format). Config **types** live in `internal/agent/config` and `internal/proxy/config`.
- **No** `pkg/config`: config loading is not a reusable library; it is application bootstrap.

## Summary

| Layer    | May use go-config? | May import internal? |
|----------|--------------------|------------------------|
| cmd/*    | Yes                | Yes                   |
| internal/* | No               | Yes (own tree only)   |
| pkg/*    | No                 | No                    |

See [package-boundaries.md](package-boundaries.md) for a short list of packages and [implementation-plan.md](implementation-plan.md) for the full phased plan.
