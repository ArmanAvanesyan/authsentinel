# Caddy integration

AuthSentinel can run behind Caddy so that Caddy forwards requests to the AuthSentinel proxy for authentication and policy, then either denies the request or passes it through with principal headers.

## Rule

All adapters call the **same** proxy/policy/token runtime. The Caddy adapter does not duplicate auth logic; it either uses an in-process `proxy.Engine` or forwards to authsentinel-proxy (forward-auth).

## Config shape

- **auth_url**: AuthSentinel proxy base URL when using forward-auth (e.g. `http://localhost:8081`).
- **upstream_url**: Upstream to proxy to when the adapter performs upstream handoff (in-process Engine).
- **require_auth**: When true, unauthenticated requests are denied (in-process mode).

See [schemas/plugins/integration/caddy.schema.json](../../schemas/plugins/integration/caddy.schema.json) and [configs/plugins/caddy.example.json](../../configs/plugins/caddy.example.json).

## Plugin registration entrypoint

- **In-process**: Use `pkg/plugins/caddy.NewMiddleware(engine, descriptor, upstreamURL)` and register the returned `Middleware` as an IntegrationPlugin. Call `Middleware.Handler(next)` to get an `http.Handler` and wire it into your HTTP stack.
- **Caddy directive (optional)**: Build with `-tags caddy` and add the Caddy dependency. Then use the `authsentinel` directive in your Caddyfile; the directive implements forward-auth to the given URL.

  ```caddyfile
  authsentinel http://localhost:8081
  ```

- **Forward-auth (no custom build)**: Use Caddy’s built-in `forward_auth` to point at authsentinel-proxy. Example: [configs/plugins/caddy.Caddyfile](../../configs/plugins/caddy.Caddyfile).

## Translation: gateway request → proxy engine input

- The adapter builds a `proxy.Request` from the incoming HTTP request via `proxy.RequestFromHTTP(r)` (protocol, method, path, headers, cookies, body). No Caddy-specific mapping is required; the common proxy input is used.

## Response mapping

- **Allow**: Upstream headers from the proxy response (e.g. `X-User-Id`, `X-Roles`, `Authorization`) are set on the response; then either the next handler is called or the request is proxied to `upstream_url` via `proxy.ProxyToUpstream`.
- **Deny**: The proxy response (status code, body, Set-Cookie) is written with `proxy.WriteResponse`.

## Upstream handoff

When using the in-process adapter with a non-empty `upstream_url`, the handler calls `proxy.ProxyToUpstream` after a successful engine decision and forwards the request to that URL with the upstream headers. When using forward-auth, Caddy’s `reverse_proxy` performs the upstream handoff after the forward_auth middleware allows the request.

## Example config

See [configs/plugins/caddy.Caddyfile](../../configs/plugins/caddy.Caddyfile) for a full example using `forward_auth` and an optional `authsentinel` directive.

## Compatibility

See [Compatibility matrix](compatibility-matrix.md).
