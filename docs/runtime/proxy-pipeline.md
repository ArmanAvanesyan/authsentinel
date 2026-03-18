# Proxy pipeline

The AuthSentinel proxy is **gateway-agnostic**: it takes a normalized request, resolves the principal, evaluates policy, and returns an allow/deny response with optional upstream headers. Gateways (Caddy, Traefik, KrakenD) or a standalone HTTP server call the same **pkg/proxy** engine.

## Flow overview

1. **Request in**: Gateway or HTTP server receives the request and converts it to **proxy.Request** (or the proxy’s HTTP middleware does it).
2. **Principal resolution**: **PrincipalResolver** resolves identity (e.g. via session cookie by calling the agent’s resolve endpoint).
3. **Auth requirement**: If `RequireAuth` is true and no principal is found, the engine returns **401 Unauthorized**.
4. **Policy evaluation**: **policy.Engine** is called with **policy.Input** (protocol, method, path, GraphQL/ gRPC info, principal, headers). Result is **policy.Decision** (allow/deny, status code, reason, headers, obligations).
5. **Response**: Decision is turned into **proxy.Response** (Allow, StatusCode, Body, UpstreamHeaders, SetCookies). Obligations (e.g. `set_header_X_User`) are merged into upstream headers.
6. **Out**: On allow with an upstream URL, the request is proxied to upstream with **ProxyToUpstream**; on deny, **WriteResponse** writes status and body (and any Set-Cookies).

## Key types (pkg/proxy)

- **Request**: Normalized protocol, method, path, headers, cookies, body (and optional GraphQL/gRPC fields). Built from HTTP via **RequestFromHTTP(r)**.
- **Response**: Allow, StatusCode, Body, UpstreamHeaders, SetCookies.
- **Engine**: `Handle(ctx, req) (*Response, error)`. **DefaultEngine** implements it with a Resolver, Policy engine, UpstreamURL, RequireAuth, and optional HeaderBuilder and Metrics.
- **PrincipalResolver**: `Resolve(ctx, req) (*token.Principal, error)`. The proxy uses **internal/proxy.AgentPrincipalResolver** to call the agent’s resolve endpoint with the session cookie.

## HTTP middleware (standalone proxy)

When running as **authsentinel-proxy** (cmd/proxy), the HTTP server uses:

- **proxy.Middleware(engine, upstreamURL)** — builds **Request** from **RequestFromHTTP(r)**, calls **engine.Handle**, then either **ProxyToUpstream** (allow + non-empty upstream URL) or **WriteResponse** (deny/error).

So the pipeline inside the binary is: **HTTP request → RequestFromHTTP → Engine.Handle (resolve → policy) → Response → ProxyToUpstream or WriteResponse**.

## Gateway adapters

Adapters (Caddy, Traefik, KrakenD) do **not** reimplement auth logic. They:

1. Build **proxy.Request** from the gateway request (e.g. **RequestFromHTTP(r)**).
2. Call the same **proxy.Engine** (or forward to authsentinel-proxy for forward-auth).
3. Map **Response** back to the gateway: set upstream headers, write deny status/body, set cookies.

See [Caddy](../integration/caddy.md), [Traefik](../integration/traefik.md), [KrakenD](../integration/krakend.md) and [Compatibility matrix](../integration/compatibility-matrix.md).

## Pipeline plugins (optional)

If configured, the proxy engine runs **pipeline plugins** immediately after principal resolution and **before** calling the main policy engine:

- Pipeline plugins are executed in the order from `pipeline_plugins` in the proxy config.
- If any plugin returns a non-nil `*policy.Decision`, the proxy short-circuits and uses that decision directly (the main policy engine is skipped for that request).
- If a plugin returns `nil`, execution continues to the next plugin; if all plugins return `nil`, the main policy engine runs as usual.

The built-in `pipeline:ratelimit` plugin follows the recommended “deny-short-circuit” behavior: it returns `nil` on allow (so policy still governs allow/deny) and returns a `429` decision on limit exceeded.

## References

- [pkg/proxy](https://pkg.go.dev/github.com/ArmanAvanesyan/authsentinel/pkg/proxy) — Engine, Request, Response, Middleware, RequestFromHTTP, WriteResponse, ProxyToUpstream.
- [Policy engine](policy-engine.md) — Input, Decision, WASM runtime.
- [Agent flows](agent-flows.md) — Session and resolve contract.
